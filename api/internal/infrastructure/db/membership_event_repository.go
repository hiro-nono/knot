package db

import (
	"context"
	"database/sql"
	"fmt"

	"knot-api/internal/domain"
)

// MembershipEventRepository はPostgreSQLを用いたMembershipEvent(履歴)の
// 永続化を行う。
type MembershipEventRepository struct {
	conn *sql.DB
}

// NewMembershipEventRepository はMembershipEventRepositoryを生成する。
func NewMembershipEventRepository(conn *sql.DB) *MembershipEventRepository {
	return &MembershipEventRepository{conn: conn}
}

// Create はMembershipEventを新規保存する。
func (r *MembershipEventRepository) Create(ctx context.Context, event *domain.MembershipEvent) error {
	_, err := executorFromContext(ctx, r.conn).ExecContext(ctx, `
		INSERT INTO membership_events (id, membership_id, event_type, created_at)
		VALUES ($1, $2, $3, $4)
	`,
		event.ID().String(),
		event.MembershipID().String(),
		string(event.EventType()),
		event.CreatedAt(),
	)
	if err != nil {
		return fmt.Errorf("insert membership event: %w", err)
	}

	return nil
}
