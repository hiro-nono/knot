package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"knot-api/internal/domain"
	"knot-api/internal/infrastructure/ai"
	"knot-api/internal/repository"

	"uuid"
)

// displayChatClient は表示生成・Chat用の構造化出力送信を抽象化する。
// テストではAIへの実際のHTTP通信を行わないfakeに差し替える。
type displayChatClient interface {
	ChatCompletionDisplay(ctx context.Context, model string, messages []ai.Message) (*ai.DisplayContent, error)
	ChatCompletionPreferenceChat(ctx context.Context, model string, messages []ai.Message) (*ai.PreferenceChatResponse, error)
}

// DisplayUsecase は確定済みのSource of Truthを、受信者(認証済みのUser)の
// Preferenceに応じて最適化した表示へ変換するアプリケーションフローを管理する。
//
// A/B比較とChatの両方を通じて受信者のPreferenceを形成し、
// そのPreferenceを次回以降の表示最適化に利用する。
// 表示生成前には、access_type・recipientsをもとにその受信者がInformationを
// 閲覧してよいかどうかをloadSourceOfTruthで判定する(アクセス制御の責務)。
type DisplayUsecase struct {
	aiClient        displayChatClient
	aiModel         string
	informationRepo repository.InformationRepository
	sourceRepo      repository.SourceRepository
	optionRepo      repository.OptionRepository
	preferenceRepo  repository.PreferenceRepository
	recipientRepo   repository.RecipientRepository
}

// NewDisplayUsecase はDisplayUsecaseを生成する。
func NewDisplayUsecase(
	aiClient displayChatClient,
	aiModel string,
	informationRepo repository.InformationRepository,
	sourceRepo repository.SourceRepository,
	optionRepo repository.OptionRepository,
	preferenceRepo repository.PreferenceRepository,
	recipientRepo repository.RecipientRepository,
) *DisplayUsecase {
	return &DisplayUsecase{
		aiClient:        aiClient,
		aiModel:         aiModel,
		informationRepo: informationRepo,
		sourceRepo:      sourceRepo,
		optionRepo:      optionRepo,
		preferenceRepo:  preferenceRepo,
		recipientRepo:   recipientRepo,
	}
}

// GenerateDisplayInput はGenerateDisplayへの入力。
// RecipientIDは認証済みの受信者(User)のIDであり、閲覧可否の判定に使われる。
type GenerateDisplayInput struct {
	InformationID uuid.UUID
	RecipientID   uuid.UUID
}

// GenerateDisplay は指定したInformationを、受信者のPreferenceに応じて
// 最適化した表示内容として生成する。
// 受信者がそのInformationを閲覧する権限を持たない場合はdomain.ErrForbiddenを返す。
func (u *DisplayUsecase) GenerateDisplay(ctx context.Context, input GenerateDisplayInput) (*DisplayContent, error) {
	sot, err := u.loadSourceOfTruth(ctx, input.InformationID, input.RecipientID)
	if err != nil {
		return nil, err
	}

	preference, err := u.loadOrCreatePreference(ctx, input.RecipientID)
	if err != nil {
		return nil, fmt.Errorf("load preference: %w", err)
	}

	content, err := u.generateDisplay(ctx, *sot, preference.Items())
	if err != nil {
		return nil, err
	}

	return content, nil
}

// PreferenceComparison はA/B比較で、どのPreferenceキーの、どの値同士を
// 比較しているかを表す。controller/クライアントとの間でもそのまま
// やり取りされ、選択結果を送る際にどの比較だったかを特定するために使う。
type PreferenceComparison struct {
	Key    string `json:"key"`
	ValueA string `json:"value_a"`
	ValueB string `json:"value_b"`
}

// PrepareComparisonInput はA/B比較用の表示を2パターン生成するための入力。
type PrepareComparisonInput struct {
	InformationID uuid.UUID
	RecipientID   uuid.UUID
	Comparison    PreferenceComparison
}

// PrepareComparisonOutput はPrepareComparisonの出力。
type PrepareComparisonOutput struct {
	DisplayA   DisplayContent       `json:"display_a"`
	DisplayB   DisplayContent       `json:"display_b"`
	Comparison PreferenceComparison `json:"comparison"`
}

// PrepareComparison は、Comparisonで指定したPreferenceキーの値をValueA/ValueBに
// それぞれ変えた場合の表示を2パターン生成する。
// この時点ではPreferenceの永続化は行わない(選択結果はSelectComparisonで反映する)。
func (u *DisplayUsecase) PrepareComparison(ctx context.Context, input PrepareComparisonInput) (*PrepareComparisonOutput, error) {
	sot, err := u.loadSourceOfTruth(ctx, input.InformationID, input.RecipientID)
	if err != nil {
		return nil, err
	}

	preference, err := u.loadOrCreatePreference(ctx, input.RecipientID)
	if err != nil {
		return nil, fmt.Errorf("load preference: %w", err)
	}

	displayA, err := u.generateDisplay(ctx, *sot, withOverride(preference.Items(), input.Comparison.Key, input.Comparison.ValueA))
	if err != nil {
		return nil, fmt.Errorf("generate display a: %w", err)
	}

	displayB, err := u.generateDisplay(ctx, *sot, withOverride(preference.Items(), input.Comparison.Key, input.Comparison.ValueB))
	if err != nil {
		return nil, fmt.Errorf("generate display b: %w", err)
	}

	return &PrepareComparisonOutput{
		DisplayA:   *displayA,
		DisplayB:   *displayB,
		Comparison: input.Comparison,
	}, nil
}

