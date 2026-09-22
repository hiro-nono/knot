package usecase

import (
	"context"
	"encoding/json"
	"fmt"

	"knot-api/internal/domain"
	"knot-api/internal/infrastructure/ai"
	"knot-api/internal/repository"

	"uuid"
)

// chatClient は構造化出力のチャット送信を抽象化する。
// テストではAIへの実際のHTTP通信を行わないfakeに差し替える。
type chatClient interface {
	ChatCompletionStructured(ctx context.Context, model string, messages []ai.Message) (*ai.StructuredInformation, error)
}

// InformationUsecase は発信者とAIの対話を通じてInformationを構造化し、
// 発信者がすべての項目を確定させた場合にのみ永続化するアプリケーションフローを管理する。
//
// AIが確定と判断するまでUseCase内部でループすることはしない。
// 1回の呼び出しは対話の1ターンのみを進め、続きは次回のHTTPリクエストで行う。
type InformationUsecase struct {
	aiClient           chatClient
	aiModel            string
	informationRepo    repository.InformationRepository
	sourceRepo         repository.SourceRepository
	optionRepo         repository.OptionRepository
	recipientRepo      repository.RecipientRepository
	transactionManager repository.TransactionManager
}

// NewInformationUsecase はInformationUsecaseを生成する。
func NewInformationUsecase(
	aiClient chatClient,
	aiModel string,
	informationRepo repository.InformationRepository,
	sourceRepo repository.SourceRepository,
	optionRepo repository.OptionRepository,
	recipientRepo repository.RecipientRepository,
	transactionManager repository.TransactionManager,
) *InformationUsecase {
	return &InformationUsecase{
		aiClient:           aiClient,
		aiModel:            aiModel,
		informationRepo:    informationRepo,
		sourceRepo:         sourceRepo,
		optionRepo:         optionRepo,
		recipientRepo:      recipientRepo,
		transactionManager: transactionManager,
	}
}

// ProcessInformationInput はProcessInformationへの入力。
// Messagesが空の場合は新しい対話として開始する。
//
// AccountID・CreatedByUserIDは認証済みの発信者から解決され、AccessTypeは
// 発信者が選択したアクセス方式(public/restricted)を表す。ResponsePolicyは
// AccessTypeがpublicの場合に、匿名回答を許すか(anonymous)ログインを必須と
// するか(authenticated)を表す(restrictedの場合は無視される)。すべての対話
// ターンで送られてくる想定だが、実際に使われるのはInformationが確定・永続化
// されるターンのみ。
type ProcessInformationInput struct {
	AccountID       uuid.UUID
	CreatedByUserID uuid.UUID
	AccessType      string
	ResponsePolicy  string
	Messages        []Message
	UserInput       string
}

// ProcessInformationOutput はProcessInformationの出力。
// Confirmedがtrueの場合のみInformationIDが設定される。
type ProcessInformationOutput struct {
	Messages      []Message              `json:"messages"`
	Structured    *StructuredInformation `json:"structured"`
	Confirmed     bool                   `json:"confirmed"`
	InformationID string                 `json:"information_id,omitempty"`
}

// ProcessInformation は発信者との対話を1ターン進める。
//
// AIに問い合わせて構造化情報を取得し、すべてのSourceがconfirmedであれば
// トランザクション内でInformation/Source/Optionを永続化する。
// 確定していない場合は永続化を行わず、AIからの質問を含む状態を返す
// (対話の続きは呼び出し元が次回のHTTPリクエストで行う)。
func (u *InformationUsecase) ProcessInformation(ctx context.Context, input ProcessInformationInput) (*ProcessInformationOutput, error) {
	messages := nextMessages(toAIMessages(input.Messages), input.UserInput)

	structured, err := u.aiClient.ChatCompletionStructured(ctx, u.aiModel, messages)
	if err != nil {
		return nil, fmt.Errorf("chat completion: %w", err)
	}

	messages, err = appendAssistantTurn(messages, structured)
	if err != nil {
		return nil, fmt.Errorf("append assistant turn: %w", err)
	}

	if !allSourcesConfirmed(structured) {
		return &ProcessInformationOutput{
			Messages:   fromAIMessages(messages),
			Structured: fromAIStructured(structured),
			Confirmed:  false,
		}, nil
	}

	informationID, err := u.persist(ctx, input, structured)
	if err != nil {
		return nil, fmt.Errorf("persist information: %w", err)
	}

	return &ProcessInformationOutput{
		Messages:      fromAIMessages(messages),
		Structured:    fromAIStructured(structured),
		Confirmed:     true,
		InformationID: informationID.String(),
	}, nil
}

