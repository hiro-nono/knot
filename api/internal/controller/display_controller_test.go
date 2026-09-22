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

type noopDisplayInformationRepository struct {
	information *domain.Information
}

func (r noopDisplayInformationRepository) Create(ctx context.Context, information *domain.Information) error {
	return nil
}

func (r noopDisplayInformationRepository) Get(ctx context.Context, id uuid.UUID) (*domain.Information, error) {
	if r.information == nil {
		return nil, domain.ErrNotFound
	}
	return r.information, nil
}

func (r noopDisplayInformationRepository) ListByAccountID(ctx context.Context, accountID uuid.UUID) ([]*domain.Information, error) {
	return nil, nil
}

type noopDisplayPreferenceRepository struct{}

func (noopDisplayPreferenceRepository) Get(ctx context.Context, recipientID uuid.UUID) (*domain.Preference, error) {
	return nil, domain.ErrNotFound
}

func (noopDisplayPreferenceRepository) Save(ctx context.Context, preference *domain.Preference) error {
	return nil
}

type fakeDisplayChatClient struct {
	displayResponse    *ai.DisplayContent
	chatResponse       *ai.InformationChatResponse
	comparisonResponse *ai.ComparisonContent
}

func (f *fakeDisplayChatClient) ChatCompletionDisplay(ctx context.Context, model string, messages []ai.Message) (*ai.DisplayContent, error) {
	return f.displayResponse, nil
}

func (f *fakeDisplayChatClient) ChatCompletionInformationChat(ctx context.Context, model string, messages []ai.Message) (*ai.InformationChatResponse, error) {
	return f.chatResponse, nil
}

func (f *fakeDisplayChatClient) ChatCompletionComparison(ctx context.Context, model string, messages []ai.Message) (*ai.ComparisonContent, error) {
	return f.comparisonResponse, nil
}

func newTestDisplayRouter(chat *fakeDisplayChatClient, information *domain.Information) (*gin.Engine, *usecase.AccountUsecase) {
	gin.SetMode(gin.TestMode)

	uc := usecase.NewDisplayUsecase(
		chat,
		"opus",
		noopDisplayInformationRepository{information: information},
		noopSourceRepository{},
		noopOptionRepository{},
		noopDisplayPreferenceRepository{},
		noopRecipientRepository{},
	)
	accountUsecase := newTestAccountUsecase()

	authMiddleware := supabase.Middleware(testKeySet.NewVerifier(context.Background()))

	r := gin.New()
	c := NewDisplayController(uc, accountUsecase)
	informations := r.Group("/informations", authMiddleware)
	informations.POST("/:id/display", c.GenerateDisplay)
	informations.POST("/:id/display/comparisons", c.GenerateComparison)
	informations.POST("/:id/display/comparisons/apply", c.ApplyPreference)
	informations.POST("/:id/display/chat", c.Chat)
	return r, accountUsecase
}

func TestDisplayController_GenerateDisplay_Unauthenticated(t *testing.T) {
	r, _ := newTestDisplayRouter(&fakeDisplayChatClient{}, nil)

	req := httptest.NewRequest(http.MethodPost, "/informations/"+uuid.New().String()+"/display", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestDisplayController_GenerateDisplay_Success(t *testing.T) {
	information, err := domain.NewInformation(uuid.New(), uuid.New(), "旅行のお知らせ", domain.InformationAccessTypePublic, domain.InformationResponsePolicyAnonymous)
	if err != nil {
		t.Fatalf("NewInformation() error = %v", err)
	}

	chat := &fakeDisplayChatClient{displayResponse: &ai.DisplayContent{Title: "お知らせ", Body: "本文"}}
	r, accountUsecase := newTestDisplayRouter(chat, information)
	registerTestIdentity(t, accountUsecase, "recipient-1")

	req := httptest.NewRequest(http.MethodPost, "/informations/"+information.ID().String()+"/display", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+signTestToken(t, "recipient-1"))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusOK, w.Body.String())
	}

	var out usecase.GenerateDisplayOutput
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if out.Title != "お知らせ" || out.Body != "本文" {
		t.Errorf("out = %+v, unexpected", out)
	}
	if out.HasPreference {
		t.Error("HasPreference = true, want false (recipient has no preference yet)")
	}
}

func TestDisplayController_GenerateDisplay_RestrictedForbidden(t *testing.T) {
	information, err := domain.NewInformation(uuid.New(), uuid.New(), "旅行のお知らせ", domain.InformationAccessTypeRestricted, domain.InformationResponsePolicyAuthenticated)
	if err != nil {
		t.Fatalf("NewInformation() error = %v", err)
	}

	chat := &fakeDisplayChatClient{displayResponse: &ai.DisplayContent{Title: "お知らせ", Body: "本文"}}
	r, accountUsecase := newTestDisplayRouter(chat, information)
	registerTestIdentity(t, accountUsecase, "not-a-recipient")

	req := httptest.NewRequest(http.MethodPost, "/informations/"+information.ID().String()+"/display", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+signTestToken(t, "not-a-recipient"))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusForbidden, w.Body.String())
	}
}

