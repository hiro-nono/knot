package domain

import (
	"errors"
	"time"

	"uuid"
)

// Response はInformationに対するUserの返答を表す集約ルートのEntity。
// ResponseItemを子として保持する。
//
// 設計上の責務分離:
//   - Recipientは「Informationを閲覧できるか」という許可を表す。
//   - Responseは「Informationに対する返答」そのものを表す。
//   - ResponseはRecipientそのものには紐付けない(recipientsテーブルへの
//     参照を持たない)。
//   - 返答できるかどうかは、RecipientではなくInformationのaccess_type・
//     response_policy・recipientsによってusecase層が判定する。
//
// userIDはnullable。PUBLIC+ANONYMOUSなInformationは未ログインでも回答できる
// (Google Formのように、誰が回答したかを記録しない)ため、userIDがnilの
// Responseが存在しうる。userIDが必須かどうかはInformation側の設定
// (access_type・response_policy)に基づきusecase層が判定する。
type Response struct {
	id            uuid.UUID
	informationID uuid.UUID
	userID        *uuid.UUID
	items         []ResponseItem
	createdAt     time.Time
	updatedAt     time.Time
}

// NewResponse は新しいResponseを生成する。userIDはnil(匿名回答)を許容する。
func NewResponse(informationID uuid.UUID, userID *uuid.UUID) (*Response, error) {
	if informationID == uuid.Nil() {
		return nil, errors.New("informationID は必須です")
	}
	if userID != nil && *userID == uuid.Nil() {
		return nil, errors.New("userID を指定する場合はNilにできません")
	}

	now := time.Now()
	return &Response{
		id:            uuid.New(),
		informationID: informationID,
		userID:        userID,
		items:         []ResponseItem{},
		createdAt:     now,
		updatedAt:     now,
	}, nil
}

// ReconstructResponse は永続化されたデータからResponseを復元する。
// 呼び出し元(repository)がすでに正当なデータであることを保証するため、
// バリデーションは行わない。
func ReconstructResponse(id, informationID uuid.UUID, userID *uuid.UUID, items []ResponseItem, createdAt, updatedAt time.Time) *Response {
	if items == nil {
		items = []ResponseItem{}
	}

	return &Response{
		id:            id,
		informationID: informationID,
		userID:        userID,
		items:         items,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
	}
}

func (r *Response) ID() uuid.UUID {
	return r.id
}

func (r *Response) InformationID() uuid.UUID {
	return r.informationID
}

// UserID は回答したUserのIDを返す。匿名回答の場合はnilを返す。
func (r *Response) UserID() *uuid.UUID {
	return r.userID
}

func (r *Response) Items() []ResponseItem {
	return r.items
}

func (r *Response) CreatedAt() time.Time {
	return r.createdAt
}

func (r *Response) UpdatedAt() time.Time {
	return r.updatedAt
}

// AddItem はこのResponseにResponseItemを追加する。
// 追加するItemは、このResponseのIDを参照している必要がある。
func (r *Response) AddItem(item ResponseItem) error {
	if item.ResponseID() != r.id {
		return errors.New("item は指定されたresponseに属していません")
	}

	r.items = append(r.items, item)
	r.updatedAt = time.Now()
	return nil
}
