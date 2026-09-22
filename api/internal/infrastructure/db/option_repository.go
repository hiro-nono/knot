package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"knot-api/internal/domain"

	"uuid"
)

// OptionRepository はPostgreSQLを用いたOptionの永続化を行う。
type OptionRepository struct {
	conn *sql.DB
}

// NewOptionRepository はOptionRepositoryを生成する。
func NewOptionRepository(conn *sql.DB) *OptionRepository {
	return &OptionRepository{conn: conn}
}

// Create はOptionを新規保存する。
func (r *OptionRepository) Create(ctx context.Context, option *domain.Option) error {
	_, err := executorFromContext(ctx, r.conn).ExecContext(ctx, `
		INSERT INTO options (id, source_id, value, sort_order, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`,
		option.ID().String(),
		option.SourceID().String(),
		option.Value(),
		option.SortOrder(),
		option.CreatedAt(),
		option.UpdatedAt(),
	)
	if err != nil {
		return fmt.Errorf("insert option: %w", err)
	}

	return nil
}

// List は指定したSourceに属するOptionを一覧取得する。
func (r *OptionRepository) List(ctx context.Context, sourceID uuid.UUID) ([]*domain.Option, error) {
	rows, err := executorFromContext(ctx, r.conn).QueryContext(ctx, `
		SELECT id, source_id, value, sort_order, created_at, updated_at
		FROM options
		WHERE source_id = $1
		ORDER BY sort_order
	`, sourceID.String())
	if err != nil {
		return nil, fmt.Errorf("select options: %w", err)
	}
	defer rows.Close()

	var options []*domain.Option
	for rows.Next() {
		var (
			rawID       string
			rawSourceID string
			value       string
			sortOrder   int
			createdAt   time.Time
			updatedAt   time.Time
		)
		if err := rows.Scan(&rawID, &rawSourceID, &value, &sortOrder, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("scan option: %w", err)
		}

		id, err := uuid.Parse(rawID)
		if err != nil {
			return nil, fmt.Errorf("parse option id: %w", err)
		}
		parsedSourceID, err := uuid.Parse(rawSourceID)
		if err != nil {
			return nil, fmt.Errorf("parse source id: %w", err)
		}

		options = append(options, domain.ReconstructOption(id, parsedSourceID, value, sortOrder, createdAt, updatedAt))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate options: %w", err)
	}

	return options, nil
}