type stubSourceRepository struct {
	sources []*domain.Source
}

func (r stubSourceRepository) Create(ctx context.Context, source *domain.Source) error { return nil }

func (r stubSourceRepository) List(ctx context.Context, informationID uuid.UUID) ([]*domain.Source, error) {
	return r.sources, nil
}

type stubOptionRepository struct {
	optionsBySource map[uuid.UUID][]*domain.Option
}

func (r stubOptionRepository) Create(ctx context.Context, option *domain.Option) error { return nil }

func (r stubOptionRepository) List(ctx context.Context, sourceID uuid.UUID) ([]*domain.Option, error) {
	return r.optionsBySource[sourceID], nil
}

func TestDisplayController_ListSources_Success(t *testing.T) {
	information, err := domain.NewInformation(uuid.New(), uuid.New(), "旅行のお知らせ", domain.InformationAccessTypePublic, domain.InformationResponsePolicyAnonymous)
	if err != nil {
		t.Fatalf("NewInformation() error = %v", err)
	}
	radio := domain.SourceInteractionTypeRadio
	source, err := domain.NewSource(information.ID(), domain.SourceTypeSchedule, "date", "2026-10-01", domain.SourceStatusConfirmed, &radio)
	if err != nil {
		t.Fatalf("NewSource() error = %v", err)
	}
	option, err := domain.NewOption(source.ID(), "参加する", 0)
	if err != nil {
		t.Fatalf("NewOption() error = %v", err)
	}

	gin.SetMode(gin.TestMode)
	uc := usecase.NewDisplayUsecase(
		&fakeDisplayChatClient{},
		"opus",
		noopDisplayInformationRepository{information: information},
		stubSourceRepository{sources: []*domain.Source{source}},
		stubOptionRepository{optionsBySource: map[uuid.UUID][]*domain.Option{source.ID(): {option}}},
		noopDisplayPreferenceRepository{},
		noopRecipientRepository{},
	)
	accountUsecase := newTestAccountUsecase()
	registerTestIdentity(t, accountUsecase, "recipient-1")

	// router.goと同様、ListSourcesはoptionalAuthMiddleware配下に置く
	// (PUBLIC+ANONYMOUSなInformationは未ログインでも一覧取得できるため)。
	optionalAuthMiddleware := supabase.OptionalMiddleware(testKeySet.NewVerifier(context.Background()))
	r := gin.New()
	c := NewDisplayController(uc, accountUsecase)
	r.GET("/informations/:id/sources", optionalAuthMiddleware, c.ListSources)

	req := httptest.NewRequest(http.MethodGet, "/informations/"+information.ID().String()+"/sources", nil)
	req.Header.Set("Authorization", "Bearer "+signTestToken(t, "recipient-1"))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusOK, w.Body.String())
	}

	var out []usecase.SourceView
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("len(out) = %d, want 1", len(out))
	}
	if out[0].Key != "date" || out[0].InteractionType == nil || *out[0].InteractionType != "radio" {
		t.Errorf("out[0] = %+v, unexpected", out[0])
	}
	if len(out[0].Options) != 1 || out[0].Options[0].Value != "参加する" {
		t.Errorf("out[0].Options = %+v, unexpected", out[0].Options)
	}
}

