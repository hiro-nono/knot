package repository

import (
	"context"

	"knot-api/internal/domain"

	"uuid"
)

// SourceRepository はSourceの永続化を抽象化する。
type SourceRepository interface {
	// Create はSourceを新規保存する。
	Create(ctx context.Context, source *domain.Source) error
	// List は指定したInformationに属するSourceを一覧取得する。
	List(ctx context.Context, informationID uuid.UUID) ([]*domain.Source, error)
}
