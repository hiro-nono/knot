package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"knot-api/internal/domain"

	"uuid"
)

// SourceRepository はPostgreSQLを用いたSourceの永続化を行う。
type SourceRepository struct {
	conn *sql.DB
}

// NewSourceRepository はSourceRepositoryを生成する。
func NewSourceRepository(conn *sql.DB) *SourceRepository {
	return &SourceRepository{conn: conn}
}

// Create はSourceを新規保存する。
func (r *SourceRepository) Create(ctx context.Context, source *domain.Source) error {
	var interactionType *string
	if source.InteractionType() != nil {
		v := string(*source.InteractionType())
		interactionType = &v
	}

	_, err := executorFromContext(ctx, r.conn).ExecContext(ctx, `
		INSERT INTO sources (id, information_id, type, key, value, status, interaction_type, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`,
		source.ID().String(),
		source.InformationID().String(),
		string(source.Type()),
		source.Key(),
		source.Value(),
		string(source.Status()),
		interactionType,
		source.CreatedAt(),
		source.UpdatedAt(),
	)
	if err != nil {
		return fmt.Errorf("insert source: %w", err)
	}

	return nil
}

// List は指定したInformationに属するSourceを一覧取得する。
func (r *SourceRepository) List(ctx context.Context, informationID uuid.UUID) ([]*domain.Source, error) {
	rows, err := executorFromContext(ctx, r.conn).QueryContext(ctx, `
		SELECT id, information_id, type, key, value, status, interaction_type, created_at, updated_at
		FROM sources
		WHERE information_id = $1
		ORDER BY created_at
	`, informationID.String())
	if err != nil {
		return nil, fmt.Errorf("select sources: %w", err)
	}
	defer rows.Close()

	var sources []*domain.Source
	for rows.Next() {
		var (
			rawID            string
			rawInformationID string
			sourceType       string
			key              string
			value            string
			status           string
			interactionType  sql.NullString
			createdAt        time.Time
			updatedAt        time.Time
		)
		if err := rows.Scan(&rawID, &rawInformationID, &sourceType, &key, &value, &status, &interactionType, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("scan source: %w", err)
		}

		id, err := uuid.Parse(rawID)
		if err != nil {
			return nil, fmt.Errorf("parse source id: %w", err)
		}
		infoID, err := uuid.Parse(rawInformationID)
		if err != nil {
			return nil, fmt.Errorf("parse information id: %w", err)
		}

		var it *domain.SourceInteractionType
		if interactionType.Valid {
			v := domain.SourceInteractionType(interactionType.String)
			it = &v
		}

		sources = append(sources, domain.ReconstructSource(
			id,
			infoID,
			domain.SourceType(sourceType),
			key,
			value,
			domain.SourceStatus(status),
			it,
			nil,
			createdAt,
			updatedAt,
		))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate sources: %w", err)
	}

	return sources, nil
}
