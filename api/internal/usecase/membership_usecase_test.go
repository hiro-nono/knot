package usecase

import (
	"context"
	"errors"
	"testing"

	"knot-api/internal/domain"

	"uuid"
)

func newTestMembershipUsecase() (*MembershipUsecase, *fakeUserRepository, *fakeMembershipRepository, *fakeMembershipEventRepository, *fakeMembershipRemovalRequestRepository) {
	userRepo := newFakeUserRepository()
	membershipRepo := newFakeMembershipRepository()
	membershipEventRepo := &fakeMembershipEventRepository{}
	removalRequestRepo := newFakeMembershipRemovalRequestRepository()
	txManager := &fakeTransactionManager{}

	uc := NewMembershipUsecase(userRepo, membershipRepo, membershipEventRepo, removalRequestRepo, txManager)
	return uc, userRepo, membershipRepo, membershipEventRepo, removalRequestRepo
}

// seedOrganization はowner1名を持つOrganization(Account)を用意する。
func seedOrganization(t *testing.T, membershipRepo *fakeMembershipRepository) (accountID, ownerUserID uuid.UUID) {
	t.Helper()
	accountID = uuid.New()
	ownerUserID = uuid.New()

	owner, err := domain.NewMembership(accountID, ownerUserID, domain.MembershipRoleOwner)
	if err != nil {
		t.Fatalf("NewMembership() error = %v", err)
	}
	membershipRepo.memberships[owner.ID()] = owner

	return accountID, ownerUserID
}

func seedUser(t *testing.T, userRepo *fakeUserRepository) uuid.UUID {
	t.Helper()
	user, err := domain.NewUser("山田", "太郎", "ja")
	if err != nil {
		t.Fatalf("NewUser() error = %v", err)
	}
	userRepo.users[user.ID()] = user
	return user.ID()
}

func TestMembershipUsecase_AddUser_ByOwner(t *testing.T) {
	uc, userRepo, membershipRepo, eventRepo, _ := newTestMembershipUsecase()
	accountID, ownerID := seedOrganization(t, membershipRepo)
	newUserID := seedUser(t, userRepo)

	view, err := uc.AddUser(context.Background(), AddUserInput{
		ActingUserID:    ownerID,
		TargetAccountID: accountID,
		TargetUserID:    newUserID,
	})
	if err != nil {
		t.Fatalf("AddUser() error = %v", err)
	}
	if view.Role != string(domain.MembershipRoleMember) {
		t.Errorf("Role = %q, want %q", view.Role, domain.MembershipRoleMember)
	}
	if view.Status != string(domain.MembershipStatusActive) {
		t.Errorf("Status = %q, want %q", view.Status, domain.MembershipStatusActive)
	}
	if len(membershipRepo.memberships) != 2 {
		t.Fatalf("len(memberships) = %d, want 2", len(membershipRepo.memberships))
	}
	if len(eventRepo.created) != 1 || eventRepo.created[0].EventType() != domain.MembershipEventCreated {
		t.Fatalf("eventRepo.created = %+v, want 1 created event", eventRepo.created)
	}
}

func TestMembershipUsecase_AddUser_ForbiddenForNonOwner(t *testing.T) {
	uc, userRepo, membershipRepo, _, _ := newTestMembershipUsecase()
	accountID, ownerID := seedOrganization(t, membershipRepo)
	_ = ownerID
	newUserID := seedUser(t, userRepo)

	nonMemberID := uuid.New()
	if _, err := uc.AddUser(context.Background(), AddUserInput{
		ActingUserID:    nonMemberID,
		TargetAccountID: accountID,
		TargetUserID:    newUserID,
	}); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("AddUser() by non-member error = %v, want ErrForbidden", err)
	}
}

func TestMembershipUsecase_AddUser_ReactivatesRemovedMembership(t *testing.T) {
	uc, userRepo, membershipRepo, eventRepo, _ := newTestMembershipUsecase()
	accountID, ownerID := seedOrganization(t, membershipRepo)
	userID := seedUser(t, userRepo)

	existing, err := domain.NewMembership(accountID, userID, domain.MembershipRoleAdmin)
	if err != nil {
		t.Fatalf("NewMembership() error = %v", err)
	}
	if _, err := existing.Remove(); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
	membershipRepo.memberships[existing.ID()] = existing

	view, err := uc.AddUser(context.Background(), AddUserInput{
		ActingUserID:    ownerID,
		TargetAccountID: accountID,
		TargetUserID:    userID,
	})
	if err != nil {
		t.Fatalf("AddUser() error = %v", err)
	}
	if view.ID != existing.ID().String() {
		t.Errorf("ID = %q, want existing membership id %q (reuse, not new row)", view.ID, existing.ID())
	}
	if view.Role != string(domain.MembershipRoleAdmin) {
		t.Errorf("Role = %q, want %q (role preserved across reactivation)", view.Role, domain.MembershipRoleAdmin)
	}
	if view.Status != string(domain.MembershipStatusActive) {
		t.Errorf("Status = %q, want %q", view.Status, domain.MembershipStatusActive)
	}
	if len(membershipRepo.memberships) != 2 {
		t.Fatalf("len(memberships) = %d, want 2 (no new row created)", len(membershipRepo.memberships))
	}
	if len(eventRepo.created) != 1 || eventRepo.created[0].EventType() != domain.MembershipEventReactivated {
		t.Fatalf("eventRepo.created = %+v, want 1 reactivated event", eventRepo.created)
	}
}

