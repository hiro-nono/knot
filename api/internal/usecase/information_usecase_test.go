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

type fakeChatClient struct {
	response     *ai.StructuredInformation
	err          error
	lastMessages []ai.Message
}

func (f *fakeChatClient) ChatCompletionStructured(ctx context.Context, model string, messages []ai.Message) (*ai.StructuredInformation, error) {
	f.lastMessages = messages
	if f.err != nil {
		return nil, f.err
	}
	return f.response, nil
}

type fakeInformationRepository struct {
	created []*domain.Information
}

func (r *fakeInformationRepository) Create(ctx context.Context, information *domain.Information) error {
	r.created = append(r.created, information)
	return nil
}

func (r *fakeInformationRepository) Get(ctx context.Context, id uuid.UUID) (*domain.Information, error) {
	return nil, domain.ErrNotFound
}

func (r *fakeInformationRepository) ListByAccountID(ctx context.Context, accountID uuid.UUID) ([]*domain.Information, error) {
	var out []*domain.Information
	for _, information := range r.created {
		if information.AccountID() == accountID {
			out = append(out, information)
		}
	}
	return out, nil
}

type fakeSourceRepository struct {
	created []*domain.Source
}

func (r *fakeSourceRepository) Create(ctx context.Context, source *domain.Source) error {
	r.created = append(r.created, source)
	return nil
}

func (r *fakeSourceRepository) List(ctx context.Context, informationID uuid.UUID) ([]*domain.Source, error) {
	return nil, nil
}

type fakeOptionRepository struct {
	created []*domain.Option
}

func (r *fakeOptionRepository) Create(ctx context.Context, option *domain.Option) error {
	r.created = append(r.created, option)
	return nil
}

func (r *fakeOptionRepository) List(ctx context.Context, sourceID uuid.UUID) ([]*domain.Option, error) {
	return nil, nil
}

type fakeTransactionManager struct {
	called bool
}

func (m *fakeTransactionManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	m.called = true
	return fn(ctx)
}

type fakeRecipientRepository struct {
	created []*domain.Recipient
}

func (r *fakeRecipientRepository) Create(ctx context.Context, recipient *domain.Recipient) error {
	r.created = append(r.created, recipient)
	return nil
}

func (r *fakeRecipientRepository) Exists(ctx context.Context, informationID, userID uuid.UUID) (bool, error) {
	for _, rec := range r.created {
		if rec.InformationID() == informationID && rec.UserID() == userID {
			return true, nil
		}
	}
	return false, nil
}

func (r *fakeRecipientRepository) ListByInformationID(ctx context.Context, informationID uuid.UUID) ([]*domain.Recipient, error) {
	var out []*domain.Recipient
	for _, rec := range r.created {
		if rec.InformationID() == informationID {
			out = append(out, rec)
		}
	}
	return out, nil
}

func (r *fakeRecipientRepository) Delete(ctx context.Context, informationID, userID uuid.UUID) error {
	for i, rec := range r.created {
		if rec.InformationID() == informationID && rec.UserID() == userID {
			r.created = append(r.created[:i], r.created[i+1:]...)
			return nil
		}
	}
	return domain.ErrNotFound
}

func strPtr(s string) *string {
	return &s
}

func newTestUsecase(chat chatClient) (*InformationUsecase, *fakeInformationRepository, *fakeSourceRepository, *fakeOptionRepository, *fakeTransactionManager) {
	infoRepo := &fakeInformationRepository{}
	sourceRepo := &fakeSourceRepository{}
	optionRepo := &fakeOptionRepository{}
	recipientRepo := &fakeRecipientRepository{}
	txManager := &fakeTransactionManager{}

	uc := NewInformationUsecase(chat, "orcarouter/auto", infoRepo, sourceRepo, optionRepo, recipientRepo, txManager)

	return uc, infoRepo, sourceRepo, optionRepo, txManager
}

func testProcessInformationInput(messages []Message, userInput string) ProcessInformationInput {
	return ProcessInformationInput{
		AccountID:       uuid.New(),
		CreatedByUserID: uuid.New(),
		AccessType:      "restricted",
		ResponsePolicy:  "authenticated",
		Messages:        messages,
		UserInput:       userInput,
	}
}

func TestInformationUsecase_ProcessInformation_NotConfirmed(t *testing.T) {
	chat := &fakeChatClient{
		response: &ai.StructuredInformation{
			Title: "旅行のお知らせ",
			Sources: []ai.StructuredSource{
				{
					Type:     "schedule",
					Key:      "departure_date",
					Value:    "",
					Status:   "undecided",
					Question: strPtr("出発日はいつですか？"),
					Options:  []ai.StructuredOption{},
				},
			},
		},
	}

	uc, infoRepo, sourceRepo, optionRepo, txManager := newTestUsecase(chat)

	out, err := uc.ProcessInformation(context.Background(), testProcessInformationInput(nil, "旅行の案内をしたい"))
	if err != nil {
		t.Fatalf("ProcessInformation() error = %v", err)
	}

	if out.Confirmed {
		t.Errorf("Confirmed = true, want false")
	}
	if out.InformationID != "" {
		t.Errorf("InformationID = %q, want empty", out.InformationID)
	}
	if len(out.Messages) != 3 {
		t.Fatalf("len(Messages) = %d, want 3 (system, user, assistant)", len(out.Messages))
	}
	if txManager.called {
		t.Error("WithinTransaction was called, want it not to be called when not confirmed")
	}
	if len(infoRepo.created) != 0 || len(sourceRepo.created) != 0 || len(optionRepo.created) != 0 {
		t.Error("repositories were called, want none to be called when not confirmed")
	}
}

