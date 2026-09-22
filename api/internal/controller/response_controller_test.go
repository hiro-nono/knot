package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"knot-api/internal/domain"
	"knot-api/internal/infrastructure/auth/supabase"
	"knot-api/internal/usecase"

	"uuid"
)

type configurableSourceRepository struct {
	sources []*domain.Source
}

func (r configurableSourceRepository) Create(ctx context.Context, source *domain.Source) error {
	return nil
}

func (r configurableSourceRepository) List(ctx context.Context, informationID uuid.UUID) ([]*domain.Source, error) {
	return r.sources, nil
}

func newTestResponseRouter(information *domain.Information, sources []*domain.Source) (*gin.Engine, *usecase.AccountUsecase) {
	gin.SetMode(gin.TestMode)

	uc := usecase.NewResponseUsecase(
		noopDisplayInformationRepository{information: information},
		configurableSourceRepository{sources: sources},
		noopOptionRepository{},
		noopRecipientRepository{},
		&noopResponseRepository{},
		noopTransactionManager{},
	)
	accountUsecase := newTestAccountUsecase()

	verifier := testKeySet.NewVerifier(context.Background())
	authMiddleware := supabase.Middleware(verifier)
	optionalAuthMiddleware := supabase.OptionalMiddleware(verifier)

	r := gin.New()
	c := NewResponseController(uc, accountUsecase)
	// 本番のrouter.goと同様、SubmitResponseのみ認証を必須としない
	// (PUBLIC+ANONYMOUSなInformationへの匿名回答を許可するため)。
	r.POST("/informations/:id/responses", optionalAuthMiddleware, c.SubmitResponse)
	informations := r.Group("/informations", authMiddleware)
	informations.GET("/:id/responses", c.ListResponses)
	return r, accountUsecase
}

type noopResponseRepository struct {
	created []*domain.Response
}

func (r *noopResponseRepository) Create(ctx context.Context, response *domain.Response) error {
	r.created = append(r.created, response)
	return nil
}

func (r *noopResponseRepository) GetByInformationAndUser(ctx context.Context, informationID, userID uuid.UUID) (*domain.Response, error) {
	return nil, domain.ErrNotFound
}

func (r *noopResponseRepository) ListByInformationID(ctx context.Context, informationID uuid.UUID) ([]*domain.Response, error) {
	return r.created, nil
}

func TestResponseController_SubmitResponse_UnauthenticatedIsNotBlockedByMiddleware(t *testing.T) {
	// SubmitResponseは認証を必須としないため、未認証でもミドルウェアで
	// 弾かれることはない(この情報が存在しないため、usecase層でNotFoundになる)。
	r, _ := newTestResponseRouter(nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/informations/"+uuid.New().String()+"/responses", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code == http.StatusUnauthorized {
		t.Fatalf("status = %d, want not %d (SubmitResponse does not require authentication)", w.Code, http.StatusUnauthorized)
	}
}

func TestResponseController_SubmitResponse_AnonymousSuccessForPublicAnonymous(t *testing.T) {
	information, err := domain.NewInformation(uuid.New(), uuid.New(), "旅行のお知らせ", domain.InformationAccessTypePublic, domain.InformationResponsePolicyAnonymous)
	if err != nil {
		t.Fatalf("NewInformation() error = %v", err)
	}

	text := domain.SourceInteractionTypeText
	source, err := domain.NewSource(information.ID(), domain.SourceTypeFact, "comment", "コメントをお願いします", domain.SourceStatusConfirmed, &text)
	if err != nil {
		t.Fatalf("NewSource() error = %v", err)
	}

	r, _ := newTestResponseRouter(information, []*domain.Source{source})

	body, _ := json.Marshal(map[string]any{
		"items": []map[string]any{{"source_id": source.ID().String(), "value": "匿名で回答します"}},
	})
	req := httptest.NewRequest(http.MethodPost, "/informations/"+information.ID().String()+"/responses", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// Authorizationヘッダを付けない(未ログイン)。
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusCreated, w.Body.String())
	}

	var view usecase.ResponseView
	if err := json.Unmarshal(w.Body.Bytes(), &view); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if view.UserID != nil {
		t.Errorf("UserID = %v, want nil (anonymous)", view.UserID)
	}
}