func TestMembershipUsecase_AddUser_AlreadyActiveMember(t *testing.T) {
	uc, userRepo, membershipRepo, _, _ := newTestMembershipUsecase()
	accountID, ownerID := seedOrganization(t, membershipRepo)
	userID := seedUser(t, userRepo)

	existing, err := domain.NewMembership(accountID, userID, domain.MembershipRoleMember)
	if err != nil {
		t.Fatalf("NewMembership() error = %v", err)
	}
	membershipRepo.memberships[existing.ID()] = existing

	if _, err := uc.AddUser(context.Background(), AddUserInput{
		ActingUserID:    ownerID,
		TargetAccountID: accountID,
		TargetUserID:    userID,
	}); err == nil {
		t.Error("AddUser() for already-active member error = nil, want error")
	}
}

func TestMembershipUsecase_GrantAdmin(t *testing.T) {
	uc, userRepo, membershipRepo, eventRepo, _ := newTestMembershipUsecase()
	accountID, ownerID := seedOrganization(t, membershipRepo)
	memberUserID := seedUser(t, userRepo)

	member, err := domain.NewMembership(accountID, memberUserID, domain.MembershipRoleMember)
	if err != nil {
		t.Fatalf("NewMembership() error = %v", err)
	}
	membershipRepo.memberships[member.ID()] = member

	view, err := uc.GrantAdmin(context.Background(), GrantAdminInput{
		ActingUserID:    ownerID,
		TargetAccountID: accountID,
		TargetUserID:    memberUserID,
	})
	if err != nil {
		t.Fatalf("GrantAdmin() error = %v", err)
	}
	if view.Role != string(domain.MembershipRoleAdmin) {
		t.Errorf("Role = %q, want %q", view.Role, domain.MembershipRoleAdmin)
	}
	if len(eventRepo.created) != 1 || eventRepo.created[0].EventType() != domain.MembershipEventRoleChanged {
		t.Fatalf("eventRepo.created = %+v, want 1 role_changed event", eventRepo.created)
	}

	// adminは他のUserにadmin権限を付与できない。
	anotherUserID := seedUser(t, userRepo)
	another, err := domain.NewMembership(accountID, anotherUserID, domain.MembershipRoleMember)
	if err != nil {
		t.Fatalf("NewMembership() error = %v", err)
	}
	membershipRepo.memberships[another.ID()] = another

	if _, err := uc.GrantAdmin(context.Background(), GrantAdminInput{
		ActingUserID:    memberUserID, // 直前でadminに昇格させたUser
		TargetAccountID: accountID,
		TargetUserID:    anotherUserID,
	}); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("GrantAdmin() by admin error = %v, want ErrForbidden", err)
	}
}

func TestMembershipUsecase_RemoveUser_ByOwnerIsImmediate(t *testing.T) {
	uc, userRepo, membershipRepo, eventRepo, _ := newTestMembershipUsecase()
	accountID, ownerID := seedOrganization(t, membershipRepo)
	memberUserID := seedUser(t, userRepo)

	member, err := domain.NewMembership(accountID, memberUserID, domain.MembershipRoleMember)
	if err != nil {
		t.Fatalf("NewMembership() error = %v", err)
	}
	membershipRepo.memberships[member.ID()] = member

	result, err := uc.RemoveUser(context.Background(), RemoveUserInput{
		ActingUserID:    ownerID,
		TargetAccountID: accountID,
		TargetUserID:    memberUserID,
	})
	if err != nil {
		t.Fatalf("RemoveUser() error = %v", err)
	}
	if result.Membership == nil {
		t.Fatal("result.Membership = nil, want a value for owner-initiated removal")
	}
	if result.Request != nil {
		t.Error("result.Request is set, want nil for owner-initiated removal")
	}
	if result.Membership.Status != string(domain.MembershipStatusRemoved) {
		t.Errorf("Status = %q, want %q", result.Membership.Status, domain.MembershipStatusRemoved)
	}
	if len(eventRepo.created) != 1 || eventRepo.created[0].EventType() != domain.MembershipEventRemoved {
		t.Fatalf("eventRepo.created = %+v, want 1 removed event", eventRepo.created)
	}
}

