package repository

import (
	"context"

	"knot-api/internal/domain"

	"uuid"
)

// PreferenceRepository はPreferenceの永続化を抽象化する。
type PreferenceRepository interface {
	// Get はuserIDに対応するPreferenceを取得する。
	// 見つからない場合はdomain.ErrNotFoundを返す。
	Get(ctx context.Context, userID uuid.UUID) (*domain.Preference, error)
	// Save はPreferenceを保存する(存在すれば更新、無ければ新規作成)。
	Save(ctx context.Context, preference *domain.Preference) error
}
