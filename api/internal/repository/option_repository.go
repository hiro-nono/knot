package repository

import (
	"context"

	"knot-api/internal/domain"

	"uuid"
)

// OptionRepository はOptionの永続化を抽象化する。
type OptionRepository interface {
	// Create はOptionを新規保存する。
	Create(ctx context.Context, option *domain.Option) error
	// List は指定したSourceに属するOptionを一覧取得する。
	List(ctx context.Context, sourceID uuid.UUID) ([]*domain.Option, error)
}