// SelectComparisonInput はSelectComparisonへの入力。
type SelectComparisonInput struct {
	InformationID uuid.UUID
	RecipientID   uuid.UUID
	Comparison    PreferenceComparison
	Selected      string // "a" または "b"
}

// SelectComparison は受信者がA/Bのどちらを選んだかをもとに、
// 比較対象となっていたPreferenceのキーを、選ばれた値で更新する。
func (u *DisplayUsecase) SelectComparison(ctx context.Context, input SelectComparisonInput) error {
	if _, err := u.loadSourceOfTruth(ctx, input.InformationID, input.RecipientID); err != nil {
		return err
	}

	var value string
	switch input.Selected {
	case "a":
		value = input.Comparison.ValueA
	case "b":
		value = input.Comparison.ValueB
	default:
		return fmt.Errorf(`selected must be "a" or "b", got %q`, input.Selected)
	}

	preference, err := u.loadOrCreatePreference(ctx, input.RecipientID)
	if err != nil {
		return fmt.Errorf("load preference: %w", err)
	}

	if err := preference.Set(input.Comparison.Key, value); err != nil {
		return fmt.Errorf("set preference: %w", err)
	}

	if err := u.preferenceRepo.Save(ctx, preference); err != nil {
		return fmt.Errorf("save preference: %w", err)
	}

	return nil
}

// ChatInput はChatへの入力。
type ChatInput struct {
	InformationID uuid.UUID
	RecipientID   uuid.UUID
	Messages      []Message
	UserInput     string
}

// ChatOutput はChatの出力。
type ChatOutput struct {
	Messages             []Message      `json:"messages"`
	Display              DisplayContent `json:"display"`
	NeedsConfirmation    bool           `json:"needs_confirmation"`
	ConfirmationQuestion *string        `json:"confirmation_question,omitempty"`
	PreferenceUpdated    bool           `json:"preference_updated"`
}

// Chat は受信者からの表示調整の要望を1ターン処理する。
//
// AIが「今後も適用するPreferenceだ」と判断した場合のみ、このUseCase層で
// Preferenceを更新する。一時的な指示か継続的なPreferenceかをAIが判断できない
// 場合は確認の質問を返し、Preferenceは更新しない(対話の続きは呼び出し元が
// 次回のHTTPリクエストで行う。AI内部でループしない)。
func (u *DisplayUsecase) Chat(ctx context.Context, input ChatInput) (*ChatOutput, error) {
	sot, err := u.loadSourceOfTruth(ctx, input.InformationID, input.RecipientID)
	if err != nil {
		return nil, err
	}

	preference, err := u.loadOrCreatePreference(ctx, input.RecipientID)
	if err != nil {
		return nil, fmt.Errorf("load preference: %w", err)
	}

	messages, err := u.nextChatMessages(toAIMessages(input.Messages), input.UserInput, *sot, preference.Items())
	if err != nil {
		return nil, err
	}

	resp, err := u.aiClient.ChatCompletionPreferenceChat(ctx, u.aiModel, messages)
	if err != nil {
		return nil, fmt.Errorf("chat completion: %w", err)
	}

	raw, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("marshal preference chat response: %w", err)
	}
	messages = append(messages, ai.Message{Role: ai.RoleAssistant, Content: string(raw)})

	preferenceUpdated := false
	if resp.IsPersistent != nil && *resp.IsPersistent && resp.PreferenceKey != nil && resp.PreferenceValue != nil {
		if err := preference.Set(*resp.PreferenceKey, *resp.PreferenceValue); err != nil {
			return nil, fmt.Errorf("set preference: %w", err)
		}
		if err := u.preferenceRepo.Save(ctx, preference); err != nil {
			return nil, fmt.Errorf("save preference: %w", err)
		}
		preferenceUpdated = true
	}

	return &ChatOutput{
		Messages:             fromAIMessages(messages),
		Display:              DisplayContent{Title: resp.Display.Title, Body: resp.Display.Body},
		NeedsConfirmation:    resp.IsPersistent == nil,
		ConfirmationQuestion: resp.ConfirmationQuestion,
		PreferenceUpdated:    preferenceUpdated,
	}, nil
}

