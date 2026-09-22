package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"knot-api/internal/domain"

	"uuid"
)

// RecipientRepository はPostgreSQLを用いたRecipientの永続化を行う。
type RecipientRepository struct {
	conn *sql.DB
}

// NewRecipientRepository はRecipientRepositoryを生成する。
func NewRecipientRepository(conn *sql.DB) *RecipientRepository {
	return &RecipientRepository{conn: conn}
}

// Create はRecipientを新規保存する。
func (r *RecipientRepository) Create(ctx context.Context, recipient *domain.Recipient) error {
	_, err := executorFromContext(ctx, r.conn).ExecContext(ctx, `
		INSERT INTO recipients (id, information_id, user_id, created_at)
		VALUES ($1, $2, $3, $4)
	`,
		recipient.ID().String(),
		recipient.InformationID().String(),
		recipient.UserID().String(),
		recipient.CreatedAt(),
	)
	if err != nil {
		return fmt.Errorf("insert recipient: %w", err)
	}

	return nil
}

// Exists はuserIDがinformationIDのRecipientとして登録されているかを返す。
func (r *RecipientRepository) Exists(ctx context.Context, informationID, userID uuid.UUID) (bool, error) {
	row := executorFromContext(ctx, r.conn).QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM recipients WHERE information_id = $1 AND user_id = $2
		)
	`, informationID.String(), userID.String())

	var exists bool
	if err := row.Scan(&exists); err != nil {
		return false, fmt.Errorf("select recipient exists: %w", err)
	}

	return exists, nil
}

// ListByInformationID はinformationIDに対応するRecipientを一覧取得する。
func (r *RecipientRepository) ListByInformationID(ctx context.Context, informationID uuid.UUID) ([]*domain.Recipient, error) {
	rows, err := executorFromContext(ctx, r.conn).QueryContext(ctx, `
		SELECT id, information_id, user_id, created_at
		FROM recipients
		WHERE information_id = $1
		ORDER BY created_at
	`, informationID.String())
	if err != nil {
		return nil, fmt.Errorf("select recipients: %w", err)
	}
	defer rows.Close()

	var recipients []*domain.Recipient
	for rows.Next() {
		recipient, err := scanRecipientRow(rows)
		if err != nil {
			return nil, err
		}
		recipients = append(recipients, recipient)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate recipients: %w", err)
	}

	return recipients, nil
}

// Delete はinformationID・userIDに対応するRecipientを削除する。
func (r *RecipientRepository) Delete(ctx context.Context, informationID, userID uuid.UUID) error {
	result, err := executorFromContext(ctx, r.conn).ExecContext(ctx, `
		DELETE FROM recipients WHERE information_id = $1 AND user_id = $2
	`, informationID.String(), userID.String())
	if err != nil {
		return fmt.Errorf("delete recipient: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check delete recipient result: %w", err)
	}
	if affected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func scanRecipientRow(scanner rowScanner) (*domain.Recipient, error) {
	var (
		rawID            string
		rawInformationID string
		rawUserID        string
		createdAt        time.Time
	)
	if err := scanner.Scan(&rawID, &rawInformationID, &rawUserID, &createdAt); err != nil {
		return nil, fmt.Errorf("scan recipient: %w", err)
	}

	id, err := uuid.Parse(rawID)
	if err != nil {
		return nil, fmt.Errorf("parse recipient id: %w", err)
	}
	informationID, err := uuid.Parse(rawInformationID)
	if err != nil {
		return nil, fmt.Errorf("parse information id: %w", err)
	}
	userID, err := uuid.Parse(rawUserID)
	if err != nil {
		return nil, fmt.Errorf("parse user id: %w", err)
	}

	return domain.ReconstructRecipient(id, informationID, userID, createdAt), nil
}
