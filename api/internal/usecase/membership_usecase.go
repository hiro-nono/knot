package usecase

import (
	"context"
	"errors"
	"fmt"

	"knot-api/internal/domain"
	"knot-api/internal/repository"

	"uuid"
)

// MembershipUsecase はOrganization(Account)内でのUserの所属・権限(role)・
// 状態(status)を管理するアプリケーションフローを担う。
//
// 権限判定はすべてこのUseCase層で行い、クライアントから送信された
// account_id・roleをそのまま信頼することはない。操作を行うUserは、
// 呼び出し元(controller)がJWTから解決したactingUserIDを起点として、
// 対象Account(targetAccountID)における「実際のMembership」をDBから
// 取得して初めて権限が確定する。
//
// ルール:
//   - owner: Organizationの所有者。新しいUserの追加、既存UserへのADMIN権限の
//     付与ができる。除外操作は即座に反映される。
//   - admin: Userの管理(除外)ができるが、ownerの承認が必要
//     (即座には反映されず、MembershipRemovalRequestとして記録される)。
//     adminは他のUserにadmin権限を付与できない。
//   - member: User管理権限を持たない。
type MembershipUsecase struct {
	userRepo            repository.UserRepository
	membershipRepo      repository.MembershipRepository
	membershipEventRepo repository.MembershipEventRepository
	removalRequestRepo  repository.MembershipRemovalRequestRepository
	transactionManager  repository.TransactionManager
}

// NewMembershipUsecase はMembershipUsecaseを生成する。
func NewMembershipUsecase(
	userRepo repository.UserRepository,
	membershipRepo repository.MembershipRepository,
	membershipEventRepo repository.MembershipEventRepository,
	removalRequestRepo repository.MembershipRemovalRequestRepository,
	transactionManager repository.TransactionManager,
) *MembershipUsecase {
	return &MembershipUsecase{
		userRepo:            userRepo,
		membershipRepo:      membershipRepo,
		membershipEventRepo: membershipEventRepo,
		removalRequestRepo:  removalRequestRepo,
		transactionManager:  transactionManager,
	}
}

// AddUserInput はAddUserへの入力。
type AddUserInput struct {
	ActingUserID    uuid.UUID
	TargetAccountID uuid.UUID
	TargetUserID    uuid.UUID
}

// AddUser はOrganizationへ新しいUserを追加する(owner専用)。
// 既に(active/removedを問わず)Membershipが存在する場合は、新規作成せず
// 既存Membershipを再利用する。activeであればエラー、removedであれば
// reactivateする(履歴を一貫して保持するため)。新規追加の場合、初期roleは
// memberとする。
func (u *MembershipUsecase) AddUser(ctx context.Context, input AddUserInput) (*MembershipView, error) {
	if _, err := u.requireActiveRole(ctx, input.TargetAccountID, input.ActingUserID, domain.MembershipRoleOwner); err != nil {
		return nil, err
	}

	if _, err := u.userRepo.Get(ctx, input.TargetUserID); err != nil {
		return nil, fmt.Errorf("get target user: %w", err)
	}

	existing, err := u.membershipRepo.GetByAccountAndUser(ctx, input.TargetAccountID, input.TargetUserID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, fmt.Errorf("get existing membership: %w", err)
	}

	if err == nil {
		if existing.IsActive() {
			return nil, fmt.Errorf("user is already a member of this account")
		}

		eventType, err := existing.Reactivate()
		if err != nil {
			return nil, fmt.Errorf("reactivate membership: %w", err)
		}

		if err := u.saveMembershipWithEvent(ctx, existing, eventType); err != nil {
			return nil, err
		}

		return newMembershipView(existing), nil
	}

	membership, err := domain.NewMembership(input.TargetAccountID, input.TargetUserID, domain.MembershipRoleMember)
	if err != nil {
		return nil, fmt.Errorf("build membership: %w", err)
	}

	if err := u.createMembershipWithEvent(ctx, membership, domain.MembershipEventCreated); err != nil {
		return nil, err
	}

	return newMembershipView(membership), nil
}

