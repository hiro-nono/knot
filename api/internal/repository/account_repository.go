package repository

import (
	"context"

	"knot-api/internal/domain"

	"uuid"
)

// AccountRepository はAccountの永続化を抽象化する。
type AccountRepository interface {
	// Create はAccountを新規保存する。
	Create(ctx context.Context, account *domain.Account) error
	// Get はIDを指定してAccountを1件取得する。
	// 見つからない場合はdomain.ErrNotFoundを返す。
	Get(ctx context.Context, id uuid.UUID) (*domain.Account, error)
	// GetByProviderID はProviderID(外部IDプロバイダのユーザーID)を指定して
	// Accountを1件取得する。見つからない場合はdomain.ErrNotFoundを返す。
	GetByProviderID(ctx context.Context, providerID string) (*domain.Account, error)
	// Update はAccountの内容(role, status等)を更新する。
	Update(ctx context.Context, account *domain.Account) error
	// ListByStatus は指定したstatusのAccountを一覧取得する。
	ListByStatus(ctx context.Context, status domain.AccountStatus) ([]*domain.Account, error)
}
