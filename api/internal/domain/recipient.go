package domain

import (
	"errors"
	"time"

	"uuid"
)

// Recipient はInformation(accessType=restricted)を閲覧できるUserを表すEntity。
// InformationとUserの多対多の関係を表す中間的なレコードで、
// 「このUserはこのInformationを閲覧してよい」という許可のみを保持する。
type Recipient struct {
	id            uuid.UUID
	informationID uuid.UUID
	userID        uuid.UUID
	createdAt     time.Time
}

// NewRecipient は新しいRecipientを生成する。
func NewRecipient(informationID, userID uuid.UUID) (*Recipient, error) {
	if informationID == uuid.Nil() {
		return nil, errors.New("informationID は必須です")
	}
	if userID == uuid.Nil() {
		return nil, errors.New("userID は必須です")
	}

	return &Recipient{
		id:            uuid.New(),
		informationID: informationID,
		userID:        userID,
		createdAt:     time.Now(),
	}, nil
}

// ReconstructRecipient は永続化されたデータからRecipientを復元する。
// バリデーションは行わない。
func ReconstructRecipient(id, informationID, userID uuid.UUID, createdAt time.Time) *Recipient {
	return &Recipient{
		id:            id,
		informationID: informationID,
		userID:        userID,
		createdAt:     createdAt,
	}
}

func (r *Recipient) ID() uuid.UUID {
	return r.id
}

func (r *Recipient) InformationID() uuid.UUID {
	return r.informationID
}

func (r *Recipient) UserID() uuid.UUID {
	return r.userID
}

func (r *Recipient) CreatedAt() time.Time {
	return r.createdAt
}
