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
	ChatCompletionInformationChat(ctx context.Context, model string, messages []ai.Message) (*ai.InformationChatResponse, error)
	ChatCompletionComparison(ctx context.Context, model string, messages []ai.Message) (*ai.ComparisonContent, error)
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
func (u *DisplayUsecase) GenerateDisplay(ctx context.Context, input GenerateDisplayInput) (*GenerateDisplayOutput, error) {
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

	return &GenerateDisplayOutput{
		Title:         content.Title,
		Body:          content.Body,
		HasPreference: len(preference.Items()) > 0,
	}, nil
}

// GenerateComparisonInput は、指定したPreferenceキーについて、AIが決めた
// 対照的な2つの値(傾向)による表示A/Bを生成するための入力。
//
// どの値を比較するかを受信者に自由入力させるのではなく、AIがそのキーにとって
// 意味のある対照的な2値を自分で決める(例: reading_levelなら"easy"と"detailed")。
// 1回のAI呼び出しでA/B両方を生成するため、無駄な追加呼び出しは発生しない。
type GenerateComparisonInput struct {
	InformationID uuid.UUID
	RecipientID   uuid.UUID
	Key           string
}

// ComparisonPattern は、比較対象のPreferenceキーについてAIが決めた
// 1つの値(傾向)と、その値を採用した場合の表示を表す。
type ComparisonPattern struct {
	Value   string         `json:"value"`
	Display DisplayContent `json:"display"`
}

// GenerateComparisonOutput はGenerateComparisonの出力。
type GenerateComparisonOutput struct {
	Key      string            `json:"key"`
	PatternA ComparisonPattern `json:"pattern_a"`
	PatternB ComparisonPattern `json:"pattern_b"`
}

// GenerateComparison は、Keyで指定したPreferenceキーについて、AIが決めた
// 対照的な2パターン(A/B)の表示を生成する。
// この時点ではPreferenceの永続化は行わない(採用はApplyPreferenceで行う)。
func (u *DisplayUsecase) GenerateComparison(ctx context.Context, input GenerateComparisonInput) (*GenerateComparisonOutput, error) {
	sot, err := u.loadSourceOfTruth(ctx, input.InformationID, input.RecipientID)
	if err != nil {
		return nil, err
	}

	preference, err := u.loadOrCreatePreference(ctx, input.RecipientID)
	if err != nil {
		return nil, fmt.Errorf("load preference: %w", err)
	}

	prompt, err := ai.BuildComparisonSystemPrompt(*sot, preference.Items(), input.Key)
	if err != nil {
		return nil, fmt.Errorf("build comparison prompt: %w", err)
	}

	messages := ai.BuildMessages(prompt, "この情報を比較用に2パターン生成してください。")
	content, err := u.aiClient.ChatCompletionComparison(ctx, u.aiModel, messages)
	if err != nil {
		return nil, fmt.Errorf("chat completion: %w", err)
	}

	return &GenerateComparisonOutput{
		Key: input.Key,
		PatternA: ComparisonPattern{
			Value:   content.PatternA.Value,
			Display: DisplayContent{Title: content.PatternA.Title, Body: content.PatternA.Body},
		},
		PatternB: ComparisonPattern{
			Value:   content.PatternB.Value,
			Display: DisplayContent{Title: content.PatternB.Title, Body: content.PatternB.Body},
		},
	}, nil
}

// ApplyPreferenceInput はApplyPreferenceへの入力。
type ApplyPreferenceInput struct {
	InformationID uuid.UUID
	RecipientID   uuid.UUID
	Key           string
	Value         string
}

// ApplyPreference は、GenerateComparisonで確認した候補の値を実際のPreferenceとして
// 永続化する。AI呼び出しは行わない(表示内容はKey/Valueから一意に決まるものではなく
// 次回のGenerateDisplayで再生成されるため、ここでは値の保存のみを行う)。
func (u *DisplayUsecase) ApplyPreference(ctx context.Context, input ApplyPreferenceInput) error {
	if _, err := u.loadSourceOfTruth(ctx, input.InformationID, input.RecipientID); err != nil {
		return err
	}

	preference, err := u.loadOrCreatePreference(ctx, input.RecipientID)
	if err != nil {
		return fmt.Errorf("load preference: %w", err)
	}

	if err := preference.Set(input.Key, input.Value); err != nil {
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
	Messages []Message `json:"messages"`
	Answer   string    `json:"answer"`
}

// Chat は、受信者からの資料の情報についての質問に1ターン回答する。
//
// これは表示の見せ方を調整する機能ではなく、確定済みの情報(Source of Truth)に
// ついてのQ&Aである。そのためPreferenceの読み取り(回答の言葉遣いを合わせる
// 参考情報として)は行うが、Chatを通じてPreferenceが更新されることはない。
// Preferenceの更新はA/B比較の選択結果(SelectComparison)からのみ行う。
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

	resp, err := u.aiClient.ChatCompletionInformationChat(ctx, u.aiModel, messages)
	if err != nil {
		return nil, fmt.Errorf("chat completion: %w", err)
	}

	raw, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("marshal information chat response: %w", err)
	}
	messages = append(messages, ai.Message{Role: ai.RoleAssistant, Content: string(raw)})

	return &ChatOutput{
		Messages: fromAIMessages(messages),
		Answer:   resp.Answer,
	}, nil
}

func (u *DisplayUsecase) nextChatMessages(history []ai.Message, userInput string, sot ai.SourceOfTruth, preference map[string]string) ([]ai.Message, error) {
	if len(history) > 0 {
		return append(history, ai.Message{Role: ai.RoleUser, Content: userInput}), nil
	}

	prompt, err := ai.BuildInformationChatSystemPrompt(sot, preference)
	if err != nil {
		return nil, fmt.Errorf("build information chat prompt: %w", err)
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
