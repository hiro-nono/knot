package repository

import (
	"context"

	"knot-api/internal/domain"

	"uuid"
)

// UserRepository はUserの永続化を抽象化する。
type UserRepository interface {
	// Create はUserを新規保存する。
	Create(ctx context.Context, user *domain.User) error
	// Get はIDを指定してUserを1件取得する。
	// 見つからない場合はdomain.ErrNotFoundを返す。
	Get(ctx context.Context, id uuid.UUID) (*domain.User, error)
	// Update はUserのプロフィールを更新する。
	Update(ctx context.Context, user *domain.User) error
}
