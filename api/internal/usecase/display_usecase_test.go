package usecase

import (
	"context"
	"errors"
	"testing"

	"knot-api/internal/domain"
	"knot-api/internal/infrastructure/ai"

	"uuid"
)

type fakeDisplayInformationRepository struct {
	information *domain.Information
}

func (r *fakeDisplayInformationRepository) Create(ctx context.Context, information *domain.Information) error {
	return nil
}

func (r *fakeDisplayInformationRepository) Get(ctx context.Context, id uuid.UUID) (*domain.Information, error) {
	if r.information == nil {
		return nil, domain.ErrNotFound
	}
	return r.information, nil
}

func (r *fakeDisplayInformationRepository) ListByAccountID(ctx context.Context, accountID uuid.UUID) ([]*domain.Information, error) {
	if r.information == nil || r.information.AccountID() != accountID {
		return nil, nil
	}
	return []*domain.Information{r.information}, nil
}

type fakeDisplaySourceRepository struct {
	sources []*domain.Source
}

func (r *fakeDisplaySourceRepository) Create(ctx context.Context, source *domain.Source) error {
	return nil
}

func (r *fakeDisplaySourceRepository) List(ctx context.Context, informationID uuid.UUID) ([]*domain.Source, error) {
	return r.sources, nil
}

type fakeDisplayOptionRepository struct {
	optionsBySource map[uuid.UUID][]*domain.Option
}

func (r *fakeDisplayOptionRepository) Create(ctx context.Context, option *domain.Option) error {
	return nil
}

func (r *fakeDisplayOptionRepository) List(ctx context.Context, sourceID uuid.UUID) ([]*domain.Option, error) {
	return r.optionsBySource[sourceID], nil
}

type fakePreferenceRepository struct {
	preference *domain.Preference
	saved      []*domain.Preference
}

func (r *fakePreferenceRepository) Get(ctx context.Context, recipientID uuid.UUID) (*domain.Preference, error) {
	if r.preference == nil {
		return nil, domain.ErrNotFound
	}
	return r.preference, nil
}

func (r *fakePreferenceRepository) Save(ctx context.Context, preference *domain.Preference) error {
	r.saved = append(r.saved, preference)
	r.preference = preference
	return nil
}

type fakeDisplayRecipientRepository struct {
	allowed map[uuid.UUID]bool
}

func newFakeDisplayRecipientRepository() *fakeDisplayRecipientRepository {
	return &fakeDisplayRecipientRepository{allowed: map[uuid.UUID]bool{}}
}

func (r *fakeDisplayRecipientRepository) Create(ctx context.Context, recipient *domain.Recipient) error {
	r.allowed[recipient.UserID()] = true
	return nil
}

func (r *fakeDisplayRecipientRepository) Exists(ctx context.Context, informationID, userID uuid.UUID) (bool, error) {
	return r.allowed[userID], nil
}

func (r *fakeDisplayRecipientRepository) ListByInformationID(ctx context.Context, informationID uuid.UUID) ([]*domain.Recipient, error) {
	return nil, nil
}

func (r *fakeDisplayRecipientRepository) Delete(ctx context.Context, informationID, userID uuid.UUID) error {
	delete(r.allowed, userID)
	return nil
}

type fakeDisplayChatClient struct {
	displayResponse    *ai.DisplayContent
	chatResponse       *ai.InformationChatResponse
	comparisonResponse *ai.ComparisonContent
	lastMessages       []ai.Message
}

func (f *fakeDisplayChatClient) ChatCompletionDisplay(ctx context.Context, model string, messages []ai.Message) (*ai.DisplayContent, error) {
	f.lastMessages = messages
	return f.displayResponse, nil
}

func (f *fakeDisplayChatClient) ChatCompletionInformationChat(ctx context.Context, model string, messages []ai.Message) (*ai.InformationChatResponse, error) {
	f.lastMessages = messages
	return f.chatResponse, nil
}

func (f *fakeDisplayChatClient) ChatCompletionComparison(ctx context.Context, model string, messages []ai.Message) (*ai.ComparisonContent, error) {
	f.lastMessages = messages
	return f.comparisonResponse, nil
}

