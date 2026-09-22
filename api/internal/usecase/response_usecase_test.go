package usecase

import (
	"context"
	"errors"
	"testing"

	"knot-api/internal/domain"

	"uuid"
)

type fakeResponseRepository struct {
	created []*domain.Response
}

func (r *fakeResponseRepository) Create(ctx context.Context, response *domain.Response) error {
	r.created = append(r.created, response)
	return nil
}

func (r *fakeResponseRepository) GetByInformationAndUser(ctx context.Context, informationID, userID uuid.UUID) (*domain.Response, error) {
	for _, resp := range r.created {
		if resp.InformationID() == informationID && resp.UserID() != nil && *resp.UserID() == userID {
			return resp, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *fakeResponseRepository) ListByInformationID(ctx context.Context, informationID uuid.UUID) ([]*domain.Response, error) {
	var out []*domain.Response
	for _, resp := range r.created {
		if resp.InformationID() == informationID {
			out = append(out, resp)
		}
	}
	return out, nil
}

type responseTestFixture struct {
	information  *domain.Information
	radioSource  *domain.Source
	radioOptionA *domain.Option
	radioOptionB *domain.Option
	textSource   *domain.Source
	staticSource *domain.Source // interaction_typeが無く、応答を必要としない
}

func newResponseTestFixture(t *testing.T, accessType domain.InformationAccessType, responsePolicy domain.InformationResponsePolicy) *responseTestFixture {
	t.Helper()

	accountID := uuid.New()
	createdByUserID := uuid.New()
	information, err := domain.NewInformation(accountID, createdByUserID, "旅行のお知らせ", accessType, responsePolicy)
	if err != nil {
		t.Fatalf("NewInformation() error = %v", err)
	}

	radio := domain.SourceInteractionTypeRadio
	radioSource, err := domain.NewSource(information.ID(), domain.SourceTypeSchedule, "attend", "未回答", domain.SourceStatusConfirmed, &radio)
	if err != nil {
		t.Fatalf("NewSource(radio) error = %v", err)
	}
	optionA, err := domain.NewOption(radioSource.ID(), "参加する", 0)
	if err != nil {
		t.Fatalf("NewOption(A) error = %v", err)
	}
	optionB, err := domain.NewOption(radioSource.ID(), "参加しない", 1)
	if err != nil {
		t.Fatalf("NewOption(B) error = %v", err)
	}

	text := domain.SourceInteractionTypeText
	textSource, err := domain.NewSource(information.ID(), domain.SourceTypeFact, "comment", "コメントをお願いします", domain.SourceStatusConfirmed, &text)
	if err != nil {
		t.Fatalf("NewSource(text) error = %v", err)
	}

	staticSource, err := domain.NewSource(information.ID(), domain.SourceTypeFact, "location", "東京", domain.SourceStatusConfirmed, nil)
	if err != nil {
		t.Fatalf("NewSource(static) error = %v", err)
	}

	return &responseTestFixture{
		information:  information,
		radioSource:  radioSource,
		radioOptionA: optionA,
		radioOptionB: optionB,
		textSource:   textSource,
		staticSource: staticSource,
	}
}

func newTestResponseUsecase(fixture *responseTestFixture) (*ResponseUsecase, *fakeResponseRepository, *fakeDisplayRecipientRepository) {
	infoRepo := &fakeDisplayInformationRepository{information: fixture.information}
	sourceRepo := &fakeDisplaySourceRepository{sources: []*domain.Source{fixture.radioSource, fixture.textSource, fixture.staticSource}}
	optionRepo := &fakeDisplayOptionRepository{optionsBySource: map[uuid.UUID][]*domain.Option{
		fixture.radioSource.ID(): {fixture.radioOptionA, fixture.radioOptionB},
	}}
	recipientRepo := newFakeDisplayRecipientRepository()
	responseRepo := &fakeResponseRepository{}
	txManager := &fakeTransactionManager{}

	uc := NewResponseUsecase(infoRepo, sourceRepo, optionRepo, recipientRepo, responseRepo, txManager)
	return uc, responseRepo, recipientRepo
}

func strPtrResp(s string) *string { return &s }

func uuidPtrResp(id uuid.UUID) *uuid.UUID { return &id }

func TestResponseUsecase_SubmitResponse_Success(t *testing.T) {
	fixture := newResponseTestFixture(t, domain.InformationAccessTypePublic, domain.InformationResponsePolicyAuthenticated)
	uc, responseRepo, _ := newTestResponseUsecase(fixture)

	userID := uuid.New()
	optionID := fixture.radioOptionA.ID().String()
	view, err := uc.SubmitResponse(context.Background(), SubmitResponseInput{
		InformationID: fixture.information.ID(),
		UserID:        &userID,
		Items: []ResponseItemInput{
			{SourceID: fixture.radioSource.ID().String(), OptionID: &optionID},
			{SourceID: fixture.textSource.ID().String(), Value: strPtrResp("よろしくお願いします")},
		},
	})
	if err != nil {
		t.Fatalf("SubmitResponse() error = %v", err)
	}
	if view.UserID == nil || *view.UserID != userID.String() {
		t.Errorf("UserID = %v, want %q", view.UserID, userID.String())
	}
	if len(view.Items) != 2 {
		t.Fatalf("len(Items) = %d, want 2", len(view.Items))
	}
	if len(responseRepo.created) != 1 {
		t.Fatalf("len(responseRepo.created) = %d, want 1", len(responseRepo.created))
	}
}

func TestResponseUsecase_SubmitResponse_RestrictedForbidsNonRecipient(t *testing.T) {
	fixture := newResponseTestFixture(t, domain.InformationAccessTypeRestricted, domain.InformationResponsePolicyAuthenticated)
	uc, _, _ := newTestResponseUsecase(fixture)

	optionID := fixture.radioOptionA.ID().String()
	_, err := uc.SubmitResponse(context.Background(), SubmitResponseInput{
		InformationID: fixture.information.ID(),
		UserID:        uuidPtrResp(uuid.New()),
		Items: []ResponseItemInput{
			{SourceID: fixture.radioSource.ID().String(), OptionID: &optionID},
		},
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("SubmitResponse() error = %v, want ErrForbidden", err)
	}
}

func TestResponseUsecase_SubmitResponse_RestrictedAllowsRecipient(t *testing.T) {
	fixture := newResponseTestFixture(t, domain.InformationAccessTypeRestricted, domain.InformationResponsePolicyAuthenticated)
	uc, _, recipientRepo := newTestResponseUsecase(fixture)

	userID := uuid.New()
	recipient, err := domain.NewRecipient(fixture.information.ID(), userID)
	if err != nil {
		t.Fatalf("NewRecipient() error = %v", err)
	}
	if err := recipientRepo.Create(context.Background(), recipient); err != nil {
		t.Fatalf("recipientRepo.Create() error = %v", err)
	}

	optionID := fixture.radioOptionA.ID().String()
	_, err = uc.SubmitResponse(context.Background(), SubmitResponseInput{
		InformationID: fixture.information.ID(),
		UserID:        &userID,
		Items: []ResponseItemInput{
			{SourceID: fixture.radioSource.ID().String(), OptionID: &optionID},
		},
	})
	if err != nil {
		t.Fatalf("SubmitResponse() error = %v", err)
	}
}

func TestResponseUsecase_SubmitResponse_UnknownSource(t *testing.T) {
	fixture := newResponseTestFixture(t, domain.InformationAccessTypePublic, domain.InformationResponsePolicyAuthenticated)
	uc, _, _ := newTestResponseUsecase(fixture)

	_, err := uc.SubmitResponse(context.Background(), SubmitResponseInput{
		InformationID: fixture.information.ID(),
		UserID:        uuidPtrResp(uuid.New()),
		Items: []ResponseItemInput{
			{SourceID: uuid.New().String(), Value: strPtrResp("hello")},
		},
	})
	if err == nil {
		t.Error("SubmitResponse() with unknown source error = nil, want error")
	}
}

func TestResponseUsecase_SubmitResponse_SourceWithoutInteractionType(t *testing.T) {
	fixture := newResponseTestFixture(t, domain.InformationAccessTypePublic, domain.InformationResponsePolicyAuthenticated)
	uc, _, _ := newTestResponseUsecase(fixture)

	_, err := uc.SubmitResponse(context.Background(), SubmitResponseInput{
		InformationID: fixture.information.ID(),
		UserID:        uuidPtrResp(uuid.New()),
		Items: []ResponseItemInput{
			{SourceID: fixture.staticSource.ID().String(), Value: strPtrResp("hello")},
		},
	})
	if err == nil {
		t.Error("SubmitResponse() on non-interactive source error = nil, want error")
	}
}

func TestResponseUsecase_SubmitResponse_RadioRequiresOptionID(t *testing.T) {
	fixture := newResponseTestFixture(t, domain.InformationAccessTypePublic, domain.InformationResponsePolicyAuthenticated)
	uc, _, _ := newTestResponseUsecase(fixture)

	_, err := uc.SubmitResponse(context.Background(), SubmitResponseInput{
		InformationID: fixture.information.ID(),
		UserID:        uuidPtrResp(uuid.New()),
		Items: []ResponseItemInput{
			{SourceID: fixture.radioSource.ID().String(), Value: strPtrResp("参加する")},
		},
	})
	if err == nil {
		t.Error("SubmitResponse() with value on radio source error = nil, want error")
	}
}

func TestResponseUsecase_SubmitResponse_OptionMustBelongToSource(t *testing.T) {
	fixture := newResponseTestFixture(t, domain.InformationAccessTypePublic, domain.InformationResponsePolicyAuthenticated)
	uc, _, _ := newTestResponseUsecase(fixture)

	foreignOptionID := uuid.New().String()
	_, err := uc.SubmitResponse(context.Background(), SubmitResponseInput{
		InformationID: fixture.information.ID(),
		UserID:        uuidPtrResp(uuid.New()),
		Items: []ResponseItemInput{
			{SourceID: fixture.radioSource.ID().String(), OptionID: &foreignOptionID},
		},
	})
	if err == nil {
		t.Error("SubmitResponse() with option not belonging to source error = nil, want error")
	}
}

func TestResponseUsecase_SubmitResponse_TextRequiresValue(t *testing.T) {
	fixture := newResponseTestFixture(t, domain.InformationAccessTypePublic, domain.InformationResponsePolicyAuthenticated)
	uc, _, _ := newTestResponseUsecase(fixture)

	optionID := fixture.radioOptionA.ID().String()
	_, err := uc.SubmitResponse(context.Background(), SubmitResponseInput{
		InformationID: fixture.information.ID(),
		UserID:        uuidPtrResp(uuid.New()),
		Items: []ResponseItemInput{
			{SourceID: fixture.textSource.ID().String(), OptionID: &optionID},
		},
	})
	if err == nil {
		t.Error("SubmitResponse() with option_id on text source error = nil, want error")
	}
}

func TestResponseUsecase_ListResponses_RequiresOwningAccount(t *testing.T) {
	fixture := newResponseTestFixture(t, domain.InformationAccessTypePublic, domain.InformationResponsePolicyAuthenticated)
	uc, responseRepo, _ := newTestResponseUsecase(fixture)

	response, err := domain.NewResponse(fixture.information.ID(), uuidPtrResp(uuid.New()))
	if err != nil {
		t.Fatalf("NewResponse() error = %v", err)
	}
	responseRepo.created = append(responseRepo.created, response)

	if _, err := uc.ListResponses(context.Background(), uuid.New(), fixture.information.ID()); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("ListResponses() by non-owner error = %v, want ErrForbidden", err)
	}

	views, err := uc.ListResponses(context.Background(), fixture.information.AccountID(), fixture.information.ID())
	if err != nil {
		t.Fatalf("ListResponses() error = %v", err)
	}
	if len(views) != 1 {
		t.Fatalf("len(views) = %d, want 1", len(views))
	}
}

func TestResponseUsecase_SubmitResponse_PublicAnonymousAllowsNilUserID(t *testing.T) {
	fixture := newResponseTestFixture(t, domain.InformationAccessTypePublic, domain.InformationResponsePolicyAnonymous)
	uc, responseRepo, _ := newTestResponseUsecase(fixture)

	view, err := uc.SubmitResponse(context.Background(), SubmitResponseInput{
		InformationID: fixture.information.ID(),
		UserID:        nil,
		Items: []ResponseItemInput{
			{SourceID: fixture.textSource.ID().String(), Value: strPtrResp("匿名です")},
		},
	})
	if err != nil {
		t.Fatalf("SubmitResponse() error = %v", err)
	}
	if view.UserID != nil {
		t.Errorf("UserID = %v, want nil", view.UserID)
	}
	if len(responseRepo.created) != 1 {
		t.Fatalf("len(responseRepo.created) = %d, want 1", len(responseRepo.created))
	}

	// public+anonymousでも、ログイン済みなら記録される。
	userID := uuid.New()
	view2, err := uc.SubmitResponse(context.Background(), SubmitResponseInput{
		InformationID: fixture.information.ID(),
		UserID:        &userID,
		Items: []ResponseItemInput{
			{SourceID: fixture.textSource.ID().String(), Value: strPtrResp("ログイン済みです")},
		},
	})
	if err != nil {
		t.Fatalf("SubmitResponse() error = %v", err)
	}
	if view2.UserID == nil || *view2.UserID != userID.String() {
		t.Errorf("UserID = %v, want %q", view2.UserID, userID.String())
	}
}

func TestResponseUsecase_SubmitResponse_PublicAuthenticatedForbidsNilUserID(t *testing.T) {
	fixture := newResponseTestFixture(t, domain.InformationAccessTypePublic, domain.InformationResponsePolicyAuthenticated)
	uc, _, _ := newTestResponseUsecase(fixture)

	_, err := uc.SubmitResponse(context.Background(), SubmitResponseInput{
		InformationID: fixture.information.ID(),
		UserID:        nil,
		Items: []ResponseItemInput{
			{SourceID: fixture.textSource.ID().String(), Value: strPtrResp("未ログインです")},
		},
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("SubmitResponse() error = %v, want ErrForbidden", err)
	}
}

func TestResponseUsecase_SubmitResponse_RestrictedForbidsNilUserID(t *testing.T) {
	fixture := newResponseTestFixture(t, domain.InformationAccessTypeRestricted, domain.InformationResponsePolicyAnonymous)
	uc, _, _ := newTestResponseUsecase(fixture)

	_, err := uc.SubmitResponse(context.Background(), SubmitResponseInput{
		InformationID: fixture.information.ID(),
		UserID:        nil,
		Items: []ResponseItemInput{
			{SourceID: fixture.textSource.ID().String(), Value: strPtrResp("未ログインです")},
		},
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("SubmitResponse() error = %v, want ErrForbidden (restricted always requires authentication regardless of response_policy)", err)
	}
}
