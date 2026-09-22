package db

import (
	"context"
	"database/sql"
	"fmt"

	"knot-api/internal/domain"
)

// AccountStatusEventRepository はPostgreSQLを用いたAccountStatusEvent(履歴)の
// 永続化を行う。
type AccountStatusEventRepository struct {
	conn *sql.DB
}

// NewAccountStatusEventRepository はAccountStatusEventRepositoryを生成する。
func NewAccountStatusEventRepository(conn *sql.DB) *AccountStatusEventRepository {
	return &AccountStatusEventRepository{conn: conn}
}

// Create はAccountStatusEventを新規保存する。
func (r *AccountStatusEventRepository) Create(ctx context.Context, event *domain.AccountStatusEvent) error {
	_, err := executorFromContext(ctx, r.conn).ExecContext(ctx, `
		INSERT INTO account_status_events (id, account_id, event_type, created_at)
		VALUES ($1, $2, $3, $4)
	`,
		event.ID().String(),
		event.AccountID().String(),
		string(event.EventType()),
		event.CreatedAt(),
	)
	if err != nil {
		return fmt.Errorf("insert account status event: %w", err)
	}

	return nil
}
