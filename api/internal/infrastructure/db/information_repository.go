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

// InformationRepository はPostgreSQLを用いたInformationの永続化を行う。
type InformationRepository struct {
	conn *sql.DB
}

// NewInformationRepository はInformationRepositoryを生成する。
func NewInformationRepository(conn *sql.DB) *InformationRepository {
	return &InformationRepository{conn: conn}
}

// Create はInformationを新規保存する。
func (r *InformationRepository) Create(ctx context.Context, information *domain.Information) error {
	_, err := executorFromContext(ctx, r.conn).ExecContext(ctx, `
		INSERT INTO informations (id, account_id, created_by_user_id, title, access_type, response_policy, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`,
		information.ID().String(),
		information.AccountID().String(),
		information.CreatedByUserID().String(),
		information.Title(),
		string(information.AccessType()),
		string(information.ResponsePolicy()),
		information.CreatedAt(),
		information.UpdatedAt(),
	)
	if err != nil {
		return fmt.Errorf("insert information: %w", err)
	}

	return nil
}

// Get はIDを指定してInformationを1件取得する。
func (r *InformationRepository) Get(ctx context.Context, id uuid.UUID) (*domain.Information, error) {
	row := executorFromContext(ctx, r.conn).QueryRowContext(ctx, `
		SELECT id, account_id, created_by_user_id, title, access_type, response_policy, created_at, updated_at
		FROM informations
		WHERE id = $1
	`, id.String())

	var (
		rawID              string
		rawAccountID       string
		rawCreatedByUserID string
		title              string
		accessType         string
		responsePolicy     string
		createdAt          time.Time
		updatedAt          time.Time
	)
	if err := row.Scan(&rawID, &rawAccountID, &rawCreatedByUserID, &title, &accessType, &responsePolicy, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("select information: %w", err)
	}

	informationID, err := uuid.Parse(rawID)
	if err != nil {
		return nil, fmt.Errorf("parse information id: %w", err)
	}
	accountID, err := uuid.Parse(rawAccountID)
	if err != nil {
		return nil, fmt.Errorf("parse account id: %w", err)
	}
	createdByUserID, err := uuid.Parse(rawCreatedByUserID)
	if err != nil {
		return nil, fmt.Errorf("parse created_by_user_id: %w", err)
	}

	return domain.ReconstructInformation(informationID, accountID, createdByUserID, title, domain.InformationAccessType(accessType), domain.InformationResponsePolicy(responsePolicy), nil, createdAt, updatedAt), nil
}

// ListByAccountID は指定したAccountが所有するInformationを、
// 作成日時の新しい順に一覧取得する。
func (r *InformationRepository) ListByAccountID(ctx context.Context, accountID uuid.UUID) ([]*domain.Information, error) {
	rows, err := executorFromContext(ctx, r.conn).QueryContext(ctx, `
		SELECT id, account_id, created_by_user_id, title, access_type, response_policy, created_at, updated_at
		FROM informations
		WHERE account_id = $1
		ORDER BY created_at DESC
	`, accountID.String())
	if err != nil {
		return nil, fmt.Errorf("select informations: %w", err)
	}
	defer rows.Close()

	result := make([]*domain.Information, 0)
	for rows.Next() {
		var (
			rawID              string
			rawAccountID       string
			rawCreatedByUserID string
			title              string
			accessType         string
			responsePolicy     string
			createdAt          time.Time
			updatedAt          time.Time
		)
		if err := rows.Scan(&rawID, &rawAccountID, &rawCreatedByUserID, &title, &accessType, &responsePolicy, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("scan information: %w", err)
		}

		informationID, err := uuid.Parse(rawID)
		if err != nil {
			return nil, fmt.Errorf("parse information id: %w", err)
		}
		parsedAccountID, err := uuid.Parse(rawAccountID)
		if err != nil {
			return nil, fmt.Errorf("parse account id: %w", err)
		}
		createdByUserID, err := uuid.Parse(rawCreatedByUserID)
		if err != nil {
			return nil, fmt.Errorf("parse created_by_user_id: %w", err)
		}

		result = append(result, domain.ReconstructInformation(informationID, parsedAccountID, createdByUserID, title, domain.InformationAccessType(accessType), domain.InformationResponsePolicy(responsePolicy), nil, createdAt, updatedAt))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate informations: %w", err)
	}

	return result, nil
}
