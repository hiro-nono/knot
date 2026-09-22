package repository

import (
	"context"

	"knot-api/internal/domain"

	"uuid"
)

// InformationRepository はInformationの永続化を抽象化する。
type InformationRepository interface {
	// Create はInformationを新規保存する。
	Create(ctx context.Context, information *domain.Information) error
	// Get はIDを指定してInformationを1件取得する。
	// 見つからない場合はdomain.ErrNotFoundを返す。
	Get(ctx context.Context, id uuid.UUID) (*domain.Information, error)
	// ListByAccountID は指定したAccountが所有するInformationを、
	// 作成日時の新しい順に一覧取得する。
	ListByAccountID(ctx context.Context, accountID uuid.UUID) ([]*domain.Information, error)
}