func TestDisplayController_ListSources_AnonymousSuccessForPublicAnonymous(t *testing.T) {
	information, err := domain.NewInformation(uuid.New(), uuid.New(), "旅行のお知らせ", domain.InformationAccessTypePublic, domain.InformationResponsePolicyAnonymous)
	if err != nil {
		t.Fatalf("NewInformation() error = %v", err)
	}
	text := domain.SourceInteractionTypeText
	source, err := domain.NewSource(information.ID(), domain.SourceTypeFact, "comment", "コメントをお願いします", domain.SourceStatusConfirmed, &text)
	if err != nil {
		t.Fatalf("NewSource() error = %v", err)
	}

	gin.SetMode(gin.TestMode)
	uc := usecase.NewDisplayUsecase(
		&fakeDisplayChatClient{},
		"opus",
		noopDisplayInformationRepository{information: information},
		stubSourceRepository{sources: []*domain.Source{source}},
		stubOptionRepository{},
		noopDisplayPreferenceRepository{},
		noopRecipientRepository{},
	)
	accountUsecase := newTestAccountUsecase()

	optionalAuthMiddleware := supabase.OptionalMiddleware(testKeySet.NewVerifier(context.Background()))
	r := gin.New()
	c := NewDisplayController(uc, accountUsecase)
	r.GET("/informations/:id/sources", optionalAuthMiddleware, c.ListSources)

	req := httptest.NewRequest(http.MethodGet, "/informations/"+information.ID().String()+"/sources", nil)
	// Authorizationヘッダを付けない(未ログイン)。
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestDisplayController_ListSources_AnonymousForbiddenForRestricted(t *testing.T) {
	information, err := domain.NewInformation(uuid.New(), uuid.New(), "旅行のお知らせ", domain.InformationAccessTypeRestricted, domain.InformationResponsePolicyAuthenticated)
	if err != nil {
		t.Fatalf("NewInformation() error = %v", err)
	}

	gin.SetMode(gin.TestMode)
	uc := usecase.NewDisplayUsecase(
		&fakeDisplayChatClient{},
		"opus",
		noopDisplayInformationRepository{information: information},
		stubSourceRepository{},
		stubOptionRepository{},
		noopDisplayPreferenceRepository{},
		noopRecipientRepository{},
	)
	accountUsecase := newTestAccountUsecase()

	optionalAuthMiddleware := supabase.OptionalMiddleware(testKeySet.NewVerifier(context.Background()))
	r := gin.New()
	c := NewDisplayController(uc, accountUsecase)
	r.GET("/informations/:id/sources", optionalAuthMiddleware, c.ListSources)

	req := httptest.NewRequest(http.MethodGet, "/informations/"+information.ID().String()+"/sources", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusForbidden, w.Body.String())
	}
}

func TestDisplayController_ApplyPreference_MissingValue(t *testing.T) {
	information, err := domain.NewInformation(uuid.New(), uuid.New(), "旅行のお知らせ", domain.InformationAccessTypePublic, domain.InformationResponsePolicyAnonymous)
	if err != nil {
		t.Fatalf("NewInformation() error = %v", err)
	}

	r, accountUsecase := newTestDisplayRouter(&fakeDisplayChatClient{}, information)
	registerTestIdentity(t, accountUsecase, "recipient-1")

	body, _ := json.Marshal(map[string]any{"key": "reading_level"})
	req := httptest.NewRequest(http.MethodPost, "/informations/"+information.ID().String()+"/display/comparisons/apply", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+signTestToken(t, "recipient-1"))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusBadRequest, w.Body.String())
	}
}

func TestDisplayController_GenerateComparison_Success(t *testing.T) {
	information, err := domain.NewInformation(uuid.New(), uuid.New(), "旅行のお知らせ", domain.InformationAccessTypePublic, domain.InformationResponsePolicyAnonymous)
	if err != nil {
		t.Fatalf("NewInformation() error = %v", err)
	}

	chat := &fakeDisplayChatClient{
		comparisonResponse: &ai.ComparisonContent{
			PatternA: ai.ComparisonPattern{Value: "easy", Title: "お知らせ", Body: "かんたんな本文"},
			PatternB: ai.ComparisonPattern{Value: "detailed", Title: "お知らせ", Body: "詳しい本文"},
		},
	}
	r, accountUsecase := newTestDisplayRouter(chat, information)
	registerTestIdentity(t, accountUsecase, "recipient-1")

	body, _ := json.Marshal(map[string]any{"key": "reading_level"})
	req := httptest.NewRequest(http.MethodPost, "/informations/"+information.ID().String()+"/display/comparisons", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+signTestToken(t, "recipient-1"))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusOK, w.Body.String())
	}

	var out usecase.GenerateComparisonOutput
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if out.PatternA.Value != "easy" || out.PatternB.Value != "detailed" {
		t.Errorf("out = %+v, unexpected", out)
	}
}

func TestDisplayController_Chat_Success(t *testing.T) {
	information, err := domain.NewInformation(uuid.New(), uuid.New(), "旅行のお知らせ", domain.InformationAccessTypePublic, domain.InformationResponsePolicyAnonymous)
	if err != nil {
		t.Fatalf("NewInformation() error = %v", err)
	}

	chat := &fakeDisplayChatClient{
		chatResponse: &ai.InformationChatResponse{Answer: "集合時間は10時です。"},
	}
	r, accountUsecase := newTestDisplayRouter(chat, information)
	registerTestIdentity(t, accountUsecase, "recipient-1")

	body, _ := json.Marshal(map[string]any{"user_input": "集合時間は何時ですか？"})
	req := httptest.NewRequest(http.MethodPost, "/informations/"+information.ID().String()+"/display/chat", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+signTestToken(t, "recipient-1"))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusOK, w.Body.String())
	}

	var out usecase.ChatOutput
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if out.Answer != "集合時間は10時です。" {
		t.Errorf("Answer = %q, want %q", out.Answer, "集合時間は10時です。")
	}
}
