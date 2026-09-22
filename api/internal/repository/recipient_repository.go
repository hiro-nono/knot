package repository

import (
	"context"

	"knot-api/internal/domain"

	"uuid"
)

// RecipientRepository はRecipient(restrictedなInformationを閲覧できるUser)の
// 永続化を抽象化する。
type RecipientRepository interface {
	// Create はRecipientを新規保存する。
	Create(ctx context.Context, recipient *domain.Recipient) error
	// Exists はuserIDがinformationIDのRecipientとして登録されているかを返す。
	Exists(ctx context.Context, informationID, userID uuid.UUID) (bool, error)
	// ListByInformationID はinformationIDに対応するRecipientを一覧取得する。
	ListByInformationID(ctx context.Context, informationID uuid.UUID) ([]*domain.Recipient, error)
	// Delete はinformationID・userIDに対応するRecipientを削除する。
	// 見つからない場合はdomain.ErrNotFoundを返す。
	Delete(ctx context.Context, informationID, userID uuid.UUID) error
}
