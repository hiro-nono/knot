package domain

import (
	"errors"
	"time"
	"uuid"
)

// SourceType はCLAUDE.mdで定義されたSource of Truthの分類を表す。
type SourceType string

const (
	SourceTypeIdentity    SourceType = "identity"
	SourceTypeFact        SourceType = "fact"
	SourceTypeSchedule    SourceType = "schedule"
	SourceTypeCondition   SourceType = "condition"
	SourceTypeInteraction SourceType = "interaction"
)

func (t SourceType) valid() bool {
	switch t {
	case SourceTypeIdentity, SourceTypeFact, SourceTypeSchedule, SourceTypeCondition, SourceTypeInteraction:
		return true
	default:
		return false
	}
}

// SourceStatus はSourceの値がどの程度確定しているかを表す。
type SourceStatus string

const (
	SourceStatusConfirmed SourceStatus = "confirmed"
	SourceStatusUndecided SourceStatus = "undecided"
	SourceStatusUnknown   SourceStatus = "unknown"
)

func (s SourceStatus) valid() bool {
	switch s {
	case SourceStatusConfirmed, SourceStatusUndecided, SourceStatusUnknown:
		return true
	default:
		return false
	}
}

// SourceInteractionType は受信者がSourceに対してどう応答すべきかを表す。
// 受信者の応答を必要としないSourceの場合はnilになる。
type SourceInteractionType string

const (
	SourceInteractionTypeRadio SourceInteractionType = "radio"
	SourceInteractionTypeCheck SourceInteractionType = "check"
	SourceInteractionTypeText  SourceInteractionType = "text"
)

func (t SourceInteractionType) valid() bool {
	switch t {
	case SourceInteractionTypeRadio, SourceInteractionTypeCheck, SourceInteractionTypeText:
		return true
	default:
		return false
	}
}

// selectable はOptionの追加を許可するinteraction_typeかどうかを表す。
func (t SourceInteractionType) selectable() bool {
	return t == SourceInteractionTypeRadio || t == SourceInteractionTypeCheck
}

// Source はInformationに属し、構造化された情報を1件表すEntity。
type Source struct {
	id              uuid.UUID
	informationID   uuid.UUID
	sourceType      SourceType
	key             string
	value           string
	status          SourceStatus
	interactionType *SourceInteractionType
	options         []Option
	createdAt       time.Time
	updatedAt       time.Time
}

// NewSource は新しいSourceを生成する。
func NewSource(
	informationID uuid.UUID,
	sourceType SourceType,
	key string,
	value string,
	status SourceStatus,
	interactionType *SourceInteractionType,
) (*Source, error) {
	if informationID == uuid.Nil() {
		return nil, errors.New("informationID は必須です")
	}
	if !sourceType.valid() {
		return nil, errors.New("type の値が不正です")
	}
	if key == "" {
		return nil, errors.New("key は必須です")
	}
	if value == "" {
		return nil, errors.New("value は必須です")
	}
	if !status.valid() {
		return nil, errors.New("status の値が不正です")
	}
	if interactionType != nil && !interactionType.valid() {
		return nil, errors.New("interaction_type の値が不正です")
	}

	now := time.Now()
	return &Source{
		id:              uuid.New(),
		informationID:   informationID,
		sourceType:      sourceType,
		key:             key,
		value:           value,
		status:          status,
		interactionType: interactionType,
		options:         []Option{},
		createdAt:       now,
		updatedAt:       now,
	}, nil
}

// ReconstructSource は永続化されたデータからSourceを復元する。
// 呼び出し元(repository)がすでに正当なデータであることを保証するため、
// バリデーションは行わない。
func ReconstructSource(
	id uuid.UUID,
	informationID uuid.UUID,
	sourceType SourceType,
	key string,
	value string,
	status SourceStatus,
	interactionType *SourceInteractionType,
	options []Option,
	createdAt time.Time,
	updatedAt time.Time,
) *Source {
	if options == nil {
		options = []Option{}
	}

	return &Source{
		id:              id,
		informationID:   informationID,
		sourceType:      sourceType,
		key:             key,
		value:           value,
		status:          status,
		interactionType: interactionType,
		options:         options,
		createdAt:       createdAt,
		updatedAt:       updatedAt,
	}
}

func (s *Source) ID() uuid.UUID {
	return s.id
}

func (s *Source) InformationID() uuid.UUID {
	return s.informationID
}

func (s *Source) Type() SourceType {
	return s.sourceType
}

func (s *Source) Key() string {
	return s.key
}

func (s *Source) Value() string {
	return s.value
}

func (s *Source) Status() SourceStatus {
	return s.status
}

func (s *Source) InteractionType() *SourceInteractionType {
	return s.interactionType
}

func (s *Source) Options() []Option {
	return s.options
}

func (s *Source) CreatedAt() time.Time {
	return s.createdAt
}

func (s *Source) UpdatedAt() time.Time {
	return s.updatedAt
}

// AddOption はこのSourceにOptionを追加する。
// interaction_typeがradioまたはcheckのSourceにのみ追加できる。
func (s *Source) AddOption(option Option) error {
	if s.interactionType == nil || !s.interactionType.selectable() {
		return errors.New("radio, check以外のinteraction_typeにはoptionを追加できません")
	}
	if option.SourceID() != s.id {
		return errors.New("option は指定されたsourceに属していません")
	}

	s.options = append(s.options, option)
	s.updatedAt = time.Now()
	return nil
}
