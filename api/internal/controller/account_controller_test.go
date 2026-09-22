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
	"knot-api/internal/infrastructure/auth/supabase/supabasetest"
	"knot-api/internal/usecase"

	"uuid"
)

type inMemoryAccountRepository struct {
	accounts map[uuid.UUID]*domain.Account
}

func newInMemoryAccountRepository() *inMemoryAccountRepository {
	return &inMemoryAccountRepository{accounts: map[uuid.UUID]*domain.Account{}}
}

func (r *inMemoryAccountRepository) Create(ctx context.Context, account *domain.Account) error {
	r.accounts[account.ID()] = account
	return nil
}

func (r *inMemoryAccountRepository) Get(ctx context.Context, id uuid.UUID) (*domain.Account, error) {
	a, ok := r.accounts[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return a, nil
}

func (r *inMemoryAccountRepository) GetByProviderID(ctx context.Context, providerID string) (*domain.Account, error) {
	for _, a := range r.accounts {
		if a.ProviderID() == providerID {
			return a, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *inMemoryAccountRepository) Update(ctx context.Context, account *domain.Account) error {
	r.accounts[account.ID()] = account
	return nil
}

func (r *inMemoryAccountRepository) ListByStatus(ctx context.Context, status domain.AccountStatus) ([]*domain.Account, error) {
	var out []*domain.Account
	for _, a := range r.accounts {
		if a.Status() == status {
			out = append(out, a)
		}
	}
	return out, nil
}

type inMemoryUserRepository struct {
	users map[uuid.UUID]*domain.User
}

func newInMemoryUserRepository() *inMemoryUserRepository {
	return &inMemoryUserRepository{users: map[uuid.UUID]*domain.User{}}
}

func (r *inMemoryUserRepository) Create(ctx context.Context, user *domain.User) error {
	r.users[user.ID()] = user
	return nil
}

func (r *inMemoryUserRepository) Get(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	u, ok := r.users[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return u, nil
}

func (r *inMemoryUserRepository) Update(ctx context.Context, user *domain.User) error {
	r.users[user.ID()] = user
	return nil
}

type inMemoryMembershipRepository struct {
	memberships map[uuid.UUID]*domain.Membership
}

func newInMemoryMembershipRepository() *inMemoryMembershipRepository {
	return &inMemoryMembershipRepository{memberships: map[uuid.UUID]*domain.Membership{}}
}

func (r *inMemoryMembershipRepository) Create(ctx context.Context, membership *domain.Membership) error {
	r.memberships[membership.ID()] = membership
	return nil
}

func (r *inMemoryMembershipRepository) Get(ctx context.Context, id uuid.UUID) (*domain.Membership, error) {
	m, ok := r.memberships[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return m, nil
}

func (r *inMemoryMembershipRepository) GetByAccountAndUser(ctx context.Context, accountID, userID uuid.UUID) (*domain.Membership, error) {
	for _, m := range r.memberships {
		if m.AccountID() == accountID && m.UserID() == userID {
			return m, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *inMemoryMembershipRepository) GetOwnerByAccountID(ctx context.Context, accountID uuid.UUID) (*domain.Membership, error) {
	for _, m := range r.memberships {
		if m.AccountID() == accountID && m.IsOwner() {
			return m, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *inMemoryMembershipRepository) ListByAccountID(ctx context.Context, accountID uuid.UUID) ([]*domain.Membership, error) {
	var out []*domain.Membership
	for _, m := range r.memberships {
		if m.AccountID() == accountID {
			out = append(out, m)
		}
	}
	return out, nil
}

func (r *inMemoryMembershipRepository) Update(ctx context.Context, membership *domain.Membership) error {
	r.memberships[membership.ID()] = membership
	return nil
}

type noopMembershipEventRepository struct{}

func (noopMembershipEventRepository) Create(ctx context.Context, event *domain.MembershipEvent) error {
	return nil
}

type noopAccountStatusEventRepository struct{}

func (noopAccountStatusEventRepository) Create(ctx context.Context, event *domain.AccountStatusEvent) error {
	return nil
}

// testKeySet はcontrollerパッケージの全テスト(account, information, display,
// response, membership)で共用するSupabase JWKSのテストフィクスチャ。
var testKeySet = supabasetest.NewKeySet()

// signTestToken はテスト用に、Supabaseが発行するJWTと同じ形のトークンに署名する。
// controllerパッケージの他のテストからも共用する。
func signTestToken(t *testing.T, userID string) string {
	t.Helper()
	return testKeySet.SignToken(userID)
}

func newTestAccountRouter() (*gin.Engine, *usecase.AccountUsecase) {
	gin.SetMode(gin.TestMode)

	uc := usecase.NewAccountUsecase(
		newInMemoryAccountRepository(),
		newInMemoryUserRepository(),
		newInMemoryMembershipRepository(),
		noopMembershipEventRepository{},
		noopAccountStatusEventRepository{},
		noopTransactionManager{},
	)

	authMiddleware := supabase.Middleware(testKeySet.NewVerifier(context.Background()))

	r := gin.New()
	c := NewAccountController(uc)
	accounts := r.Group("/accounts", authMiddleware)
	accounts.POST("", c.Register)
	accounts.GET("", c.ListByStatus)
	accounts.GET("/me", c.GetMe)
	accounts.PATCH("/me", c.UpdateMyProfile)
	accounts.DELETE("/me", c.Delete)
	accounts.PATCH("/:id/status", c.UpdateStatus)
	r.GET("/users/:id", authMiddleware, c.GetUserProfile)
	return r, uc
}

func TestAccountController_Register_Success(t *testing.T) {
	r, _ := newTestAccountRouter()

	body, _ := json.Marshal(map[string]any{
		"account_type": "personal",
		"last_name":    "山田",
		"first_name":   "太郎",
		"language":     "ja",
	})
	req := httptest.NewRequest(http.MethodPost, "/accounts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+signTestToken(t, "supabase-1"))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusCreated, w.Body.String())
	}

	var view usecase.AccountView
	if err := json.Unmarshal(w.Body.Bytes(), &view); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if view.Role != "user" || view.Status != "active" {
		t.Errorf("view = %+v, unexpected", view)
	}
	if view.ProviderID != "supabase-1" {
		t.Errorf("ProviderID = %q, want %q (from verified token, not request body)", view.ProviderID, "supabase-1")
	}
}

func TestAccountController_Register_Organization_WithName(t *testing.T) {
	r, _ := newTestAccountRouter()

	body, _ := json.Marshal(map[string]any{
		"account_type": "organization",
		"name":         "Acme, Inc.",
		"last_name":    "山田",
		"first_name":   "太郎",
		"language":     "ja",
	})
	req := httptest.NewRequest(http.MethodPost, "/accounts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+signTestToken(t, "supabase-org-1"))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusCreated, w.Body.String())
	}

	var view usecase.AccountView
	if err := json.Unmarshal(w.Body.Bytes(), &view); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if view.Name == nil || *view.Name != "Acme, Inc." {
		t.Errorf("Name = %v, want %q", view.Name, "Acme, Inc.")
	}
}

func TestAccountController_Register_Organization_WithoutName_Fails(t *testing.T) {
	r, _ := newTestAccountRouter()

	body, _ := json.Marshal(map[string]any{
		"account_type": "organization",
		"last_name":    "山田",
		"first_name":   "太郎",
		"language":     "ja",
	})
	req := httptest.NewRequest(http.MethodPost, "/accounts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+signTestToken(t, "supabase-org-2"))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code == http.StatusCreated {
		t.Fatalf("status = %d, want not %d (organization requires a name)", w.Code, http.StatusCreated)
	}
}

func TestAccountController_Register_Unauthenticated(t *testing.T) {
	r, _ := newTestAccountRouter()

	body, _ := json.Marshal(map[string]any{"last_name": "山田", "first_name": "太郎", "language": "ja"})
	req := httptest.NewRequest(http.MethodPost, "/accounts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAccountController_Register_BadRequest(t *testing.T) {
	r, _ := newTestAccountRouter()

	req := httptest.NewRequest(http.MethodPost, "/accounts", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+signTestToken(t, "supabase-1"))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestAccountController_GetMe_NotRegistered(t *testing.T) {
	r, _ := newTestAccountRouter()

	req := httptest.NewRequest(http.MethodGet, "/accounts/me", nil)
	req.Header.Set("Authorization", "Bearer "+signTestToken(t, "supabase-never-registered"))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusNotFound, w.Body.String())
	}
}

func TestAccountController_GetUserProfile_Success(t *testing.T) {
	r, uc := newTestAccountRouter()

	view, err := uc.Register(context.Background(), usecase.RegisterInput{
		ProviderID: "p1", AccountType: "personal", LastName: "山田", FirstName: "太郎", Language: "ja",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/users/"+view.UserID, nil)
	req.Header.Set("Authorization", "Bearer "+signTestToken(t, "p1"))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusOK, w.Body.String())
	}

	var out usecase.UserProfileView
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if out.LastName != "山田" || out.FirstName != "太郎" {
		t.Errorf("out = %+v, unexpected", out)
	}
}

func TestAccountController_GetUserProfile_NotFound(t *testing.T) {
	r, _ := newTestAccountRouter()

	req := httptest.NewRequest(http.MethodGet, "/users/"+uuid.New().String(), nil)
	req.Header.Set("Authorization", "Bearer "+signTestToken(t, "someone"))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusNotFound, w.Body.String())
	}
}

func TestAccountController_ListByStatus_Forbidden(t *testing.T) {
	r, uc := newTestAccountRouter()

	if _, err := uc.Register(context.Background(), usecase.RegisterInput{
		ProviderID: "p1", AccountType: "personal", LastName: "山田", FirstName: "太郎", Language: "ja",
	}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/accounts?status=active", nil)
	req.Header.Set("Authorization", "Bearer "+signTestToken(t, "p1"))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusForbidden, w.Body.String())
	}
}