func TestMembershipUsecase_RemoveUser_ByAdminCreatesPendingRequest(t *testing.T) {
	uc, userRepo, membershipRepo, eventRepo, removalRequestRepo := newTestMembershipUsecase()
	accountID, _ := seedOrganization(t, membershipRepo)
	adminUserID := seedUser(t, userRepo)
	memberUserID := seedUser(t, userRepo)

	admin, err := domain.NewMembership(accountID, adminUserID, domain.MembershipRoleAdmin)
	if err != nil {
		t.Fatalf("NewMembership() error = %v", err)
	}
	membershipRepo.memberships[admin.ID()] = admin

	member, err := domain.NewMembership(accountID, memberUserID, domain.MembershipRoleMember)
	if err != nil {
		t.Fatalf("NewMembership() error = %v", err)
	}
	membershipRepo.memberships[member.ID()] = member

	result, err := uc.RemoveUser(context.Background(), RemoveUserInput{
		ActingUserID:    adminUserID,
		TargetAccountID: accountID,
		TargetUserID:    memberUserID,
	})
	if err != nil {
		t.Fatalf("RemoveUser() error = %v", err)
	}
	if result.Request == nil {
		t.Fatal("result.Request = nil, want a value for admin-initiated removal")
	}
	if result.Membership != nil {
		t.Error("result.Membership is set, want nil for admin-initiated removal (pending approval)")
	}
	if result.Request.Status != string(domain.MembershipRemovalRequestStatusPending) {
		t.Errorf("Status = %q, want %q", result.Request.Status, domain.MembershipRemovalRequestStatusPending)
	}
	// admin操作ではMembershipのstatusはまだ変更されない。
	if membershipRepo.memberships[member.ID()].Status() != domain.MembershipStatusActive {
		t.Error("membership status changed immediately for admin-initiated removal, want it to remain active until owner approves")
	}
	if len(eventRepo.created) != 0 {
		t.Errorf("eventRepo.created = %+v, want no membership events until approved", eventRepo.created)
	}
	if len(removalRequestRepo.requests) != 1 {
		t.Fatalf("len(requests) = %d, want 1", len(removalRequestRepo.requests))
	}
}

func TestMembershipUsecase_RemoveUser_ForbiddenForMember(t *testing.T) {
	uc, userRepo, membershipRepo, _, _ := newTestMembershipUsecase()
	accountID, _ := seedOrganization(t, membershipRepo)
	memberUserID := seedUser(t, userRepo)
	targetUserID := seedUser(t, userRepo)

	member, err := domain.NewMembership(accountID, memberUserID, domain.MembershipRoleMember)
	if err != nil {
		t.Fatalf("NewMembership() error = %v", err)
	}
	membershipRepo.memberships[member.ID()] = member

	target, err := domain.NewMembership(accountID, targetUserID, domain.MembershipRoleMember)
	if err != nil {
		t.Fatalf("NewMembership() error = %v", err)
	}
	membershipRepo.memberships[target.ID()] = target

	if _, err := uc.RemoveUser(context.Background(), RemoveUserInput{
		ActingUserID:    memberUserID,
		TargetAccountID: accountID,
		TargetUserID:    targetUserID,
	}); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("RemoveUser() by member error = %v, want ErrForbidden", err)
	}
}

func TestMembershipUsecase_RemoveUser_CannotRemoveOwner(t *testing.T) {
	uc, _, membershipRepo, _, _ := newTestMembershipUsecase()
	accountID, ownerID := seedOrganization(t, membershipRepo)

	if _, err := uc.RemoveUser(context.Background(), RemoveUserInput{
		ActingUserID:    ownerID,
		TargetAccountID: accountID,
		TargetUserID:    ownerID,
	}); err == nil {
		t.Error("RemoveUser() on owner error = nil, want error")
	}
}

