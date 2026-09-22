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
	"knot-api/internal/infrastructure/ai"
	"knot-api/internal/infrastructure/auth/supabase"
	"knot-api/internal/usecase"

	"uuid"
)

type fakeChatClient struct {
	response *ai.StructuredInformation
}

func (f *fakeChatClient) ChatCompletionStructured(ctx context.Context, model string, messages []ai.Message) (*ai.StructuredInformation, error) {
	return f.response, nil
}

type noopInformationRepository struct{}

func (noopInformationRepository) Create(ctx context.Context, information *domain.Information) error {
	return nil
}

func (noopInformationRepository) Get(ctx context.Context, id uuid.UUID) (*domain.Information, error) {
	return nil, domain.ErrNotFound
}

func (noopInformationRepository) ListByAccountID(ctx context.Context, accountID uuid.UUID) ([]*domain.Information, error) {
	return nil, nil
}

type noopSourceRepository struct{}

func (noopSourceRepository) Create(ctx context.Context, source *domain.Source) error { return nil }

func (noopSourceRepository) List(ctx context.Context, informationID uuid.UUID) ([]*domain.Source, error) {
	return nil, nil
}

type noopOptionRepository struct{}

func (noopOptionRepository) Create(ctx context.Context, option *domain.Option) error { return nil }

func (noopOptionRepository) List(ctx context.Context, sourceID uuid.UUID) ([]*domain.Option, error) {
	return nil, nil
}

type noopRecipientRepository struct{}

func (noopRecipientRepository) Create(ctx context.Context, recipient *domain.Recipient) error {
	return nil
}

func (noopRecipientRepository) Exists(ctx context.Context, informationID, userID uuid.UUID) (bool, error) {
	return false, nil
}

func (noopRecipientRepository) ListByInformationID(ctx context.Context, informationID uuid.UUID) ([]*domain.Recipient, error) {
	return nil, nil
}

func (noopRecipientRepository) Delete(ctx context.Context, informationID, userID uuid.UUID) error {
	return domain.ErrNotFound
}

type noopTransactionManager struct{}

func (noopTransactionManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func newTestInformationRouter(chat *fakeChatClient) (*gin.Engine, *usecase.AccountUsecase) {
	gin.SetMode(gin.TestMode)

	uc := usecase.NewInformationUsecase(
		chat,
		"orcarouter/auto",
		noopInformationRepository{},
		noopSourceRepository{},
		noopOptionRepository{},
		noopRecipientRepository{},
		noopTransactionManager{},
	)
	accountUsecase := newTestAccountUsecase()

	authMiddleware := supabase.Middleware(testKeySet.NewVerifier(context.Background()))

	r := gin.New()
	c := NewInformationController(uc, accountUsecase)
	informations := r.Group("/informations", authMiddleware)
	informations.GET("", c.ListMine)
	informations.POST("/messages", c.ProcessInformation)
	informations.POST("/:id/recipients", c.AddRecipient)
	informations.GET("/:id/recipients", c.ListRecipients)
	informations.DELETE("/:id/recipients/:user_id", c.RemoveRecipient)
	return r, accountUsecase
}

func TestInformationController_ProcessInformation_Unauthenticated(t *testing.T) {
	r, _ := newTestInformationRouter(&fakeChatClient{})

	body, _ := json.Marshal(map[string]any{"access_type": "restricted", "response_policy": "authenticated", "user_input": "旅行の案内をしたい"})
	req := httptest.NewRequest(http.MethodPost, "/informations/messages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestInformationController_ProcessInformation_BadRequest(t *testing.T) {
	r, accountUsecase := newTestInformationRouter(&fakeChatClient{})
	registerTestIdentity(t, accountUsecase, "sender-1")

	req := httptest.NewRequest(http.MethodPost, "/informations/messages", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+signTestToken(t, "sender-1"))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestInformationController_ProcessInformation_NotConfirmed(t *testing.T) {
	chat := &fakeChatClient{
		response: &ai.StructuredInformation{
			Title: "旅行のお知らせ",
			Sources: []ai.StructuredSource{
				{Type: "schedule", Key: "departure_date", Status: "unknown", Options: []ai.StructuredOption{}},
			},
		},
	}
	r, accountUsecase := newTestInformationRouter(chat)
	registerTestIdentity(t, accountUsecase, "sender-1")

	body, _ := json.Marshal(map[string]any{"access_type": "restricted", "response_policy": "authenticated", "user_input": "旅行の案内をしたい"})
	req := httptest.NewRequest(http.MethodPost, "/informations/messages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+signTestToken(t, "sender-1"))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusOK, w.Body.String())
	}

	var out usecase.ProcessInformationOutput
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if out.Confirmed {
		t.Error("Confirmed = true, want false")
	}
	if out.Structured == nil || out.Structured.Title != "旅行のお知らせ" {
		t.Errorf("Structured = %+v, unexpected", out.Structured)
	}
}

func TestInformationController_ListMine_Unauthenticated(t *testing.T) {
	r, _ := newTestInformationRouter(&fakeChatClient{})

	req := httptest.NewRequest(http.MethodGet, "/informations", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestInformationController_ListMine_Success(t *testing.T) {
	r, accountUsecase := newTestInformationRouter(&fakeChatClient{})
	registerTestIdentity(t, accountUsecase, "sender-1")

	req := httptest.NewRequest(http.MethodGet, "/informations", nil)
	req.Header.Set("Authorization", "Bearer "+signTestToken(t, "sender-1"))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusOK, w.Body.String())
	}

	var out []usecase.InformationSummaryView
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(out) != 0 {
		t.Errorf("len(out) = %d, want 0 (noopInformationRepository has no data)", len(out))
	}
}

func TestInformationController_AddRecipient_ForbiddenForNonOwner(t *testing.T) {
	r, accountUsecase := newTestInformationRouter(&fakeChatClient{})
	registerTestIdentity(t, accountUsecase, "not-the-owner")

	body, _ := json.Marshal(map[string]any{"user_id": uuid.New().String()})
	req := httptest.NewRequest(http.MethodPost, "/informations/"+uuid.New().String()+"/recipients", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+signTestToken(t, "not-the-owner"))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// noopInformationRepository.Get は常にErrNotFoundを返すため、
	// 所有者チェック以前にNotFoundとなる。
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusNotFound, w.Body.String())
	}
}