// GrantAdminInput はGrantAdminへの入力。
type GrantAdminInput struct {
	ActingUserID    uuid.UUID
	TargetAccountID uuid.UUID
	TargetUserID    uuid.UUID
}

// GrantAdmin は既存UserのMembershipにADMIN権限を付与する(owner専用)。
// adminは他のUserにadmin権限を付与できない。
func (u *MembershipUsecase) GrantAdmin(ctx context.Context, input GrantAdminInput) (*MembershipView, error) {
	if _, err := u.requireActiveRole(ctx, input.TargetAccountID, input.ActingUserID, domain.MembershipRoleOwner); err != nil {
		return nil, err
	}

	target, err := u.membershipRepo.GetByAccountAndUser(ctx, input.TargetAccountID, input.TargetUserID)
	if err != nil {
		return nil, fmt.Errorf("get target membership: %w", err)
	}

	eventType, err := target.ChangeRole(domain.MembershipRoleAdmin)
	if err != nil {
		return nil, fmt.Errorf("change role: %w", err)
	}

	if err := u.saveMembershipWithEvent(ctx, target, eventType); err != nil {
		return nil, err
	}

	return newMembershipView(target), nil
}

// RemoveUserInput はRemoveUserへの入力。
type RemoveUserInput struct {
	ActingUserID    uuid.UUID
	TargetAccountID uuid.UUID
	TargetUserID    uuid.UUID
}

// RemoveUserResult はRemoveUserの出力。ownerが実行した場合はMembershipが
// 即座にremovedされてMembershipに値が入り、adminが実行した場合は
// ownerの承認待ちのMembershipRemovalRequestが作成されてRequestに値が入る。
type RemoveUserResult struct {
	Membership *MembershipView
	Request    *MembershipRemovalRequestView
}

// RemoveUser はUserをOrganizationから除外する。
// ownerが実行した場合は即座にMembershipがremoved状態になる。
// adminが実行した場合は、ownerの承認が必要なため即座には反映されず、
// MembershipRemovalRequest(pending)として記録される。
// memberにはUser管理権限が無いため実行できない。
func (u *MembershipUsecase) RemoveUser(ctx context.Context, input RemoveUserInput) (*RemoveUserResult, error) {
	actingMembership, err := u.requireActiveMembership(ctx, input.TargetAccountID, input.ActingUserID)
	if err != nil {
		return nil, err
	}
	if !actingMembership.IsOwner() && !actingMembership.IsAdmin() {
		return nil, domain.ErrForbidden
	}

	target, err := u.membershipRepo.GetByAccountAndUser(ctx, input.TargetAccountID, input.TargetUserID)
	if err != nil {
		return nil, fmt.Errorf("get target membership: %w", err)
	}

	if actingMembership.IsOwner() {
		eventType, err := target.Remove()
		if err != nil {
			return nil, fmt.Errorf("remove membership: %w", err)
		}
		if err := u.saveMembershipWithEvent(ctx, target, eventType); err != nil {
			return nil, err
		}
		return &RemoveUserResult{Membership: newMembershipView(target)}, nil
	}

	// admin: ownerの承認が必要なため、即座にはstatusを変更せず
	// MembershipRemovalRequestを作成するのみに留める。
	if !target.IsActive() {
		return nil, fmt.Errorf("target membership is not active")
	}
	if target.IsOwner() {
		return nil, domain.ErrForbidden
	}
	if _, err := u.removalRequestRepo.GetPendingByMembershipID(ctx, target.ID()); err == nil {
		return nil, fmt.Errorf("a pending removal request already exists for this membership")
	} else if !errors.Is(err, domain.ErrNotFound) {
		return nil, fmt.Errorf("get pending removal request: %w", err)
	}

	request, err := domain.NewMembershipRemovalRequest(target.ID(), input.ActingUserID)
	if err != nil {
		return nil, fmt.Errorf("build removal request: %w", err)
	}

	if err := u.transactionManager.WithinTransaction(ctx, func(ctx context.Context) error {
		return u.removalRequestRepo.Create(ctx, request)
	}); err != nil {
		return nil, err
	}

	return &RemoveUserResult{Request: newMembershipRemovalRequestView(request)}, nil
}