func TestMembershipUsecase_ApproveRemovalRequest(t *testing.T) {
	uc, userRepo, membershipRepo, eventRepo, _ := newTestMembershipUsecase()
	accountID, ownerID := seedOrganization(t, membershipRepo)
	adminUserID := seedUser(t, userRepo)
	memberUserID := seedUser(t, userRepo)

	admin, err := domain.NewMembership(accountID, adminUserID, domain.MembershipRoleAdmin)
	if err != nil {
		t.Fatalf("NewMembership() error = %v", err)
	}
	membershipRepo.memberships[admin.ID()] = admin

	member, err := domain.NewMembership(accountID, memberUserID, domain.MembershipRoleMember)
	if err != nil {
		t.Fatalf("NewMembership() error = %v", err)
	}
	membershipRepo.memberships[member.ID()] = member

	removeResult, err := uc.RemoveUser(context.Background(), RemoveUserInput{
		ActingUserID:    adminUserID,
		TargetAccountID: accountID,
		TargetUserID:    memberUserID,
	})
	if err != nil {
		t.Fatalf("RemoveUser() error = %v", err)
	}
	requestID, err := uuid.Parse(removeResult.Request.ID)
	if err != nil {
		t.Fatalf("uuid.Parse() error = %v", err)
	}

	// ownerでない場合は承認できない。
	if _, err := uc.ApproveRemovalRequest(context.Background(), ResolveRemovalRequestInput{
		ActingUserID:    adminUserID,
		TargetAccountID: accountID,
		RequestID:       requestID,
	}); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("ApproveRemovalRequest() by admin error = %v, want ErrForbidden", err)
	}

	view, err := uc.ApproveRemovalRequest(context.Background(), ResolveRemovalRequestInput{
		ActingUserID:    ownerID,
		TargetAccountID: accountID,
		RequestID:       requestID,
	})
	if err != nil {
		t.Fatalf("ApproveRemovalRequest() error = %v", err)
	}
	if view.Status != string(domain.MembershipStatusRemoved) {
		t.Errorf("Status = %q, want %q", view.Status, domain.MembershipStatusRemoved)
	}
	if len(eventRepo.created) != 1 || eventRepo.created[0].EventType() != domain.MembershipEventRemoved {
		t.Fatalf("eventRepo.created = %+v, want 1 removed event", eventRepo.created)
	}
}

func TestMembershipUsecase_RejectRemovalRequest(t *testing.T) {
	uc, userRepo, membershipRepo, eventRepo, _ := newTestMembershipUsecase()
	accountID, ownerID := seedOrganization(t, membershipRepo)
	adminUserID := seedUser(t, userRepo)
	memberUserID := seedUser(t, userRepo)

	admin, err := domain.NewMembership(accountID, adminUserID, domain.MembershipRoleAdmin)
	if err != nil {
		t.Fatalf("NewMembership() error = %v", err)
	}
	membershipRepo.memberships[admin.ID()] = admin

	member, err := domain.NewMembership(accountID, memberUserID, domain.MembershipRoleMember)
	if err != nil {
		t.Fatalf("NewMembership() error = %v", err)
	}
	membershipRepo.memberships[member.ID()] = member

	removeResult, err := uc.RemoveUser(context.Background(), RemoveUserInput{
		ActingUserID:    adminUserID,
		TargetAccountID: accountID,
		TargetUserID:    memberUserID,
	})
	if err != nil {
		t.Fatalf("RemoveUser() error = %v", err)
	}
	requestID, err := uuid.Parse(removeResult.Request.ID)
	if err != nil {
		t.Fatalf("uuid.Parse() error = %v", err)
	}

	view, err := uc.RejectRemovalRequest(context.Background(), ResolveRemovalRequestInput{
		ActingUserID:    ownerID,
		TargetAccountID: accountID,
		RequestID:       requestID,
	})
	if err != nil {
		t.Fatalf("RejectRemovalRequest() error = %v", err)
	}
	if view.Status != string(domain.MembershipRemovalRequestStatusRejected) {
		t.Errorf("Status = %q, want %q", view.Status, domain.MembershipRemovalRequestStatusRejected)
	}
	if membershipRepo.memberships[member.ID()].Status() != domain.MembershipStatusActive {
		t.Error("membership status changed after rejection, want it to remain active")
	}
	if len(eventRepo.created) != 0 {
		t.Errorf("eventRepo.created = %+v, want no membership events after rejection", eventRepo.created)
	}
}

func TestMembershipUsecase_ListMembers_RequiresOwnerOrAdmin(t *testing.T) {
	uc, userRepo, membershipRepo, _, _ := newTestMembershipUsecase()
	accountID, ownerID := seedOrganization(t, membershipRepo)
	memberUserID := seedUser(t, userRepo)

	member, err := domain.NewMembership(accountID, memberUserID, domain.MembershipRoleMember)
	if err != nil {
		t.Fatalf("NewMembership() error = %v", err)
	}
	membershipRepo.memberships[member.ID()] = member

	if _, err := uc.ListMembers(context.Background(), memberUserID, accountID); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("ListMembers() by member error = %v, want ErrForbidden", err)
	}

	views, err := uc.ListMembers(context.Background(), ownerID, accountID)
	if err != nil {
		t.Fatalf("ListMembers() error = %v", err)
	}
	if len(views) != 2 {
		t.Fatalf("len(views) = %d, want 2", len(views))
	}
}