func newTestDisplayUsecase(chat *fakeDisplayChatClient, information *domain.Information, sources []*domain.Source) (*DisplayUsecase, *fakePreferenceRepository) {
	uc, preferenceRepo, _ := newTestDisplayUsecaseWithRecipients(chat, information, sources)
	return uc, preferenceRepo
}

func newTestDisplayUsecaseWithRecipients(chat *fakeDisplayChatClient, information *domain.Information, sources []*domain.Source) (*DisplayUsecase, *fakePreferenceRepository, *fakeDisplayRecipientRepository) {
	infoRepo := &fakeDisplayInformationRepository{information: information}
	sourceRepo := &fakeDisplaySourceRepository{sources: sources}
	optionRepo := &fakeDisplayOptionRepository{optionsBySource: map[uuid.UUID][]*domain.Option{}}
	preferenceRepo := &fakePreferenceRepository{}
	recipientRepo := newFakeDisplayRecipientRepository()

	uc := NewDisplayUsecase(chat, "opus", infoRepo, sourceRepo, optionRepo, preferenceRepo, recipientRepo)
	return uc, preferenceRepo, recipientRepo
}

func newTestPublicInformation(t *testing.T) *domain.Information {
	t.Helper()
	information, err := domain.NewInformation(uuid.New(), uuid.New(), "旅行のお知らせ", domain.InformationAccessTypePublic, domain.InformationResponsePolicyAnonymous)
	if err != nil {
		t.Fatalf("NewInformation() error = %v", err)
	}
	return information
}

func TestDisplayUsecase_GenerateDisplay(t *testing.T) {
	information := newTestPublicInformation(t)
	source, err := domain.NewSource(information.ID(), domain.SourceTypeSchedule, "date", "2026-10-01", domain.SourceStatusConfirmed, nil)
	if err != nil {
		t.Fatalf("NewSource() error = %v", err)
	}

	chat := &fakeDisplayChatClient{displayResponse: &ai.DisplayContent{Title: "お知らせ", Body: "本文"}}
	uc, preferenceRepo := newTestDisplayUsecase(chat, information, []*domain.Source{source})

	recipientID := uuid.New()
	out, err := uc.GenerateDisplay(context.Background(), GenerateDisplayInput{
		InformationID: information.ID(),
		RecipientID:   recipientID,
	})
	if err != nil {
		t.Fatalf("GenerateDisplay() error = %v", err)
	}

	if out.Title != "お知らせ" || out.Body != "本文" {
		t.Errorf("out = %+v, unexpected", out)
	}
	if out.HasPreference {
		t.Error("HasPreference = true, want false (recipient has no preference yet)")
	}
	if len(preferenceRepo.saved) != 0 {
		t.Error("preference was saved, want no save for a plain display generation")
	}
}