// ResolveRemovalRequestInput はApproveRemovalRequest/RejectRemovalRequestへの入力。
type ResolveRemovalRequestInput struct {
	ActingUserID    uuid.UUID
	TargetAccountID uuid.UUID
	RequestID       uuid.UUID
}

// ApproveRemovalRequest はadminが作成した除外申請をownerが承認し、
// Membershipを実際にremoved状態にする(owner専用)。
func (u *MembershipUsecase) ApproveRemovalRequest(ctx context.Context, input ResolveRemovalRequestInput) (*MembershipView, error) {
	if _, err := u.requireActiveRole(ctx, input.TargetAccountID, input.ActingUserID, domain.MembershipRoleOwner); err != nil {
		return nil, err
	}

	request, membership, err := u.loadRemovalRequestForAccount(ctx, input.TargetAccountID, input.RequestID)
	if err != nil {
		return nil, err
	}

	if err := request.Approve(); err != nil {
		return nil, fmt.Errorf("approve removal request: %w", err)
	}
	eventType, err := membership.Remove()
	if err != nil {
		return nil, fmt.Errorf("remove membership: %w", err)
	}

	event, err := domain.NewMembershipEvent(membership.ID(), eventType)
	if err != nil {
		return nil, fmt.Errorf("build membership event: %w", err)
	}

	if err := u.transactionManager.WithinTransaction(ctx, func(ctx context.Context) error {
		if err := u.removalRequestRepo.Update(ctx, request); err != nil {
			return fmt.Errorf("update removal request: %w", err)
		}
		if err := u.membershipRepo.Update(ctx, membership); err != nil {
			return fmt.Errorf("update membership: %w", err)
		}
		return u.membershipEventRepo.Create(ctx, event)
	}); err != nil {
		return nil, err
	}

	return newMembershipView(membership), nil
}

// RejectRemovalRequest はadminが作成した除外申請をownerが却下する(owner専用)。
// Membershipのstatusは変更されない。
func (u *MembershipUsecase) RejectRemovalRequest(ctx context.Context, input ResolveRemovalRequestInput) (*MembershipRemovalRequestView, error) {
	if _, err := u.requireActiveRole(ctx, input.TargetAccountID, input.ActingUserID, domain.MembershipRoleOwner); err != nil {
		return nil, err
	}

	request, _, err := u.loadRemovalRequestForAccount(ctx, input.TargetAccountID, input.RequestID)
	if err != nil {
		return nil, err
	}

	if err := request.Reject(); err != nil {
		return nil, fmt.Errorf("reject removal request: %w", err)
	}

	if err := u.transactionManager.WithinTransaction(ctx, func(ctx context.Context) error {
		return u.removalRequestRepo.Update(ctx, request)
	}); err != nil {
		return nil, err
	}

	return newMembershipRemovalRequestView(request), nil
}

// ListMembers はOrganizationに所属する(statusを問わない)Membershipを一覧取得する
// (owner・admin専用。memberはUser管理権限を持たないため参照できない)。
func (u *MembershipUsecase) ListMembers(ctx context.Context, actingUserID, targetAccountID uuid.UUID) ([]*MembershipView, error) {
	actingMembership, err := u.requireActiveMembership(ctx, targetAccountID, actingUserID)
	if err != nil {
		return nil, err
	}
	if !actingMembership.IsOwner() && !actingMembership.IsAdmin() {
		return nil, domain.ErrForbidden
	}

	memberships, err := u.membershipRepo.ListByAccountID(ctx, targetAccountID)
	if err != nil {
		return nil, fmt.Errorf("list memberships: %w", err)
	}

	views := make([]*MembershipView, 0, len(memberships))
	for _, m := range memberships {
		views = append(views, newMembershipView(m))
	}

	return views, nil
}

