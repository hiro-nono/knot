package domain

import (
	"errors"
	"time"
	"uuid"
)

// Option はSourceに属し、選択肢を1件表すEntity。
type Option struct {
	id        uuid.UUID
	sourceID  uuid.UUID
	value     string
	sortOrder int
	createdAt time.Time
	updatedAt time.Time
}

// NewOption は新しいOptionを生成する。
func NewOption(sourceID uuid.UUID, value string, sortOrder int) (*Option, error) {
	if sourceID == uuid.Nil() {
		return nil, errors.New("sourceID は必須です")
	}
	if value == "" {
		return nil, errors.New("value は必須です")
	}
	if sortOrder < 0 {
		return nil, errors.New("sort_order は0以上で指定してください")
	}

	now := time.Now()
	return &Option{
		id:        uuid.New(),
		sourceID:  sourceID,
		value:     value,
		sortOrder: sortOrder,
		createdAt: now,
		updatedAt: now,
	}, nil
}

// ReconstructOption は永続化されたデータからOptionを復元する。
// 呼び出し元(repository)がすでに正当なデータであることを保証するため、
// バリデーションは行わない。
func ReconstructOption(id uuid.UUID, sourceID uuid.UUID, value string, sortOrder int, createdAt, updatedAt time.Time) *Option {
	return &Option{
		id:        id,
		sourceID:  sourceID,
		value:     value,
		sortOrder: sortOrder,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

func (o *Option) ID() uuid.UUID {
	return o.id
}

func (o *Option) SourceID() uuid.UUID {
	return o.sourceID
}

func (o *Option) Value() string {
	return o.value
}

func (o *Option) SortOrder() int {
	return o.sortOrder
}

func (o *Option) CreatedAt() time.Time {
	return o.createdAt
}

func (o *Option) UpdatedAt() time.Time {
	return o.updatedAt
}