func TestDisplayUsecase_GenerateDisplay_HasPreferenceTrueWhenPreferenceExists(t *testing.T) {
	information := newTestPublicInformation(t)
	source, err := domain.NewSource(information.ID(), domain.SourceTypeSchedule, "date", "2026-10-01", domain.SourceStatusConfirmed, nil)
	if err != nil {
		t.Fatalf("NewSource() error = %v", err)
	}

	recipientID := uuid.New()
	preference, err := domain.NewPreference(recipientID)
	if err != nil {
		t.Fatalf("NewPreference() error = %v", err)
	}
	if err := preference.Set("reading_level", "easy"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	chat := &fakeDisplayChatClient{displayResponse: &ai.DisplayContent{Title: "お知らせ", Body: "本文"}}
	infoRepo := &fakeDisplayInformationRepository{information: information}
	sourceRepo := &fakeDisplaySourceRepository{sources: []*domain.Source{source}}
	optionRepo := &fakeDisplayOptionRepository{optionsBySource: map[uuid.UUID][]*domain.Option{}}
	preferenceRepo := &fakePreferenceRepository{preference: preference}
	uc := NewDisplayUsecase(chat, "opus", infoRepo, sourceRepo, optionRepo, preferenceRepo, newFakeDisplayRecipientRepository())

	out, err := uc.GenerateDisplay(context.Background(), GenerateDisplayInput{
		InformationID: information.ID(),
		RecipientID:   recipientID,
	})
	if err != nil {
		t.Fatalf("GenerateDisplay() error = %v", err)
	}
	if !out.HasPreference {
		t.Error("HasPreference = false, want true (recipient already has a preference)")
	}
}

func TestDisplayUsecase_GenerateComparison_And_ApplyPreference(t *testing.T) {
	information := newTestPublicInformation(t)
	source, err := domain.NewSource(information.ID(), domain.SourceTypeSchedule, "date", "2026-10-01", domain.SourceStatusConfirmed, nil)
	if err != nil {
		t.Fatalf("NewSource() error = %v", err)
	}

	chat := &fakeDisplayChatClient{
		comparisonResponse: &ai.ComparisonContent{
			PatternA: ai.ComparisonPattern{Value: "easy", Title: "お知らせ", Body: "かんたんな本文"},
			PatternB: ai.ComparisonPattern{Value: "detailed", Title: "お知らせ", Body: "詳しい本文"},
		},
	}
	uc, preferenceRepo := newTestDisplayUsecase(chat, information, []*domain.Source{source})

	recipientID := uuid.New()

	generated, err := uc.GenerateComparison(context.Background(), GenerateComparisonInput{
		InformationID: information.ID(),
		RecipientID:   recipientID,
		Key:           "reading_level",
	})
	if err != nil {
		t.Fatalf("GenerateComparison() error = %v", err)
	}
	if generated.PatternA.Value != "easy" || generated.PatternB.Value != "detailed" {
		t.Errorf("generated = %+v, unexpected", generated)
	}
	if generated.PatternB.Display.Body != "詳しい本文" {
		t.Errorf("PatternB.Display.Body = %q, want %q", generated.PatternB.Display.Body, "詳しい本文")
	}
	if len(preferenceRepo.saved) != 0 {
		t.Error("preference was saved during GenerateComparison, want no save until ApplyPreference")
	}

	err = uc.ApplyPreference(context.Background(), ApplyPreferenceInput{
		InformationID: information.ID(),
		RecipientID:   recipientID,
		Key:           "reading_level",
		Value:         generated.PatternB.Value,
	})
	if err != nil {
		t.Fatalf("ApplyPreference() error = %v", err)
	}

	if len(preferenceRepo.saved) != 1 {
		t.Fatalf("len(saved) = %d, want 1", len(preferenceRepo.saved))
	}
	v, ok := preferenceRepo.saved[0].Get("reading_level")
	if !ok || v != "detailed" {
		t.Errorf("saved preference reading_level = (%q, %v), want (\"detailed\", true)", v, ok)
	}
}

func TestDisplayUsecase_Chat_ReturnsAnswerWithoutUpdatingPreference(t *testing.T) {
	information := newTestPublicInformation(t)

	chat := &fakeDisplayChatClient{
		chatResponse: &ai.InformationChatResponse{Answer: "集合時間は10時です。"},
	}
	uc, preferenceRepo := newTestDisplayUsecase(chat, information, nil)

	out, err := uc.Chat(context.Background(), ChatInput{
		InformationID: information.ID(),
		RecipientID:   uuid.New(),
		UserInput:     "集合時間は何時ですか？",
	})
	if err != nil {
		t.Fatalf("Chat() error = %v", err)
	}

	if out.Answer != "集合時間は10時です。" {
		t.Errorf("Answer = %q, want %q", out.Answer, "集合時間は10時です。")
	}
	if len(preferenceRepo.saved) != 0 {
		t.Error("preference was saved, want Chat to never write Preference (only A/B selection may)")
	}
	if len(out.Messages) != 3 {
		t.Errorf("len(Messages) = %d, want 3 (system, user, assistant)", len(out.Messages))
	}
}

func TestDisplayUsecase_Chat_NoDuplicateSystemPromptOnSecondTurn(t *testing.T) {
	information := newTestPublicInformation(t)

	chat := &fakeDisplayChatClient{
		chatResponse: &ai.InformationChatResponse{Answer: "本文の内容です。"},
	}
	uc, _ := newTestDisplayUsecase(chat, information, nil)
	recipientID := uuid.New()

	out, err := uc.Chat(context.Background(), ChatInput{
		InformationID: information.ID(),
		RecipientID:   recipientID,
		UserInput:     "この資料について教えてください",
	})
	if err != nil {
		t.Fatalf("Chat() error = %v", err)
	}
	if len(chat.lastMessages) == 0 || chat.lastMessages[0].Role != ai.RoleSystem {
		t.Fatalf("lastMessages[0] is not a system message: %+v", chat.lastMessages)
	}

	if _, err := uc.Chat(context.Background(), ChatInput{
		InformationID: information.ID(),
		RecipientID:   recipientID,
		Messages:      out.Messages,
		UserInput:     "ありがとう",
	}); err != nil {
		t.Fatalf("Chat() (2nd turn) error = %v", err)
	}
	systemCount := 0
	for _, m := range chat.lastMessages {
		if m.Role == ai.RoleSystem {
			systemCount++
		}
	}
	if systemCount != 1 {
		t.Errorf("systemCount = %d, want 1 (no duplicate system prompt on 2nd turn)", systemCount)
	}
}

func TestDisplayUsecase_GenerateDisplay_RestrictedForbidsNonRecipient(t *testing.T) {
	information, err := domain.NewInformation(uuid.New(), uuid.New(), "旅行のお知らせ", domain.InformationAccessTypeRestricted, domain.InformationResponsePolicyAuthenticated)
	if err != nil {
		t.Fatalf("NewInformation() error = %v", err)
	}

	chat := &fakeDisplayChatClient{displayResponse: &ai.DisplayContent{Title: "お知らせ", Body: "本文"}}
	uc, _, _ := newTestDisplayUsecaseWithRecipients(chat, information, nil)

	_, err = uc.GenerateDisplay(context.Background(), GenerateDisplayInput{
		InformationID: information.ID(),
		RecipientID:   uuid.New(),
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("GenerateDisplay() error = %v, want ErrForbidden", err)
	}
}

func TestDisplayUsecase_ListSources(t *testing.T) {
	information := newTestPublicInformation(t)

	radio := domain.SourceInteractionTypeRadio
	source, err := domain.NewSource(information.ID(), domain.SourceTypeSchedule, "date", "2026-10-01", domain.SourceStatusConfirmed, &radio)
	if err != nil {
		t.Fatalf("NewSource() error = %v", err)
	}
	option, err := domain.NewOption(source.ID(), "参加する", 0)
	if err != nil {
		t.Fatalf("NewOption() error = %v", err)
	}

	infoRepo := &fakeDisplayInformationRepository{information: information}
	sourceRepo := &fakeDisplaySourceRepository{sources: []*domain.Source{source}}
	optionRepo := &fakeDisplayOptionRepository{optionsBySource: map[uuid.UUID][]*domain.Option{source.ID(): {option}}}
	uc := NewDisplayUsecase(&fakeDisplayChatClient{}, "opus", infoRepo, sourceRepo, optionRepo, &fakePreferenceRepository{}, newFakeDisplayRecipientRepository())

	viewerID := uuid.New()
	views, err := uc.ListSources(context.Background(), information.ID(), &viewerID)
	if err != nil {
		t.Fatalf("ListSources() error = %v", err)
	}
	if len(views) != 1 {
		t.Fatalf("len(views) = %d, want 1", len(views))
	}
	if views[0].Key != "date" || views[0].InteractionType == nil || *views[0].InteractionType != "radio" {
		t.Errorf("views[0] = %+v, unexpected", views[0])
	}
	if len(views[0].Options) != 1 || views[0].Options[0].Value != "参加する" {
		t.Errorf("views[0].Options = %+v, unexpected", views[0].Options)
	}
}

func TestDisplayUsecase_ListSources_RestrictedForbidsNonRecipient(t *testing.T) {
	information, err := domain.NewInformation(uuid.New(), uuid.New(), "旅行のお知らせ", domain.InformationAccessTypeRestricted, domain.InformationResponsePolicyAuthenticated)
	if err != nil {
		t.Fatalf("NewInformation() error = %v", err)
	}

	uc, _, _ := newTestDisplayUsecaseWithRecipients(&fakeDisplayChatClient{}, information, nil)

	nonRecipient := uuid.New()
	if _, err := uc.ListSources(context.Background(), information.ID(), &nonRecipient); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("ListSources() error = %v, want ErrForbidden", err)
	}
}

func TestDisplayUsecase_ListSources_RestrictedForbidsAnonymous(t *testing.T) {
	information, err := domain.NewInformation(uuid.New(), uuid.New(), "旅行のお知らせ", domain.InformationAccessTypeRestricted, domain.InformationResponsePolicyAuthenticated)
	if err != nil {
		t.Fatalf("NewInformation() error = %v", err)
	}

	uc, _, _ := newTestDisplayUsecaseWithRecipients(&fakeDisplayChatClient{}, information, nil)

	if _, err := uc.ListSources(context.Background(), information.ID(), nil); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("ListSources() error = %v, want ErrForbidden", err)
	}
}

