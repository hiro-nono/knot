package repository

import (
	"context"

	"knot-api/internal/domain"

	"uuid"
)

// ResponseRepository はResponse(および配下のResponseItem)の永続化を抽象化する。
type ResponseRepository interface {
	// Create はResponseとその配下のResponseItemをまとめて新規保存する。
	Create(ctx context.Context, response *domain.Response) error
	// GetByInformationAndUser はinformationID・userIDに対応するResponseを取得する。
	// 見つからない場合はdomain.ErrNotFoundを返す。
	GetByInformationAndUser(ctx context.Context, informationID, userID uuid.UUID) (*domain.Response, error)
	// ListByInformationID はinformationIDに対応するResponseを一覧取得する
	// (発信者が受信者の返答を確認する用途)。
	ListByInformationID(ctx context.Context, informationID uuid.UUID) ([]*domain.Response, error)
}
