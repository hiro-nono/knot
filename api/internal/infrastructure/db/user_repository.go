package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"knot-api/internal/domain"

	"uuid"
)

// UserRepository はPostgreSQLを用いたUserの永続化を行う。
type UserRepository struct {
	conn *sql.DB
}

// NewUserRepository はUserRepositoryを生成する。
func NewUserRepository(conn *sql.DB) *UserRepository {
	return &UserRepository{conn: conn}
}

// Create はUserを新規保存する。
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	_, err := executorFromContext(ctx, r.conn).ExecContext(ctx, `
		INSERT INTO users (id, last_name, first_name, language, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`,
		user.ID().String(),
		user.LastName(),
		user.FirstName(),
		user.Language(),
		user.CreatedAt(),
		user.UpdatedAt(),
	)
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}

	return nil
}

// Get はIDを指定してUserを1件取得する。
func (r *UserRepository) Get(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	row := executorFromContext(ctx, r.conn).QueryRowContext(ctx, `
		SELECT id, last_name, first_name, language, created_at, updated_at
		FROM users
		WHERE id = $1
	`, id.String())

	var (
		rawID     string
		lastName  string
		firstName string
		language  string
		createdAt time.Time
		updatedAt time.Time
	)
	if err := row.Scan(&rawID, &lastName, &firstName, &language, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("select user: %w", err)
	}

	parsedID, err := uuid.Parse(rawID)
	if err != nil {
		return nil, fmt.Errorf("parse user id: %w", err)
	}

	return domain.ReconstructUser(parsedID, lastName, firstName, language, createdAt, updatedAt), nil
}

// Update はUserのプロフィールを更新する。
func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	result, err := executorFromContext(ctx, r.conn).ExecContext(ctx, `
		UPDATE users
		SET last_name = $2, first_name = $3, language = $4, updated_at = $5
		WHERE id = $1
	`,
		user.ID().String(),
		user.LastName(),
		user.FirstName(),
		user.Language(),
		user.UpdatedAt(),
	)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check update user result: %w", err)
	}
	if affected == 0 {
		return domain.ErrNotFound
	}

	return nil
}
