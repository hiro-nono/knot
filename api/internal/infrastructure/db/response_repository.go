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

// ResponseRepository はPostgreSQLを用いたResponse(および配下のResponseItem)の
// 永続化を行う。
type ResponseRepository struct {
	conn *sql.DB
}

// NewResponseRepository はResponseRepositoryを生成する。
func NewResponseRepository(conn *sql.DB) *ResponseRepository {
	return &ResponseRepository{conn: conn}
}

// Create はResponseとその配下のResponseItemをまとめて新規保存する。
func (r *ResponseRepository) Create(ctx context.Context, response *domain.Response) error {
	executor := executorFromContext(ctx, r.conn)

	var rawUserID *string
	if response.UserID() != nil {
		v := response.UserID().String()
		rawUserID = &v
	}

	_, err := executor.ExecContext(ctx, `
		INSERT INTO responses (id, information_id, user_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`,
		response.ID().String(),
		response.InformationID().String(),
		rawUserID,
		response.CreatedAt(),
		response.UpdatedAt(),
	)
	if err != nil {
		return fmt.Errorf("insert response: %w", err)
	}

	for _, item := range response.Items() {
		var rawOptionID *string
		if item.OptionID() != nil {
			v := item.OptionID().String()
			rawOptionID = &v
		}

		_, err := executor.ExecContext(ctx, `
			INSERT INTO response_items (id, response_id, source_id, option_id, value, created_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`,
			item.ID().String(),
			item.ResponseID().String(),
			item.SourceID().String(),
			rawOptionID,
			item.Value(),
			item.CreatedAt(),
		)
		if err != nil {
			return fmt.Errorf("insert response item: %w", err)
		}
	}

	return nil
}

// GetByInformationAndUser はinformationID・userIDに対応するResponseを取得する。
func (r *ResponseRepository) GetByInformationAndUser(ctx context.Context, informationID, userID uuid.UUID) (*domain.Response, error) {
	row := executorFromContext(ctx, r.conn).QueryRowContext(ctx, `
		SELECT id, information_id, user_id, created_at, updated_at
		FROM responses
		WHERE information_id = $1 AND user_id = $2
	`, informationID.String(), userID.String())

	response, err := scanResponseRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	items, err := r.listItems(ctx, response.ID())
	if err != nil {
		return nil, err
	}

	return domain.ReconstructResponse(response.ID(), response.InformationID(), response.UserID(), items, response.CreatedAt(), response.UpdatedAt()), nil
}

// ListByInformationID はinformationIDに対応するResponseを一覧取得する。
func (r *ResponseRepository) ListByInformationID(ctx context.Context, informationID uuid.UUID) ([]*domain.Response, error) {
	rows, err := executorFromContext(ctx, r.conn).QueryContext(ctx, `
		SELECT id, information_id, user_id, created_at, updated_at
		FROM responses
		WHERE information_id = $1
		ORDER BY created_at
	`, informationID.String())
	if err != nil {
		return nil, fmt.Errorf("select responses: %w", err)
	}
	defer rows.Close()

	var responses []*domain.Response
	for rows.Next() {
		response, err := scanResponseRow(rows)
		if err != nil {
			return nil, err
		}
		responses = append(responses, response)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate responses: %w", err)
	}

	result := make([]*domain.Response, 0, len(responses))
	for _, response := range responses {
		items, err := r.listItems(ctx, response.ID())
		if err != nil {
			return nil, err
		}
		result = append(result, domain.ReconstructResponse(response.ID(), response.InformationID(), response.UserID(), items, response.CreatedAt(), response.UpdatedAt()))
	}

	return result, nil
}

func (r *ResponseRepository) listItems(ctx context.Context, responseID uuid.UUID) ([]domain.ResponseItem, error) {
	rows, err := executorFromContext(ctx, r.conn).QueryContext(ctx, `
		SELECT id, response_id, source_id, option_id, value, created_at
		FROM response_items
		WHERE response_id = $1
		ORDER BY created_at
	`, responseID.String())
	if err != nil {
		return nil, fmt.Errorf("select response items: %w", err)
	}
	defer rows.Close()

	var items []domain.ResponseItem
	for rows.Next() {
		var (
			rawID         string
			rawResponseID string
			rawSourceID   string
			rawOptionID   sql.NullString
			value         sql.NullString
			createdAt     time.Time
		)
		if err := rows.Scan(&rawID, &rawResponseID, &rawSourceID, &rawOptionID, &value, &createdAt); err != nil {
			return nil, fmt.Errorf("scan response item: %w", err)
		}

		id, err := uuid.Parse(rawID)
		if err != nil {
			return nil, fmt.Errorf("parse response item id: %w", err)
		}
		responseID, err := uuid.Parse(rawResponseID)
		if err != nil {
			return nil, fmt.Errorf("parse response id: %w", err)
		}
		sourceID, err := uuid.Parse(rawSourceID)
		if err != nil {
			return nil, fmt.Errorf("parse source id: %w", err)
		}

		var optionID *uuid.UUID
		if rawOptionID.Valid {
			parsed, err := uuid.Parse(rawOptionID.String)
			if err != nil {
				return nil, fmt.Errorf("parse option id: %w", err)
			}
			optionID = &parsed
		}

		var itemValue *string
		if value.Valid {
			v := value.String
			itemValue = &v
		}

		items = append(items, *domain.ReconstructResponseItem(id, responseID, sourceID, optionID, itemValue, createdAt))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate response items: %w", err)
	}

	return items, nil
}

func scanResponseRow(scanner rowScanner) (*domain.Response, error) {
	var (
		rawID            string
		rawInformationID string
		rawUserID        sql.NullString
		createdAt        time.Time
		updatedAt        time.Time
	)
	if err := scanner.Scan(&rawID, &rawInformationID, &rawUserID, &createdAt, &updatedAt); err != nil {
		return nil, fmt.Errorf("scan response: %w", err)
	}

	id, err := uuid.Parse(rawID)
	if err != nil {
		return nil, fmt.Errorf("parse response id: %w", err)
	}
	informationID, err := uuid.Parse(rawInformationID)
	if err != nil {
		return nil, fmt.Errorf("parse information id: %w", err)
	}

	var userID *uuid.UUID
	if rawUserID.Valid {
		parsed, err := uuid.Parse(rawUserID.String)
		if err != nil {
			return nil, fmt.Errorf("parse user id: %w", err)
		}
		userID = &parsed
	}

	return domain.ReconstructResponse(id, informationID, userID, nil, createdAt, updatedAt), nil
}
