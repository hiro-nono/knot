package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"knot-api/internal/domain"

	"uuid"
)

// PreferenceRepository はPostgreSQLを用いたPreferenceの永続化を行う。
type PreferenceRepository struct {
	conn *sql.DB
}

// NewPreferenceRepository はPreferenceRepositoryを生成する。
func NewPreferenceRepository(conn *sql.DB) *PreferenceRepository {
	return &PreferenceRepository{conn: conn}
}

// Get はuserIDに対応するPreferenceを取得する。
func (r *PreferenceRepository) Get(ctx context.Context, userID uuid.UUID) (*domain.Preference, error) {
	row := executorFromContext(ctx, r.conn).QueryRowContext(ctx, `
		SELECT id, user_id, items, created_at, updated_at
		FROM preferences
		WHERE user_id = $1
	`, userID.String())

	var (
		rawID     string
		rawUserID string
		rawItems  []byte
		createdAt time.Time
		updatedAt time.Time
	)
	if err := row.Scan(&rawID, &rawUserID, &rawItems, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("select preference: %w", err)
	}

	id, err := uuid.Parse(rawID)
	if err != nil {
		return nil, fmt.Errorf("parse preference id: %w", err)
	}
	parsedUserID, err := uuid.Parse(rawUserID)
	if err != nil {
		return nil, fmt.Errorf("parse user id: %w", err)
	}

	var items map[string]string
	if err := json.Unmarshal(rawItems, &items); err != nil {
		return nil, fmt.Errorf("unmarshal preference items: %w", err)
	}

	return domain.ReconstructPreference(id, parsedUserID, items, createdAt, updatedAt), nil
}

// Save はPreferenceを保存する(user_idに既存レコードがあれば更新、無ければ新規作成)。
func (r *PreferenceRepository) Save(ctx context.Context, preference *domain.Preference) error {
	items, err := json.Marshal(preference.Items())
	if err != nil {
		return fmt.Errorf("marshal preference items: %w", err)
	}

	_, err = executorFromContext(ctx, r.conn).ExecContext(ctx, `
		INSERT INTO preferences (id, user_id, items, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id) DO UPDATE
		SET items = EXCLUDED.items, updated_at = EXCLUDED.updated_at
	`,
		preference.ID().String(),
		preference.UserID().String(),
		items,
		preference.CreatedAt(),
		preference.UpdatedAt(),
	)
	if err != nil {
		return fmt.Errorf("upsert preference: %w", err)
	}

	return nil
}