func TestDisplayUsecase_ListSources_AllowsAnonymousForPublicAnonymous(t *testing.T) {
	information, err := domain.NewInformation(uuid.New(), uuid.New(), "旅行のお知らせ", domain.InformationAccessTypePublic, domain.InformationResponsePolicyAnonymous)
	if err != nil {
		t.Fatalf("NewInformation() error = %v", err)
	}

	radio := domain.SourceInteractionTypeRadio
	source, err := domain.NewSource(information.ID(), domain.SourceTypeSchedule, "date", "2026-10-01", domain.SourceStatusConfirmed, &radio)
	if err != nil {
		t.Fatalf("NewSource() error = %v", err)
	}

	infoRepo := &fakeDisplayInformationRepository{information: information}
	sourceRepo := &fakeDisplaySourceRepository{sources: []*domain.Source{source}}
	optionRepo := &fakeDisplayOptionRepository{}
	uc := NewDisplayUsecase(&fakeDisplayChatClient{}, "opus", infoRepo, sourceRepo, optionRepo, &fakePreferenceRepository{}, newFakeDisplayRecipientRepository())

	if _, err := uc.ListSources(context.Background(), information.ID(), nil); err != nil {
		t.Fatalf("ListSources() error = %v, want nil (anonymous allowed for public+anonymous)", err)
	}
}

