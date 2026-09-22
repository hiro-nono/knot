package repository

import (
	"context"

	"knot-api/internal/domain"

	"uuid"
)

// MembershipRemovalRequestRepository はMembershipRemovalRequest
// (Adminによる除外申請とOwnerの承認待ち状態)の永続化を抽象化する。
type MembershipRemovalRequestRepository interface {
	// Create はMembershipRemovalRequestを新規保存する。
	Create(ctx context.Context, request *domain.MembershipRemovalRequest) error
	// Get はIDを指定してMembershipRemovalRequestを1件取得する。
	// 見つからない場合はdomain.ErrNotFoundを返す。
	Get(ctx context.Context, id uuid.UUID) (*domain.MembershipRemovalRequest, error)
	// GetPendingByMembershipID はmembershipIDに対応するpending状態の
	// MembershipRemovalRequestを取得する(二重申請の防止に使う)。
	// 見つからない場合はdomain.ErrNotFoundを返す。
	GetPendingByMembershipID(ctx context.Context, membershipID uuid.UUID) (*domain.MembershipRemovalRequest, error)
	// ListPendingByAccountID はaccountIDに属するMembershipのうち、
	// pending状態のMembershipRemovalRequestを一覧取得する(Ownerのレビュー用)。
	ListPendingByAccountID(ctx context.Context, accountID uuid.UUID) ([]*domain.MembershipRemovalRequest, error)
	// Update はMembershipRemovalRequestのstatus等を更新する。
	Update(ctx context.Context, request *domain.MembershipRemovalRequest) error
}