func TestResponseController_SubmitResponse_AnonymousForbiddenForPublicAuthenticated(t *testing.T) {
	information, err := domain.NewInformation(uuid.New(), uuid.New(), "旅行のお知らせ", domain.InformationAccessTypePublic, domain.InformationResponsePolicyAuthenticated)
	if err != nil {
		t.Fatalf("NewInformation() error = %v", err)
	}

	text := domain.SourceInteractionTypeText
	source, err := domain.NewSource(information.ID(), domain.SourceTypeFact, "comment", "コメントをお願いします", domain.SourceStatusConfirmed, &text)
	if err != nil {
		t.Fatalf("NewSource() error = %v", err)
	}

	r, _ := newTestResponseRouter(information, []*domain.Source{source})

	body, _ := json.Marshal(map[string]any{
		"items": []map[string]any{{"source_id": source.ID().String(), "value": "匿名で回答します"}},
	})
	req := httptest.NewRequest(http.MethodPost, "/informations/"+information.ID().String()+"/responses", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusForbidden, w.Body.String())
	}
}

func TestResponseController_SubmitResponse_BadRequest_NoItems(t *testing.T) {
	information, err := domain.NewInformation(uuid.New(), uuid.New(), "旅行のお知らせ", domain.InformationAccessTypePublic, domain.InformationResponsePolicyAnonymous)
	if err != nil {
		t.Fatalf("NewInformation() error = %v", err)
	}

	r, accountUsecase := newTestResponseRouter(information, nil)
	registerTestIdentity(t, accountUsecase, "recipient-1")

	req := httptest.NewRequest(http.MethodPost, "/informations/"+information.ID().String()+"/responses", bytes.NewReader([]byte(`{"items":[]}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+signTestToken(t, "recipient-1"))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusBadRequest, w.Body.String())
	}
}

func TestResponseController_SubmitResponse_RestrictedForbidden(t *testing.T) {
	information, err := domain.NewInformation(uuid.New(), uuid.New(), "旅行のお知らせ", domain.InformationAccessTypeRestricted, domain.InformationResponsePolicyAuthenticated)
	if err != nil {
		t.Fatalf("NewInformation() error = %v", err)
	}

	text := domain.SourceInteractionTypeText
	source, err := domain.NewSource(information.ID(), domain.SourceTypeFact, "comment", "コメントをお願いします", domain.SourceStatusConfirmed, &text)
	if err != nil {
		t.Fatalf("NewSource() error = %v", err)
	}

	r, accountUsecase := newTestResponseRouter(information, []*domain.Source{source})
	registerTestIdentity(t, accountUsecase, "not-a-recipient")

	body, _ := json.Marshal(map[string]any{
		"items": []map[string]any{{"source_id": source.ID().String(), "value": "よろしくお願いします"}},
	})
	req := httptest.NewRequest(http.MethodPost, "/informations/"+information.ID().String()+"/responses", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+signTestToken(t, "not-a-recipient"))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusForbidden, w.Body.String())
	}
}

func TestResponseController_SubmitResponse_Success(t *testing.T) {
	information, err := domain.NewInformation(uuid.New(), uuid.New(), "旅行のお知らせ", domain.InformationAccessTypePublic, domain.InformationResponsePolicyAnonymous)
	if err != nil {
		t.Fatalf("NewInformation() error = %v", err)
	}

	text := domain.SourceInteractionTypeText
	source, err := domain.NewSource(information.ID(), domain.SourceTypeFact, "comment", "コメントをお願いします", domain.SourceStatusConfirmed, &text)
	if err != nil {
		t.Fatalf("NewSource() error = %v", err)
	}

	r, accountUsecase := newTestResponseRouter(information, []*domain.Source{source})
	registerTestIdentity(t, accountUsecase, "recipient-1")

	body, _ := json.Marshal(map[string]any{
		"items": []map[string]any{{"source_id": source.ID().String(), "value": "よろしくお願いします"}},
	})
	req := httptest.NewRequest(http.MethodPost, "/informations/"+information.ID().String()+"/responses", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+signTestToken(t, "recipient-1"))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusCreated, w.Body.String())
	}

	var view usecase.ResponseView
	if err := json.Unmarshal(w.Body.Bytes(), &view); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(view.Items) != 1 || view.Items[0].SourceID != source.ID().String() {
		t.Errorf("view = %+v, unexpected", view)
	}
}

func TestResponseController_ListResponses_ForbiddenForNonOwner(t *testing.T) {
	information, err := domain.NewInformation(uuid.New(), uuid.New(), "旅行のお知らせ", domain.InformationAccessTypePublic, domain.InformationResponsePolicyAnonymous)
	if err != nil {
		t.Fatalf("NewInformation() error = %v", err)
	}

	r, accountUsecase := newTestResponseRouter(information, nil)
	registerTestIdentity(t, accountUsecase, "not-the-owner")

	req := httptest.NewRequest(http.MethodGet, "/informations/"+information.ID().String()+"/responses", nil)
	req.Header.Set("Authorization", "Bearer "+signTestToken(t, "not-the-owner"))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusForbidden, w.Body.String())
	}
}
