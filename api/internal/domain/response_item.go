package domain

import (
	"errors"
	"time"

	"uuid"
)

// ResponseItem はResponseに属し、1つのSourceに対する回答を1件表すEntity。
// optionID・valueはともにnullableであり、どちらか一方のみを設定する。
// Sourceのinteraction_typeがradio・checkの場合はoptionIDを、
// textの場合はvalueを設定する(どのSourceに対する回答かに応じてどちらを
// 使うべきかを判定するのはusecase層の責務であり、本Entityは
// 「両方は不可、片方は必須」という形だけを保証する)。
type ResponseItem struct {
	id         uuid.UUID
	responseID uuid.UUID
	sourceID   uuid.UUID
	optionID   *uuid.UUID
	value      *string
	createdAt  time.Time
}

// NewResponseItem は新しいResponseItemを生成する。
func NewResponseItem(responseID, sourceID uuid.UUID, optionID *uuid.UUID, value *string) (*ResponseItem, error) {
	if responseID == uuid.Nil() {
		return nil, errors.New("responseID は必須です")
	}
	if sourceID == uuid.Nil() {
		return nil, errors.New("sourceID は必須です")
	}
	if optionID == nil && value == nil {
		return nil, errors.New("option_id または value のいずれかが必須です")
	}
	if optionID != nil && value != nil {
		return nil, errors.New("option_id と value を同時に設定することはできません")
	}

	return &ResponseItem{
		id:         uuid.New(),
		responseID: responseID,
		sourceID:   sourceID,
		optionID:   optionID,
		value:      value,
		createdAt:  time.Now(),
	}, nil
}

// ReconstructResponseItem は永続化されたデータからResponseItemを復元する。
// バリデーションは行わない。
func ReconstructResponseItem(id, responseID, sourceID uuid.UUID, optionID *uuid.UUID, value *string, createdAt time.Time) *ResponseItem {
	return &ResponseItem{
		id:         id,
		responseID: responseID,
		sourceID:   sourceID,
		optionID:   optionID,
		value:      value,
		createdAt:  createdAt,
	}
}

func (i *ResponseItem) ID() uuid.UUID {
	return i.id
}

func (i *ResponseItem) ResponseID() uuid.UUID {
	return i.responseID
}

func (i *ResponseItem) SourceID() uuid.UUID {
	return i.sourceID
}

func (i *ResponseItem) OptionID() *uuid.UUID {
	return i.optionID
}

func (i *ResponseItem) Value() *string {
	return i.value
}

func (i *ResponseItem) CreatedAt() time.Time {
	return i.createdAt
}
