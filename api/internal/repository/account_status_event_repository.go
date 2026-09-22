package repository

import (
	"context"

	"knot-api/internal/domain"
)

// AccountStatusEventRepository はAccountStatusEvent(ステータス更新履歴)の
// 永続化を抽象化する。履歴は追記のみで、更新・削除は行わない。
type AccountStatusEventRepository interface {
	// Create はAccountStatusEventを新規保存する。
	Create(ctx context.Context, event *domain.AccountStatusEvent) error
}
