package domain

import (
	"errors"
	"time"

	"uuid"
)

// MembershipEventType はMembershipの状態更新イベントの種別を表す。
type MembershipEventType string

const (
	// MembershipEventCreated はMembership作成時のイベント。
	MembershipEventCreated MembershipEventType = "created"
	// MembershipEventRoleChanged はroleが変更された時のイベント。
	MembershipEventRoleChanged MembershipEventType = "role_changed"
	// MembershipEventRemoved はAccountから除外(論理削除)された時のイベント。
	MembershipEventRemoved MembershipEventType = "removed"
	// MembershipEventReactivated はremoved状態からactiveへ戻された時のイベント。
	MembershipEventReactivated MembershipEventType = "reactivated"
)

func (t MembershipEventType) valid() bool {
	switch t {
	case MembershipEventCreated, MembershipEventRoleChanged, MembershipEventRemoved, MembershipEventReactivated:
		return true
	default:
		return false
	}
}

// MembershipEvent はMembershipに対して何が起きたかを記録するEntity(履歴、追記のみ)。
// Membershipの「現在の状態」を表すものではなく、変化の記録そのものを表す。
type MembershipEvent struct {
	id           uuid.UUID
	membershipID uuid.UUID
	eventType    MembershipEventType
	createdAt    time.Time
}

// NewMembershipEvent は新しいMembershipEventを生成する。
func NewMembershipEvent(membershipID uuid.UUID, eventType MembershipEventType) (*MembershipEvent, error) {
	if membershipID == uuid.Nil() {
		return nil, errors.New("membershipID は必須です")
	}
	if !eventType.valid() {
		return nil, errors.New("event_type の値が不正です")
	}

	return &MembershipEvent{
		id:           uuid.New(),
		membershipID: membershipID,
		eventType:    eventType,
		createdAt:    time.Now(),
	}, nil
}

// ReconstructMembershipEvent は永続化されたデータからMembershipEventを復元する。
// バリデーションは行わない。
func ReconstructMembershipEvent(id, membershipID uuid.UUID, eventType MembershipEventType, createdAt time.Time) *MembershipEvent {
	return &MembershipEvent{
		id:           id,
		membershipID: membershipID,
		eventType:    eventType,
		createdAt:    createdAt,
	}
}

func (e *MembershipEvent) ID() uuid.UUID {
	return e.id
}

func (e *MembershipEvent) MembershipID() uuid.UUID {
	return e.membershipID
}

func (e *MembershipEvent) EventType() MembershipEventType {
	return e.eventType
}

func (e *MembershipEvent) CreatedAt() time.Time {
	return e.createdAt
}