func TestInformationUsecase_ProcessInformation_Confirmed(t *testing.T) {
	radio := "radio"
	chat := &fakeChatClient{
		response: &ai.StructuredInformation{
			Title: "旅行のお知らせ",
			Sources: []ai.StructuredSource{
				{
					Type:            "schedule",
					Key:             "departure_date",
					Value:           "2026-10-01",
					Status:          "confirmed",
					InteractionType: &radio,
					Options: []ai.StructuredOption{
						{Value: "参加する", SortOrder: 0},
						{Value: "参加しない", SortOrder: 1},
					},
				},
			},
		},
	}

	uc, infoRepo, sourceRepo, optionRepo, txManager := newTestUsecase(chat)

	out, err := uc.ProcessInformation(context.Background(), testProcessInformationInput(nil, "10/1に旅行します、参加確認をとりたい"))
	if err != nil {
		t.Fatalf("ProcessInformation() error = %v", err)
	}

	if !out.Confirmed {
		t.Fatalf("Confirmed = false, want true")
	}
	if out.InformationID == "" {
		t.Error("InformationID is empty, want a value")
	}
	if !txManager.called {
		t.Error("WithinTransaction was not called, want it to be called when confirmed")
	}
	if len(infoRepo.created) != 1 {
		t.Fatalf("len(infoRepo.created) = %d, want 1", len(infoRepo.created))
	}
	if len(sourceRepo.created) != 1 {
		t.Fatalf("len(sourceRepo.created) = %d, want 1", len(sourceRepo.created))
	}
	if len(optionRepo.created) != 2 {
		t.Fatalf("len(optionRepo.created) = %d, want 2", len(optionRepo.created))
	}
	if infoRepo.created[0].ID().String() != out.InformationID {
		t.Errorf("created information id = %s, want %s", infoRepo.created[0].ID().String(), out.InformationID)
	}
}

func TestInformationUsecase_ProcessInformation_UsesPredefinedKeysOnFirstTurnOnly(t *testing.T) {
	chat := &fakeChatClient{
		response: &ai.StructuredInformation{
			Title: "旅行のお知らせ",
			Sources: []ai.StructuredSource{
				{Type: "schedule", Key: "date", Status: "confirmed", Value: "2026-10-01", Options: []ai.StructuredOption{}},
			},
		},
	}

	uc, _, _, _, _ := newTestUsecase(chat)

	// 1ターン目: システムプロンプトにdomain.PredefinedSourceKeysの内容が含まれること。
	out, err := uc.ProcessInformation(context.Background(), testProcessInformationInput(nil, "旅行の案内をしたい"))
	if err != nil {
		t.Fatalf("ProcessInformation() error = %v", err)
	}
	if len(chat.lastMessages) == 0 || chat.lastMessages[0].Role != ai.RoleSystem {
		t.Fatalf("lastMessages[0] is not a system message: %+v", chat.lastMessages)
	}
	for _, key := range domain.PredefinedSourceKeys[domain.SourceTypeSchedule] {
		if !strings.Contains(chat.lastMessages[0].Content, key) {
			t.Errorf("system prompt does not contain predefined key %q:\n%s", key, chat.lastMessages[0].Content)
		}
	}

	// 2ターン目: 既存の会話履歴を渡した場合、systemメッセージを重複して追加しないこと。
	_, err = uc.ProcessInformation(context.Background(), testProcessInformationInput(out.Messages, "10/1です"))
	if err != nil {
		t.Fatalf("ProcessInformation() (2nd turn) error = %v", err)
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

func TestInformationUsecase_ListMine(t *testing.T) {
	uc, infoRepo, _, _, _ := newTestUsecase(&fakeChatClient{})

	accountID := uuid.New()
	other, err := domain.NewInformation(uuid.New(), uuid.New(), "他人のお知らせ", domain.InformationAccessTypePublic, domain.InformationResponsePolicyAnonymous)
	if err != nil {
		t.Fatalf("NewInformation() error = %v", err)
	}
	mine, err := domain.NewInformation(accountID, uuid.New(), "自分のお知らせ", domain.InformationAccessTypeRestricted, domain.InformationResponsePolicyAuthenticated)
	if err != nil {
		t.Fatalf("NewInformation() error = %v", err)
	}
	if err := infoRepo.Create(context.Background(), other); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if err := infoRepo.Create(context.Background(), mine); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	views, err := uc.ListMine(context.Background(), accountID)
	if err != nil {
		t.Fatalf("ListMine() error = %v", err)
	}
	if len(views) != 1 {
		t.Fatalf("len(views) = %d, want 1", len(views))
	}
	if views[0].ID != mine.ID().String() || views[0].Title != "自分のお知らせ" {
		t.Errorf("views[0] = %+v, unexpected", views[0])
	}
}

func TestInformationUsecase_ProcessInformation_ChatError(t *testing.T) {
	chat := &fakeChatClient{err: errors.New("boom")}

	uc, _, _, _, txManager := newTestUsecase(chat)

	_, err := uc.ProcessInformation(context.Background(), testProcessInformationInput(nil, "hello"))
	if err == nil {
		t.Fatal("ProcessInformation() error = nil, want error")
	}
	if txManager.called {
		t.Error("WithinTransaction was called, want it not to be called on chat error")
	}
}