// buildInitialSystemPrompt は対話の最初のターンに使うシステムプロンプトを組み立てる。
// domain.PredefinedSourceKeysに定義された既知のkeyをtype別に含め、
// AIが新しいkeyを作る前に再利用を検討できるようにする。
func buildInitialSystemPrompt() string {
	prompt := ai.BuildSystemPrompt()
	if knownKeysPrompt := ai.BuildKnownKeysPrompt(domain.PredefinedSourceKeys); knownKeysPrompt != "" {
		prompt += "\n\n" + knownKeysPrompt
	}
	return prompt
}

func nextMessages(history []ai.Message, userInput string) []ai.Message {
	if len(history) == 0 {
		return ai.BuildMessages(buildInitialSystemPrompt(), userInput)
	}
	return append(history, ai.Message{Role: ai.RoleUser, Content: userInput})
}

func appendAssistantTurn(messages []ai.Message, structured *ai.StructuredInformation) ([]ai.Message, error) {
	raw, err := json.Marshal(structured)
	if err != nil {
		return nil, fmt.Errorf("marshal structured response: %w", err)
	}
	return append(messages, ai.Message{Role: ai.RoleAssistant, Content: string(raw)}), nil
}

func allSourcesConfirmed(structured *ai.StructuredInformation) bool {
	if len(structured.Sources) == 0 {
		return false
	}
	for _, s := range structured.Sources {
		if s.Status != string(domain.SourceStatusConfirmed) {
			return false
		}
	}
	return true
}

// persist はAIが確定させた構造化情報をdomain Entityへ変換し、
// 1つのトランザクション内でInformation/Source/Optionを永続化する。
func (u *InformationUsecase) persist(ctx context.Context, input ProcessInformationInput, structured *ai.StructuredInformation) (uuid.UUID, error) {
	information, err := buildInformation(input, structured)
	if err != nil {
		return uuid.Nil(), fmt.Errorf("build information: %w", err)
	}

	err = u.transactionManager.WithinTransaction(ctx, func(ctx context.Context) error {
		return u.saveInformation(ctx, information)
	})
	if err != nil {
		return uuid.Nil(), err
	}

	return information.ID(), nil
}

func buildInformation(input ProcessInformationInput, structured *ai.StructuredInformation) (*domain.Information, error) {
	information, err := domain.NewInformation(input.AccountID, input.CreatedByUserID, structured.Title, domain.InformationAccessType(input.AccessType), domain.InformationResponsePolicy(input.ResponsePolicy))
	if err != nil {
		return nil, err
	}

	for _, s := range structured.Sources {
		source, err := buildSource(information.ID(), s)
		if err != nil {
			return nil, err
		}
		if err := information.AddSource(*source); err != nil {
			return nil, err
		}
	}

	return information, nil
}

func buildSource(informationID uuid.UUID, s ai.StructuredSource) (*domain.Source, error) {
	var interactionType *domain.SourceInteractionType
	if s.InteractionType != nil {
		it := domain.SourceInteractionType(*s.InteractionType)
		interactionType = &it
	}

	source, err := domain.NewSource(
		informationID,
		domain.SourceType(s.Type),
		s.Key,
		s.Value,
		domain.SourceStatus(s.Status),
		interactionType,
	)
	if err != nil {
		return nil, err
	}

	for _, o := range s.Options {
		option, err := domain.NewOption(source.ID(), o.Value, o.SortOrder)
		if err != nil {
			return nil, err
		}
		if err := source.AddOption(*option); err != nil {
			return nil, err
		}
	}

	return source, nil
}