func (u *DisplayUsecase) nextChatMessages(history []ai.Message, userInput string, sot ai.SourceOfTruth, preference map[string]string) ([]ai.Message, error) {
	if len(history) > 0 {
		return append(history, ai.Message{Role: ai.RoleUser, Content: userInput}), nil
	}

	prompt, err := ai.BuildPreferenceChatSystemPrompt(sot, preference)
	if err != nil {
		return nil, fmt.Errorf("build preference chat prompt: %w", err)
	}
	if knownKeysPrompt := ai.BuildKnownPreferenceKeysPrompt(domain.PredefinedPreferenceKeys); knownKeysPrompt != "" {
		prompt += "\n\n" + knownKeysPrompt
	}

	return ai.BuildMessages(prompt, userInput), nil
}

func (u *DisplayUsecase) generateDisplay(ctx context.Context, sot ai.SourceOfTruth, preference map[string]string) (*DisplayContent, error) {
	prompt, err := ai.BuildDisplaySystemPrompt(sot, preference)
	if err != nil {
		return nil, fmt.Errorf("build display prompt: %w", err)
	}

	messages := ai.BuildMessages(prompt, "この情報を表示用に最適化してください。")

	content, err := u.aiClient.ChatCompletionDisplay(ctx, u.aiModel, messages)
	if err != nil {
		return nil, fmt.Errorf("chat completion: %w", err)
	}

	return &DisplayContent{Title: content.Title, Body: content.Body}, nil
}

// ListSources は指定したInformationに属するSourceを、Optionを含めて一覧取得する。
// 受信者がResponseを送信する際、対象のsource_id・interaction_type・option_idを
// 知るために使う。表示生成(GenerateDisplay)とは異なり、返答(Response)と同じ
// アクセス制御(canRespondToInformation)を適用するため、PUBLIC+ANONYMOUSな
// InformationについてはviewerUserIDがnil(未認証)でも一覧取得を許可する。
func (u *DisplayUsecase) ListSources(ctx context.Context, informationID uuid.UUID, viewerUserID *uuid.UUID) ([]SourceView, error) {
	information, err := u.informationRepo.Get(ctx, informationID)
	if err != nil {
		return nil, fmt.Errorf("get information: %w", err)
	}

	allowed, err := canRespondToInformation(ctx, u.recipientRepo, information, viewerUserID)
	if err != nil {
		return nil, fmt.Errorf("check information access: %w", err)
	}
	if !allowed {
		return nil, domain.ErrForbidden
	}

	sources, err := u.sourceRepo.List(ctx, informationID)
	if err != nil {
		return nil, fmt.Errorf("list sources: %w", err)
	}

	views := make([]SourceView, 0, len(sources))
	for _, s := range sources {
		options, err := u.optionRepo.List(ctx, s.ID())
		if err != nil {
			return nil, fmt.Errorf("list options: %w", err)
		}
		views = append(views, newSourceView(s, options))
	}

	return views, nil
}

// loadSourceOfTruth はinformationIDのInformationを取得し、viewerUserIDが
// それを閲覧する権限を持つかどうかをアクセス制御の責務としてここで判定する。
// 権限が無い場合はdomain.ErrForbiddenを返す。
func (u *DisplayUsecase) loadSourceOfTruth(ctx context.Context, informationID, viewerUserID uuid.UUID) (*ai.SourceOfTruth, error) {
	information, err := u.informationRepo.Get(ctx, informationID)
	if err != nil {
		return nil, fmt.Errorf("get information: %w", err)
	}

	allowed, err := canAccessInformation(ctx, u.recipientRepo, information, viewerUserID)
	if err != nil {
		return nil, fmt.Errorf("check information access: %w", err)
	}
	if !allowed {
		return nil, domain.ErrForbidden
	}

	sources, err := u.sourceRepo.List(ctx, informationID)
	if err != nil {
		return nil, fmt.Errorf("list sources: %w", err)
	}

	sot := &ai.SourceOfTruth{Title: information.Title()}
	for _, s := range sources {
		options, err := u.optionRepo.List(ctx, s.ID())
		if err != nil {
			return nil, fmt.Errorf("list options: %w", err)
		}

		var interactionType *string
		if s.InteractionType() != nil {
			v := string(*s.InteractionType())
			interactionType = &v
		}

		optionValues := make([]string, 0, len(options))
		for _, o := range options {
			optionValues = append(optionValues, o.Value())
		}

		sot.Sources = append(sot.Sources, ai.SourceOfTruthSource{
			Type:            string(s.Type()),
			Key:             s.Key(),
			Value:           s.Value(),
			InteractionType: interactionType,
			Options:         optionValues,
		})
	}

	return sot, nil
}

func (u *DisplayUsecase) loadOrCreatePreference(ctx context.Context, recipientID uuid.UUID) (*domain.Preference, error) {
	preference, err := u.preferenceRepo.Get(ctx, recipientID)
	if err == nil {
		return preference, nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return nil, fmt.Errorf("get preference: %w", err)
	}

	return domain.NewPreference(recipientID)
}

func withOverride(items map[string]string, key, value string) map[string]string {
	items[key] = value
	return items
}
