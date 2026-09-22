package usecase

import (
	"context"
	"errors"
	"testing"

	"knot-api/internal/domain"

	"uuid"
)

type fakeAccountRepository struct {
	accounts map[uuid.UUID]*domain.Account
}

func newFakeAccountRepository() *fakeAccountRepository {
	return &fakeAccountRepository{accounts: map[uuid.UUID]*domain.Account{}}
}

func (r *fakeAccountRepository) Create(ctx context.Context, account *domain.Account) error {
	r.accounts[account.ID()] = account
	return nil
}

func (r *fakeAccountRepository) Get(ctx context.Context, id uuid.UUID) (*domain.Account, error) {
	a, ok := r.accounts[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return a, nil
}

func (r *fakeAccountRepository) GetByProviderID(ctx context.Context, providerID string) (*domain.Account, error) {
	for _, a := range r.accounts {
		if a.ProviderID() == providerID {
			return a, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *fakeAccountRepository) Update(ctx context.Context, account *domain.Account) error {
	if _, ok := r.accounts[account.ID()]; !ok {
		return domain.ErrNotFound
	}
	r.accounts[account.ID()] = account
	return nil
}

func (r *fakeAccountRepository) ListByStatus(ctx context.Context, status domain.AccountStatus) ([]*domain.Account, error) {
	var out []*domain.Account
	for _, a := range r.accounts {
		if a.Status() == status {
			out = append(out, a)
		}
	}
	return out, nil
}

type fakeUserRepository struct {
	users map[uuid.UUID]*domain.User
}

func newFakeUserRepository() *fakeUserRepository {
	return &fakeUserRepository{users: map[uuid.UUID]*domain.User{}}
}

func (r *fakeUserRepository) Create(ctx context.Context, user *domain.User) error {
	r.users[user.ID()] = user
	return nil
}

func (r *fakeUserRepository) Get(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	u, ok := r.users[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return u, nil
}

func (r *fakeUserRepository) Update(ctx context.Context, user *domain.User) error {
	if _, ok := r.users[user.ID()]; !ok {
		return domain.ErrNotFound
	}
	r.users[user.ID()] = user
	return nil
}

type fakeMembershipRepository struct {
	memberships map[uuid.UUID]*domain.Membership
}

func newFakeMembershipRepository() *fakeMembershipRepository {
	return &fakeMembershipRepository{memberships: map[uuid.UUID]*domain.Membership{}}
}

func (r *fakeMembershipRepository) Create(ctx context.Context, membership *domain.Membership) error {
	r.memberships[membership.ID()] = membership
	return nil
}

func (r *fakeMembershipRepository) Get(ctx context.Context, id uuid.UUID) (*domain.Membership, error) {
	m, ok := r.memberships[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return m, nil
}

func (r *fakeMembershipRepository) GetByAccountAndUser(ctx context.Context, accountID, userID uuid.UUID) (*domain.Membership, error) {
	for _, m := range r.memberships {
		if m.AccountID() == accountID && m.UserID() == userID {
			return m, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *fakeMembershipRepository) GetOwnerByAccountID(ctx context.Context, accountID uuid.UUID) (*domain.Membership, error) {
	for _, m := range r.memberships {
		if m.AccountID() == accountID && m.IsOwner() {
			return m, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *fakeMembershipRepository) ListByAccountID(ctx context.Context, accountID uuid.UUID) ([]*domain.Membership, error) {
	var out []*domain.Membership
	for _, m := range r.memberships {
		if m.AccountID() == accountID {
			out = append(out, m)
		}
	}
	return out, nil
}

func (r *fakeMembershipRepository) Update(ctx context.Context, membership *domain.Membership) error {
	if _, ok := r.memberships[membership.ID()]; !ok {
		return domain.ErrNotFound
	}
	r.memberships[membership.ID()] = membership
	return nil
}

type fakeMembershipEventRepository struct {
	created []*domain.MembershipEvent
}

func (r *fakeMembershipEventRepository) Create(ctx context.Context, event *domain.MembershipEvent) error {
	r.created = append(r.created, event)
	return nil
}

type fakeMembershipRemovalRequestRepository struct {
	requests map[uuid.UUID]*domain.MembershipRemovalRequest
}

func newFakeMembershipRemovalRequestRepository() *fakeMembershipRemovalRequestRepository {
	return &fakeMembershipRemovalRequestRepository{requests: map[uuid.UUID]*domain.MembershipRemovalRequest{}}
}

func (r *fakeMembershipRemovalRequestRepository) Create(ctx context.Context, request *domain.MembershipRemovalRequest) error {
	r.requests[request.ID()] = request
	return nil
}

func (r *fakeMembershipRemovalRequestRepository) Get(ctx context.Context, id uuid.UUID) (*domain.MembershipRemovalRequest, error) {
	req, ok := r.requests[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return req, nil
}

func (r *fakeMembershipRemovalRequestRepository) GetPendingByMembershipID(ctx context.Context, membershipID uuid.UUID) (*domain.MembershipRemovalRequest, error) {
	for _, req := range r.requests {
		if req.MembershipID() == membershipID && req.IsPending() {
			return req, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *fakeMembershipRemovalRequestRepository) ListPendingByAccountID(ctx context.Context, accountID uuid.UUID) ([]*domain.MembershipRemovalRequest, error) {
	return nil, nil
}

func (r *fakeMembershipRemovalRequestRepository) Update(ctx context.Context, request *domain.MembershipRemovalRequest) error {
	if _, ok := r.requests[request.ID()]; !ok {
		return domain.ErrNotFound
	}
	r.requests[request.ID()] = request
	return nil
}

type fakeAccountStatusEventRepository struct {
	created []*domain.AccountStatusEvent
}

func (r *fakeAccountStatusEventRepository) Create(ctx context.Context, event *domain.AccountStatusEvent) error {
	r.created = append(r.created, event)
	return nil
}

func newTestAccountUsecase() (*AccountUsecase, *fakeAccountRepository, *fakeUserRepository, *fakeMembershipRepository, *fakeAccountStatusEventRepository) {
	accountRepo := newFakeAccountRepository()
	userRepo := newFakeUserRepository()
	membershipRepo := newFakeMembershipRepository()
	membershipEventRepo := &fakeMembershipEventRepository{}
	eventRepo := &fakeAccountStatusEventRepository{}
	txManager := &fakeTransactionManager{}

	uc := NewAccountUsecase(accountRepo, userRepo, membershipRepo, membershipEventRepo, eventRepo, txManager)
	return uc, accountRepo, userRepo, membershipRepo, eventRepo
}

func registerTestAccount(t *testing.T, uc *AccountUsecase, providerID string) *AccountView {
	t.Helper()
	view, err := uc.Register(context.Background(), RegisterInput{
		ProviderID:  providerID,
		AccountType: string(domain.AccountTypePersonal),
		LastName:    "山田",
		FirstName:   "太郎",
		Language:    "ja",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	return view
}

func TestAccountUsecase_GetUserProfile(t *testing.T) {
	uc, _, _, _, _ := newTestAccountUsecase()
	view := registerTestAccount(t, uc, "provider-1")

	userID, err := uuid.Parse(view.UserID)
	if err != nil {
		t.Fatalf("uuid.Parse() error = %v", err)
	}

	profile, err := uc.GetUserProfile(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetUserProfile() error = %v", err)
	}
	if profile.ID != view.UserID || profile.LastName != "山田" || profile.FirstName != "太郎" {
		t.Errorf("profile = %+v, unexpected", profile)
	}
}

func TestAccountUsecase_GetUserProfile_NotFound(t *testing.T) {
	uc, _, _, _, _ := newTestAccountUsecase()

	if _, err := uc.GetUserProfile(context.Background(), uuid.New()); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("GetUserProfile() error = %v, want ErrNotFound", err)
	}
}

func TestAccountUsecase_Register(t *testing.T) {
	uc, accountRepo, userRepo, membershipRepo, eventRepo := newTestAccountUsecase()

	view := registerTestAccount(t, uc, "provider-1")

	if view.Role != string(domain.AccountRoleUser) {
		t.Errorf("Role = %q, want %q", view.Role, domain.AccountRoleUser)
	}
	if view.Status != string(domain.AccountStatusActive) {
		t.Errorf("Status = %q, want %q", view.Status, domain.AccountStatusActive)
	}
	if len(accountRepo.accounts) != 1 || len(userRepo.users) != 1 {
		t.Errorf("accounts = %d, users = %d, want 1 each", len(accountRepo.accounts), len(userRepo.users))
	}
	if len(membershipRepo.memberships) != 1 {
		t.Fatalf("len(memberships) = %d, want 1", len(membershipRepo.memberships))
	}
	for _, m := range membershipRepo.memberships {
		if !m.IsOwner() {
			t.Errorf("Role() = %v, want %v", m.Role(), domain.MembershipRoleOwner)
		}
		if !m.IsActive() {
			t.Error("IsActive() = false, want true")
		}
	}
	if len(eventRepo.created) != 1 || eventRepo.created[0].EventType() != domain.AccountStatusEventCreated {
		t.Fatalf("eventRepo.created = %+v, want 1 created event", eventRepo.created)
	}
	if view.Name != nil {
		t.Errorf("Name = %v, want nil for personal account", view.Name)
	}
}

func TestAccountUsecase_Register_Organization(t *testing.T) {
	uc, _, _, _, _ := newTestAccountUsecase()

	orgName := "Acme, Inc."
	view, err := uc.Register(context.Background(), RegisterInput{
		ProviderID:  "provider-org-1",
		AccountType: string(domain.AccountTypeOrganization),
		Name:        &orgName,
		LastName:    "山田",
		FirstName:   "太郎",
		Language:    "ja",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if view.Name == nil || *view.Name != orgName {
		t.Errorf("Name = %v, want %q", view.Name, orgName)
	}

	// personalなAccountにnameを指定するとエラーになる。
	badName := "not allowed"
	if _, err := uc.Register(context.Background(), RegisterInput{
		ProviderID:  "provider-bad-1",
		AccountType: string(domain.AccountTypePersonal),
		Name:        &badName,
		LastName:    "山田",
		FirstName:   "太郎",
		Language:    "ja",
	}); err == nil {
		t.Error("Register(personal, name) error = nil, want error")
	}

	// organizationなAccountにnameが無いとエラーになる。
	if _, err := uc.Register(context.Background(), RegisterInput{
		ProviderID:  "provider-bad-2",
		AccountType: string(domain.AccountTypeOrganization),
		LastName:    "山田",
		FirstName:   "太郎",
		Language:    "ja",
	}); err == nil {
		t.Error("Register(organization, nil name) error = nil, want error")
	}
}

func TestAccountUsecase_GetMe_And_UpdateMyProfile(t *testing.T) {
	uc, _, _, _, _ := newTestAccountUsecase()
	view := registerTestAccount(t, uc, "provider-1")
	accountID, _ := uuid.Parse(view.ID)

	got, err := uc.GetMe(context.Background(), accountID)
	if err != nil {
		t.Fatalf("GetMe() error = %v", err)
	}
	if got.LastName != "山田" {
		t.Errorf("LastName = %q, want %q", got.LastName, "山田")
	}

	updated, err := uc.UpdateMyProfile(context.Background(), accountID, UpdateProfileInput{
		LastName:  "鈴木",
		FirstName: "花子",
		Language:  "en",
	})
	if err != nil {
		t.Fatalf("UpdateMyProfile() error = %v", err)
	}
	if updated.LastName != "鈴木" || updated.FirstName != "花子" || updated.Language != "en" {
		t.Errorf("updated = %+v, unexpected", updated)
	}
}

func TestAccountUsecase_ListByStatus_RequiresAdmin(t *testing.T) {
	uc, accountRepo, _, _, _ := newTestAccountUsecase()
	nonAdmin := registerTestAccount(t, uc, "provider-user")

	nonAdminID, _ := uuid.Parse(nonAdmin.ID)
	if _, err := uc.ListByStatus(context.Background(), nonAdminID, string(domain.AccountStatusActive)); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("ListByStatus() error = %v, want ErrForbidden", err)
	}

	if _, err := uc.ListByStatus(context.Background(), nonAdminID, string(domain.AccountStatusWithdrawn)); err == nil {
		t.Error("ListByStatus(withdrawn) error = nil, want error")
	}

	// admin に昇格させて再試行する。
	adminAccount := accountRepo.accounts[nonAdminID]
	promoted := domain.ReconstructAccount(adminAccount.ID(), adminAccount.ProviderID(), adminAccount.AccountType(), adminAccount.Name(), domain.AccountRoleAdmin, adminAccount.Status(), adminAccount.CreatedAt(), adminAccount.UpdatedAt())
	accountRepo.accounts[nonAdminID] = promoted

	views, err := uc.ListByStatus(context.Background(), nonAdminID, string(domain.AccountStatusActive))
	if err != nil {
		t.Fatalf("ListByStatus() error = %v", err)
	}
	if len(views) != 1 {
		t.Fatalf("len(views) = %d, want 1", len(views))
	}
}

func TestAccountUsecase_UpdateStatus(t *testing.T) {
	uc, accountRepo, _, _, eventRepo := newTestAccountUsecase()

	adminView := registerTestAccount(t, uc, "provider-admin")
	targetView := registerTestAccount(t, uc, "provider-target")

	adminID, _ := uuid.Parse(adminView.ID)
	targetID, _ := uuid.Parse(targetView.ID)

	admin := accountRepo.accounts[adminID]
	accountRepo.accounts[adminID] = domain.ReconstructAccount(admin.ID(), admin.ProviderID(), admin.AccountType(), admin.Name(), domain.AccountRoleAdmin, admin.Status(), admin.CreatedAt(), admin.UpdatedAt())

	updated, err := uc.UpdateStatus(context.Background(), adminID, targetID, AccountStatusActionFreeze)
	if err != nil {
		t.Fatalf("UpdateStatus(freeze) error = %v", err)
	}
	if updated.Status != string(domain.AccountStatusFrozen) {
		t.Errorf("Status = %q, want %q", updated.Status, domain.AccountStatusFrozen)
	}
	if len(eventRepo.created) != 3 || eventRepo.created[2].EventType() != domain.AccountStatusEventFrozen {
		t.Fatalf("eventRepo.created = %+v, unexpected", eventRepo.created)
	}

	// 非adminからの操作は拒否される。
	if _, err := uc.UpdateStatus(context.Background(), targetID, targetID, AccountStatusActionUnfreeze); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("UpdateStatus() by non-admin error = %v, want ErrForbidden", err)
	}

	// adminによるunfreeze。
	updated, err = uc.UpdateStatus(context.Background(), adminID, targetID, AccountStatusActionUnfreeze)
	if err != nil {
		t.Fatalf("UpdateStatus(unfreeze) error = %v", err)
	}
	if updated.Status != string(domain.AccountStatusActive) {
		t.Errorf("Status = %q, want %q", updated.Status, domain.AccountStatusActive)
	}
	if len(eventRepo.created) != 4 || eventRepo.created[3].EventType() != domain.AccountStatusEventUnfrozen {
		t.Fatalf("eventRepo.created = %+v, unexpected", eventRepo.created)
	}
}

func TestAccountUsecase_Delete(t *testing.T) {
	uc, accountRepo, _, _, eventRepo := newTestAccountUsecase()
	view := registerTestAccount(t, uc, "provider-1")
	accountID, _ := uuid.Parse(view.ID)

	if err := uc.Delete(context.Background(), accountID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if accountRepo.accounts[accountID].Status() != domain.AccountStatusWithdrawn {
		t.Errorf("Status = %v, want %v", accountRepo.accounts[accountID].Status(), domain.AccountStatusWithdrawn)
	}
	if len(eventRepo.created) != 2 || eventRepo.created[1].EventType() != domain.AccountStatusEventWithdrawn {
		t.Fatalf("eventRepo.created = %+v, unexpected", eventRepo.created)
	}

	if err := uc.Delete(context.Background(), accountID); err == nil {
		t.Error("Delete() on already-withdrawn account error = nil, want error")
	}
}