func TestDisplayUsecase_GenerateDisplay_RestrictedAllowsRecipient(t *testing.T) {
	information, err := domain.NewInformation(uuid.New(), uuid.New(), "旅行のお知らせ", domain.InformationAccessTypeRestricted, domain.InformationResponsePolicyAuthenticated)
	if err != nil {
		t.Fatalf("NewInformation() error = %v", err)
	}

	chat := &fakeDisplayChatClient{displayResponse: &ai.DisplayContent{Title: "お知らせ", Body: "本文"}}
	uc, _, recipientRepo := newTestDisplayUsecaseWithRecipients(chat, information, nil)

	recipientID := uuid.New()
	recipient, err := domain.NewRecipient(information.ID(), recipientID)
	if err != nil {
		t.Fatalf("NewRecipient() error = %v", err)
	}
	if err := recipientRepo.Create(context.Background(), recipient); err != nil {
		t.Fatalf("recipientRepo.Create() error = %v", err)
	}

	out, err := uc.GenerateDisplay(context.Background(), GenerateDisplayInput{
		InformationID: information.ID(),
		RecipientID:   recipientID,
	})
	if err != nil {
		t.Fatalf("GenerateDisplay() error = %v", err)
	}
	if out.Title != "お知らせ" || out.Body != "本文" {
		t.Errorf("out = %+v, unexpected", out)
	}
}
