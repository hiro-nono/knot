package domain

import (
	"errors"
	"time"

	"uuid"
)

// AccountStatusEventType はAccountのステータス更新イベントの種別を表す。
// AccountStatusの5値に加え、凍結解除・再有効化のイベントを持つ。
type AccountStatusEventType string

const (
	// AccountStatusEventCreated はAccount作成時のイベント。
	AccountStatusEventCreated   AccountStatusEventType = "created"
	AccountStatusEventFrozen    AccountStatusEventType = "frozen"
	AccountStatusEventSuspended AccountStatusEventType = "suspended"
	AccountStatusEventWithdrawn AccountStatusEventType = "withdrawn"
	AccountStatusEventBanned    AccountStatusEventType = "banned"
	// AccountStatusEventUnfrozen は凍結解除(frozen -> active)のイベント。
	AccountStatusEventUnfrozen AccountStatusEventType = "unfrozen"
	// AccountStatusEventReactivated は再有効化(suspended/banned -> active)のイベント。
	AccountStatusEventReactivated AccountStatusEventType = "reactivated"
)

func (t AccountStatusEventType) valid() bool {
	switch t {
	case AccountStatusEventCreated, AccountStatusEventFrozen, AccountStatusEventSuspended,
		AccountStatusEventWithdrawn, AccountStatusEventBanned,
		AccountStatusEventUnfrozen, AccountStatusEventReactivated:
		return true
	default:
		return false
	}
}

// AccountStatusEvent はAccountのステータス更新イベント1件を表すEntity(履歴)。
type AccountStatusEvent struct {
	id        uuid.UUID
	accountID uuid.UUID
	eventType AccountStatusEventType
	createdAt time.Time
}

// NewAccountStatusEvent は新しいAccountStatusEventを生成する。
func NewAccountStatusEvent(accountID uuid.UUID, eventType AccountStatusEventType) (*AccountStatusEvent, error) {
	if accountID == uuid.Nil() {
		return nil, errors.New("accountID は必須です")
	}
	if !eventType.valid() {
		return nil, errors.New("event_type の値が不正です")
	}

	return &AccountStatusEvent{
		id:        uuid.New(),
		accountID: accountID,
		eventType: eventType,
		createdAt: time.Now(),
	}, nil
}

// ReconstructAccountStatusEvent は永続化されたデータからAccountStatusEventを復元する。
// バリデーションは行わない。
func ReconstructAccountStatusEvent(id uuid.UUID, accountID uuid.UUID, eventType AccountStatusEventType, createdAt time.Time) *AccountStatusEvent {
	return &AccountStatusEvent{
		id:        id,
		accountID: accountID,
		eventType: eventType,
		createdAt: createdAt,
	}
}

func (e *AccountStatusEvent) ID() uuid.UUID {
	return e.id
}

func (e *AccountStatusEvent) AccountID() uuid.UUID {
	return e.accountID
}

func (e *AccountStatusEvent) EventType() AccountStatusEventType {
	return e.eventType
}

func (e *AccountStatusEvent) CreatedAt() time.Time {
	return e.createdAt
}
