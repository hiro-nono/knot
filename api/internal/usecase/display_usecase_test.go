package usecase

import (
	"context"
	"errors"
	"strings"
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
	displayResponse *ai.DisplayContent
	chatResponse    *ai.PreferenceChatResponse
	lastMessages    []ai.Message
}

func (f *fakeDisplayChatClient) ChatCompletionDisplay(ctx context.Context, model string, messages []ai.Message) (*ai.DisplayContent, error) {
	f.lastMessages = messages
	return f.displayResponse, nil
}

func (f *fakeDisplayChatClient) ChatCompletionPreferenceChat(ctx context.Context, model string, messages []ai.Message) (*ai.PreferenceChatResponse, error) {
	f.lastMessages = messages
	return f.chatResponse, nil
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
	if len(preferenceRepo.saved) != 0 {
		t.Error("preference was saved, want no save for a plain display generation")
	}
}

func TestDisplayUsecase_PrepareComparison_And_SelectComparison(t *testing.T) {
	information := newTestPublicInformation(t)
	source, err := domain.NewSource(information.ID(), domain.SourceTypeSchedule, "date", "2026-10-01", domain.SourceStatusConfirmed, nil)
	if err != nil {
		t.Fatalf("NewSource() error = %v", err)
	}

	chat := &fakeDisplayChatClient{displayResponse: &ai.DisplayContent{Title: "お知らせ", Body: "本文"}}
	uc, preferenceRepo := newTestDisplayUsecase(chat, information, []*domain.Source{source})

	recipientID := uuid.New()
	comparison := PreferenceComparison{Key: "reading_level", ValueA: "easy", ValueB: "detailed"}

	prepared, err := uc.PrepareComparison(context.Background(), PrepareComparisonInput{
		InformationID: information.ID(),
		RecipientID:   recipientID,
		Comparison:    comparison,
	})
	if err != nil {
		t.Fatalf("PrepareComparison() error = %v", err)
	}
	if prepared.Comparison != comparison {
		t.Errorf("Comparison = %+v, want %+v", prepared.Comparison, comparison)
	}
	if len(preferenceRepo.saved) != 0 {
		t.Error("preference was saved during PrepareComparison, want no save until SelectComparison")
	}

	err = uc.SelectComparison(context.Background(), SelectComparisonInput{
		InformationID: information.ID(),
		RecipientID:   recipientID,
		Comparison:    comparison,
		Selected:      "a",
	})
	if err != nil {
		t.Fatalf("SelectComparison() error = %v", err)
	}

	if len(preferenceRepo.saved) != 1 {
		t.Fatalf("len(saved) = %d, want 1", len(preferenceRepo.saved))
	}
	v, ok := preferenceRepo.saved[0].Get("reading_level")
	if !ok || v != "easy" {
		t.Errorf("saved preference reading_level = (%q, %v), want (\"easy\", true)", v, ok)
	}

	if err := uc.SelectComparison(context.Background(), SelectComparisonInput{
		InformationID: information.ID(),
		RecipientID:   recipientID,
		Comparison:    comparison,
		Selected:      "invalid",
	}); err == nil {
		t.Error(`SelectComparison(Selected: "invalid") error = nil, want error`)
	}
}

func TestDisplayUsecase_Chat_Persistent(t *testing.T) {
	information := newTestPublicInformation(t)

	persistent := true
	key := "reading_level"
	value := "easy"
	chat := &fakeDisplayChatClient{
		chatResponse: &ai.PreferenceChatResponse{
			Display:         ai.DisplayContent{Title: "お知らせ", Body: "かんたんな本文"},
			IsPersistent:    &persistent,
			PreferenceKey:   &key,
			PreferenceValue: &value,
		},
	}
	uc, preferenceRepo := newTestDisplayUsecase(chat, information, nil)

	out, err := uc.Chat(context.Background(), ChatInput{
		InformationID: information.ID(),
		RecipientID:   uuid.New(),
		UserInput:     "もっと簡単にして、これからずっとそうして",
	})
	if err != nil {
		t.Fatalf("Chat() error = %v", err)
	}

	if !out.PreferenceUpdated {
		t.Error("PreferenceUpdated = false, want true")
	}
	if out.NeedsConfirmation {
		t.Error("NeedsConfirmation = true, want false")
	}
	if len(preferenceRepo.saved) != 1 {
		t.Fatalf("len(saved) = %d, want 1", len(preferenceRepo.saved))
	}
	if v, _ := preferenceRepo.saved[0].Get("reading_level"); v != "easy" {
		t.Errorf("saved reading_level = %q, want \"easy\"", v)
	}
	if len(out.Messages) != 3 {
		t.Errorf("len(Messages) = %d, want 3 (system, user, assistant)", len(out.Messages))
	}
}

func TestDisplayUsecase_Chat_NeedsConfirmation(t *testing.T) {
	information := newTestPublicInformation(t)

	question := "今後も常にこの設定にしますか？"
	chat := &fakeDisplayChatClient{
		chatResponse: &ai.PreferenceChatResponse{
			Display:              ai.DisplayContent{Title: "お知らせ", Body: "本文"},
			ConfirmationQuestion: &question,
		},
	}
	uc, preferenceRepo := newTestDisplayUsecase(chat, information, nil)

	out, err := uc.Chat(context.Background(), ChatInput{
		InformationID: information.ID(),
		RecipientID:   uuid.New(),
		UserInput:     "簡単にして",
	})
	if err != nil {
		t.Fatalf("Chat() error = %v", err)
	}

	if out.PreferenceUpdated {
		t.Error("PreferenceUpdated = true, want false")
	}
	if !out.NeedsConfirmation {
		t.Error("NeedsConfirmation = false, want true")
	}
	if out.ConfirmationQuestion == nil || *out.ConfirmationQuestion != question {
		t.Errorf("ConfirmationQuestion = %v, want %q", out.ConfirmationQuestion, question)
	}
	if len(preferenceRepo.saved) != 0 {
		t.Error("preference was saved, want no save while confirmation is pending")
	}
}

func TestDisplayUsecase_Chat_IncludesPredefinedPreferenceKeysOnFirstTurnOnly(t *testing.T) {
	information := newTestPublicInformation(t)

	chat := &fakeDisplayChatClient{
		chatResponse: &ai.PreferenceChatResponse{
			Display: ai.DisplayContent{Title: "お知らせ", Body: "本文"},
		},
	}
	uc, _ := newTestDisplayUsecase(chat, information, nil)
	recipientID := uuid.New()

	out, err := uc.Chat(context.Background(), ChatInput{
		InformationID: information.ID(),
		RecipientID:   recipientID,
		UserInput:     "もっと簡単にして",
	})
	if err != nil {
		t.Fatalf("Chat() error = %v", err)
	}
	if len(chat.lastMessages) == 0 || chat.lastMessages[0].Role != ai.RoleSystem {
		t.Fatalf("lastMessages[0] is not a system message: %+v", chat.lastMessages)
	}
	for _, key := range domain.PredefinedPreferenceKeys {
		if !strings.Contains(chat.lastMessages[0].Content, key) {
			t.Errorf("system prompt does not contain predefined preference key %q:\n%s", key, chat.lastMessages[0].Content)
		}
	}

	if _, err := uc.Chat(context.Background(), ChatInput{
		InformationID: information.ID(),
		RecipientID:   recipientID,
		Messages:      out.Messages,
		UserInput:     "はい",
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