// ListPendingRemovalRequests はOrganization内の承認待ちMembershipRemovalRequestを
// 一覧取得する(owner専用)。
func (u *MembershipUsecase) ListPendingRemovalRequests(ctx context.Context, actingUserID, targetAccountID uuid.UUID) ([]*MembershipRemovalRequestView, error) {
	if _, err := u.requireActiveRole(ctx, targetAccountID, actingUserID, domain.MembershipRoleOwner); err != nil {
		return nil, err
	}

	requests, err := u.removalRequestRepo.ListPendingByAccountID(ctx, targetAccountID)
	if err != nil {
		return nil, fmt.Errorf("list pending removal requests: %w", err)
	}

	views := make([]*MembershipRemovalRequestView, 0, len(requests))
	for _, r := range requests {
		views = append(views, newMembershipRemovalRequestView(r))
	}

	return views, nil
}

// loadRemovalRequestForAccount はrequestIDのMembershipRemovalRequestと、
// その対象Membershipを取得する。Membershipが指定accountIDに属していない
// 場合はErrForbiddenを返す(他Organizationのリクエストを操作させないため)。
func (u *MembershipUsecase) loadRemovalRequestForAccount(ctx context.Context, accountID, requestID uuid.UUID) (*domain.MembershipRemovalRequest, *domain.Membership, error) {
	request, err := u.removalRequestRepo.Get(ctx, requestID)
	if err != nil {
		return nil, nil, fmt.Errorf("get removal request: %w", err)
	}

	membership, err := u.membershipRepo.Get(ctx, request.MembershipID())
	if err != nil {
		return nil, nil, fmt.Errorf("get membership: %w", err)
	}
	if membership.AccountID() != accountID {
		return nil, nil, domain.ErrForbidden
	}

	return request, membership, nil
}

// requireActiveMembership はaccountID・userIDに対応するactiveなMembershipを取得する。
// 見つからない、またはactiveでない場合はdomain.ErrForbiddenを返す
// (クライアントが送信したaccount_id・roleを信頼せず、必ずDB上のMembershipで判定する)。
func (u *MembershipUsecase) requireActiveMembership(ctx context.Context, accountID, userID uuid.UUID) (*domain.Membership, error) {
	membership, err := u.membershipRepo.GetByAccountAndUser(ctx, accountID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrForbidden
		}
		return nil, fmt.Errorf("get membership: %w", err)
	}
	if !membership.IsActive() {
		return nil, domain.ErrForbidden
	}
	return membership, nil
}

// requireActiveRole はaccountID・userIDに対応するactiveなMembershipを取得し、
// その役割がrequiredRoleであることを検証する。
func (u *MembershipUsecase) requireActiveRole(ctx context.Context, accountID, userID uuid.UUID, requiredRole domain.MembershipRole) (*domain.Membership, error) {
	membership, err := u.requireActiveMembership(ctx, accountID, userID)
	if err != nil {
		return nil, err
	}
	if membership.Role() != requiredRole {
		return nil, domain.ErrForbidden
	}
	return membership, nil
}

func (u *MembershipUsecase) createMembershipWithEvent(ctx context.Context, membership *domain.Membership, eventType domain.MembershipEventType) error {
	event, err := domain.NewMembershipEvent(membership.ID(), eventType)
	if err != nil {
		return fmt.Errorf("build membership event: %w", err)
	}

	return u.transactionManager.WithinTransaction(ctx, func(ctx context.Context) error {
		if err := u.membershipRepo.Create(ctx, membership); err != nil {
			return fmt.Errorf("create membership: %w", err)
		}
		return u.membershipEventRepo.Create(ctx, event)
	})
}

func (u *MembershipUsecase) saveMembershipWithEvent(ctx context.Context, membership *domain.Membership, eventType domain.MembershipEventType) error {
	event, err := domain.NewMembershipEvent(membership.ID(), eventType)
	if err != nil {
		return fmt.Errorf("build membership event: %w", err)
	}

	return u.transactionManager.WithinTransaction(ctx, func(ctx context.Context) error {
		if err := u.membershipRepo.Update(ctx, membership); err != nil {
			return fmt.Errorf("update membership: %w", err)
		}
		return u.membershipEventRepo.Create(ctx, event)
	})
}
