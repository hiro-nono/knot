package repository

import (
	"context"

	"knot-api/internal/domain"
)

// MembershipEventRepository はMembershipEvent(履歴、追記のみ)の永続化を抽象化する。
type MembershipEventRepository interface {
	// Create はMembershipEventを新規保存する。
	Create(ctx context.Context, event *domain.MembershipEvent) error
}
