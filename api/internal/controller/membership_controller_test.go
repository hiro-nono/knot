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

// membershipTestEnv はAccountUsecaseとMembershipUsecaseが同じ永続化状態を
// 共有するテスト環境を表す。
type membershipTestEnv struct {
	router          *gin.Engine
	accountUsecase  *usecase.AccountUsecase
	removalRequests *inMemoryMembershipRemovalRequestRepository
}

func newMembershipTestEnv() *membershipTestEnv {
	gin.SetMode(gin.TestMode)

	accountRepo := newInMemoryAccountRepository()
	userRepo := newInMemoryUserRepository()
	membershipRepo := newInMemoryMembershipRepository()
	removalRequestRepo := newInMemoryMembershipRemovalRequestRepository()

	accountUsecase := usecase.NewAccountUsecase(
		accountRepo,
		userRepo,
		membershipRepo,
		noopMembershipEventRepository{},
		noopAccountStatusEventRepository{},
		noopTransactionManager{},
	)
	membershipUsecase := usecase.NewMembershipUsecase(
		userRepo,
		membershipRepo,
		noopMembershipEventRepository{},
		removalRequestRepo,
		noopTransactionManager{},
	)

	authMiddleware := supabase.Middleware(testKeySet.NewVerifier(context.Background()))

	r := gin.New()
	c := NewMembershipController(membershipUsecase, accountUsecase)
	accounts := r.Group("/accounts", authMiddleware)
	accounts.POST("/:id/members", c.AddMember)
	accounts.GET("/:id/members", c.ListMembers)
	accounts.PATCH("/:id/members/:user_id", c.GrantAdmin)
	accounts.DELETE("/:id/members/:user_id", c.RemoveMember)
	accounts.GET("/:id/removal-requests", c.ListRemovalRequests)
	accounts.POST("/:id/removal-requests/:request_id/approve", c.ApproveRemovalRequest)
	accounts.POST("/:id/removal-requests/:request_id/reject", c.RejectRemovalRequest)

	return &membershipTestEnv{router: r, accountUsecase: accountUsecase, removalRequests: removalRequestRepo}
}

type inMemoryMembershipRemovalRequestRepository struct {
	requests map[uuid.UUID]*domain.MembershipRemovalRequest
}

func newInMemoryMembershipRemovalRequestRepository() *inMemoryMembershipRemovalRequestRepository {
	return &inMemoryMembershipRemovalRequestRepository{requests: map[uuid.UUID]*domain.MembershipRemovalRequest{}}
}

func (r *inMemoryMembershipRemovalRequestRepository) Create(ctx context.Context, request *domain.MembershipRemovalRequest) error {
	r.requests[request.ID()] = request
	return nil
}

func (r *inMemoryMembershipRemovalRequestRepository) Get(ctx context.Context, id uuid.UUID) (*domain.MembershipRemovalRequest, error) {
	req, ok := r.requests[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return req, nil
}

func (r *inMemoryMembershipRemovalRequestRepository) GetPendingByMembershipID(ctx context.Context, membershipID uuid.UUID) (*domain.MembershipRemovalRequest, error) {
	for _, req := range r.requests {
		if req.MembershipID() == membershipID && req.IsPending() {
			return req, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *inMemoryMembershipRemovalRequestRepository) ListPendingByAccountID(ctx context.Context, accountID uuid.UUID) ([]*domain.MembershipRemovalRequest, error) {
	var out []*domain.MembershipRemovalRequest
	for _, req := range r.requests {
		if req.IsPending() {
			out = append(out, req)
		}
	}
	return out, nil
}

func (r *inMemoryMembershipRemovalRequestRepository) Update(ctx context.Context, request *domain.MembershipRemovalRequest) error {
	if _, ok := r.requests[request.ID()]; !ok {
		return domain.ErrNotFound
	}
	r.requests[request.ID()] = request
	return nil
}

func TestMembershipController_AddMember_Success(t *testing.T) {
	env := newMembershipTestEnv()
	ownerAccountID, _ := registerTestIdentity(t, env.accountUsecase, "owner-1")
	_, newUserID := registerTestIdentity(t, env.accountUsecase, "new-user-1")

	body, _ := json.Marshal(map[string]any{"user_id": newUserID.String()})
	req := httptest.NewRequest(http.MethodPost, "/accounts/"+ownerAccountID.String()+"/members", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+signTestToken(t, "owner-1"))
	w := httptest.NewRecorder()

	env.router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusCreated, w.Body.String())
	}

	var view usecase.MembershipView
	if err := json.Unmarshal(w.Body.Bytes(), &view); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if view.Role != "member" {
		t.Errorf("Role = %q, want %q", view.Role, "member")
	}
}

func TestMembershipController_AddMember_ForbiddenForNonOwner(t *testing.T) {
	env := newMembershipTestEnv()
	ownerAccountID, _ := registerTestIdentity(t, env.accountUsecase, "owner-1")
	_, otherUserID := registerTestIdentity(t, env.accountUsecase, "not-the-owner")
	_, newUserID := registerTestIdentity(t, env.accountUsecase, "new-user-1")
	_ = otherUserID

	body, _ := json.Marshal(map[string]any{"user_id": newUserID.String()})
	req := httptest.NewRequest(http.MethodPost, "/accounts/"+ownerAccountID.String()+"/members", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+signTestToken(t, "not-the-owner"))
	w := httptest.NewRecorder()

	env.router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusForbidden, w.Body.String())
	}
}

func TestMembershipController_RemoveMember_ByAdminIsPending(t *testing.T) {
	env := newMembershipTestEnv()
	ownerAccountID, _ := registerTestIdentity(t, env.accountUsecase, "owner-1")
	_, adminUserID := registerTestIdentity(t, env.accountUsecase, "admin-1")
	_, memberUserID := registerTestIdentity(t, env.accountUsecase, "member-1")

	addMember(t, env, ownerAccountID, "owner-1", adminUserID)
	addMember(t, env, ownerAccountID, "owner-1", memberUserID)
	grantAdmin(t, env, ownerAccountID, "owner-1", adminUserID)

	req := httptest.NewRequest(http.MethodDelete, "/accounts/"+ownerAccountID.String()+"/members/"+memberUserID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+signTestToken(t, "admin-1"))
	w := httptest.NewRecorder()

	env.router.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusAccepted, w.Body.String())
	}
}

func addMember(t *testing.T, env *membershipTestEnv, accountID uuid.UUID, ownerProviderID string, userID uuid.UUID) {
	t.Helper()
	body, _ := json.Marshal(map[string]any{"user_id": userID.String()})
	req := httptest.NewRequest(http.MethodPost, "/accounts/"+accountID.String()+"/members", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+signTestToken(t, ownerProviderID))
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("addMember: status = %d, body = %s", w.Code, w.Body.String())
	}
}

func grantAdmin(t *testing.T, env *membershipTestEnv, accountID uuid.UUID, ownerProviderID string, userID uuid.UUID) {
	t.Helper()
	body, _ := json.Marshal(map[string]any{"role": "admin"})
	req := httptest.NewRequest(http.MethodPatch, "/accounts/"+accountID.String()+"/members/"+userID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+signTestToken(t, ownerProviderID))
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("grantAdmin: status = %d, body = %s", w.Code, w.Body.String())
	}
}