func (u *InformationUsecase) saveInformation(ctx context.Context, information *domain.Information) error {
	if err := u.informationRepo.Create(ctx, information); err != nil {
		return fmt.Errorf("create information: %w", err)
	}

	for _, source := range information.Sources() {
		if err := u.sourceRepo.Create(ctx, &source); err != nil {
			return fmt.Errorf("create source: %w", err)
		}

		for _, option := range source.Options() {
			if err := u.optionRepo.Create(ctx, &option); err != nil {
				return fmt.Errorf("create option: %w", err)
			}
		}
	}

	return nil
}

// ListMine は自分(呼び出し元のAccount)が作成したInformationを、
// 作成日時の新しい順に一覧取得する。
func (u *InformationUsecase) ListMine(ctx context.Context, accountID uuid.UUID) ([]*InformationSummaryView, error) {
	informations, err := u.informationRepo.ListByAccountID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("list informations: %w", err)
	}

	views := make([]*InformationSummaryView, 0, len(informations))
	for _, information := range informations {
		views = append(views, newInformationSummaryView(information))
	}

	return views, nil
}

// RecipientInput はAddRecipient/RemoveRecipientへの入力。
// ActingAccountIDはInformationを所有するAccountと一致する必要がある
// (Informationの共有設定を変更できるのはそのAccountのみ)。
type RecipientInput struct {
	ActingAccountID uuid.UUID
	InformationID   uuid.UUID
	UserID          uuid.UUID
}

// AddRecipient はrestrictedなInformationを閲覧できるUserを1件追加する。
func (u *InformationUsecase) AddRecipient(ctx context.Context, input RecipientInput) error {
	if _, err := u.requireOwningAccount(ctx, input.ActingAccountID, input.InformationID); err != nil {
		return err
	}

	recipient, err := domain.NewRecipient(input.InformationID, input.UserID)
	if err != nil {
		return fmt.Errorf("build recipient: %w", err)
	}

	if err := u.recipientRepo.Create(ctx, recipient); err != nil {
		return fmt.Errorf("create recipient: %w", err)
	}

	return nil
}

// RemoveRecipient はInformationからUserの閲覧許可を取り消す。
func (u *InformationUsecase) RemoveRecipient(ctx context.Context, input RecipientInput) error {
	if _, err := u.requireOwningAccount(ctx, input.ActingAccountID, input.InformationID); err != nil {
		return err
	}

	if err := u.recipientRepo.Delete(ctx, input.InformationID, input.UserID); err != nil {
		return fmt.Errorf("delete recipient: %w", err)
	}

	return nil
}

// ListRecipients はInformationに登録されているRecipient(閲覧可能なUserのID)を一覧取得する。
func (u *InformationUsecase) ListRecipients(ctx context.Context, actingAccountID, informationID uuid.UUID) ([]string, error) {
	if _, err := u.requireOwningAccount(ctx, actingAccountID, informationID); err != nil {
		return nil, err
	}

	recipients, err := u.recipientRepo.ListByInformationID(ctx, informationID)
	if err != nil {
		return nil, fmt.Errorf("list recipients: %w", err)
	}

	userIDs := make([]string, 0, len(recipients))
	for _, r := range recipients {
		userIDs = append(userIDs, r.UserID().String())
	}

	return userIDs, nil
}

// requireOwningAccount はinformationIDのInformationを取得し、
// actingAccountIDがそのInformationを所有するAccountであることを検証する。
func (u *InformationUsecase) requireOwningAccount(ctx context.Context, actingAccountID, informationID uuid.UUID) (*domain.Information, error) {
	information, err := u.informationRepo.Get(ctx, informationID)
	if err != nil {
		return nil, fmt.Errorf("get information: %w", err)
	}
	if information.AccountID() != actingAccountID {
		return nil, domain.ErrForbidden
	}
	return information, nil
}
