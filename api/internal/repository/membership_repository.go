package repository

import (
	"context"

	"knot-api/internal/domain"

	"uuid"
)

// MembershipRepository はMembership(UserのAccountへの所属・権限)の永続化を抽象化する。
type MembershipRepository interface {
	// Create はMembershipを新規保存する。
	Create(ctx context.Context, membership *domain.Membership) error
	// Get はIDを指定してMembershipを1件取得する。
	// 見つからない場合はdomain.ErrNotFoundを返す。
	Get(ctx context.Context, id uuid.UUID) (*domain.Membership, error)
	// GetByAccountAndUser はaccountID・userIDに対応するMembershipを取得する
	// (statusを問わない)。同一Account・同一UserのMembershipは常に1件のみ
	// 存在するため、再追加時はこれを取得して再利用する。
	// 見つからない場合はdomain.ErrNotFoundを返す。
	GetByAccountAndUser(ctx context.Context, accountID, userID uuid.UUID) (*domain.Membership, error)
	// GetOwnerByAccountID はaccountIDに対応するowner roleのMembershipを取得する。
	// 見つからない場合はdomain.ErrNotFoundを返す。
	GetOwnerByAccountID(ctx context.Context, accountID uuid.UUID) (*domain.Membership, error)
	// ListByAccountID はaccountIDに対応するMembershipを一覧取得する
	// (Organizationに所属する全member、statusを問わない)。
	ListByAccountID(ctx context.Context, accountID uuid.UUID) ([]*domain.Membership, error)
	// Update はMembershipのrole・status等を更新する。
	Update(ctx context.Context, membership *domain.Membership) error
}
