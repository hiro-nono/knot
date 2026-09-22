package domain

import (
	"errors"
	"fmt"
	"time"

	"uuid"
)

// MembershipRole はMembershipの権限区分を表す。ownerが最上位、続いてadmin、memberの順。
type MembershipRole string

const (
	MembershipRoleOwner  MembershipRole = "owner"
	MembershipRoleAdmin  MembershipRole = "admin"
	MembershipRoleMember MembershipRole = "member"
)

func (r MembershipRole) valid() bool {
	switch r {
	case MembershipRoleOwner, MembershipRoleAdmin, MembershipRoleMember:
		return true
	default:
		return false
	}
}

// MembershipStatus はMembershipの状態を表す。
type MembershipStatus string

const (
	// MembershipStatusActive はUserがAccount(Organization)に所属している状態を表す。
	MembershipStatusActive MembershipStatus = "active"
	// MembershipStatusRemoved はUserがAccount(Organization)から除外された
	// (論理削除された)状態を表す。
	MembershipStatusRemoved MembershipStatus = "removed"
)

func (s MembershipStatus) valid() bool {
	switch s {
	case MembershipStatusActive, MembershipStatusRemoved:
		return true
	default:
		return false
	}
}

// Membership はUserがどのAccount(Organization)に所属しているか、および
// そのAccount内での権限(role)・状態(status)を表すEntity。
//
// UserとAccountの所属関係はMembershipのみが管理する(Userはaccount_idを
// 持たない)。同一Account・同一UserのMembershipは常に1件のみ存在し
// (再追加時は既存Membershipを再利用してactiveへ戻す)、履歴は
// MembershipEventが一貫して記録する。
type Membership struct {
	id        uuid.UUID
	accountID uuid.UUID
	userID    uuid.UUID
	role      MembershipRole
	status    MembershipStatus
	createdAt time.Time
	updatedAt time.Time
}

// NewMembership は新しいMembershipを生成する。常にactive状態で作成される。
func NewMembership(accountID, userID uuid.UUID, role MembershipRole) (*Membership, error) {
	if accountID == uuid.Nil() {
		return nil, errors.New("accountID は必須です")
	}
	if userID == uuid.Nil() {
		return nil, errors.New("userID は必須です")
	}
	if !role.valid() {
		return nil, fmt.Errorf("role の値が不正です: %q", role)
	}

	now := time.Now()
	return &Membership{
		id:        uuid.New(),
		accountID: accountID,
		userID:    userID,
		role:      role,
		status:    MembershipStatusActive,
		createdAt: now,
		updatedAt: now,
	}, nil
}

// ReconstructMembership は永続化されたデータからMembershipを復元する。
// バリデーションは行わない。
func ReconstructMembership(id, accountID, userID uuid.UUID, role MembershipRole, status MembershipStatus, createdAt, updatedAt time.Time) *Membership {
	return &Membership{
		id:        id,
		accountID: accountID,
		userID:    userID,
		role:      role,
		status:    status,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

func (m *Membership) ID() uuid.UUID {
	return m.id
}

func (m *Membership) AccountID() uuid.UUID {
	return m.accountID
}

func (m *Membership) UserID() uuid.UUID {
	return m.userID
}

func (m *Membership) Role() MembershipRole {
	return m.role
}

func (m *Membership) Status() MembershipStatus {
	return m.status
}

func (m *Membership) IsOwner() bool {
	return m.role == MembershipRoleOwner
}

func (m *Membership) IsAdmin() bool {
	return m.role == MembershipRoleAdmin
}

// IsActive はUserが現在Accountに所属しているか(除外されていないか)を返す。
func (m *Membership) IsActive() bool {
	return m.status == MembershipStatusActive
}

func (m *Membership) CreatedAt() time.Time {
	return m.createdAt
}

func (m *Membership) UpdatedAt() time.Time {
	return m.updatedAt
}

// ChangeRole はMembershipのroleを変更する。owner権限の付与・剥奪はこの操作では
// 扱わない(オーナーシップの委譲は別の関心事であり、本Entityの対象外)。
func (m *Membership) ChangeRole(role MembershipRole) (MembershipEventType, error) {
	if !m.IsActive() {
		return "", errors.New("removed状態のmembershipのroleは変更できません")
	}
	if !role.valid() {
		return "", fmt.Errorf("role の値が不正です: %q", role)
	}
	if m.role == MembershipRoleOwner || role == MembershipRoleOwner {
		return "", errors.New("owner権限の変更はこの操作では行えません")
	}
	if m.role == role {
		return "", fmt.Errorf("既にrole %qです", role)
	}

	m.role = role
	m.updatedAt = time.Now()
	return MembershipEventRoleChanged, nil
}

// Remove はMembershipをremoved状態にする(Accountからの除外・論理削除)。
// ownerのMembershipは除外できない(オーナー不在のOrganizationを防ぐため)。
func (m *Membership) Remove() (MembershipEventType, error) {
	if m.role == MembershipRoleOwner {
		return "", errors.New("ownerのmembershipを削除することはできません")
	}
	if m.status == MembershipStatusRemoved {
		return "", errors.New("既にremoved状態です")
	}

	m.status = MembershipStatusRemoved
	m.updatedAt = time.Now()
	return MembershipEventRemoved, nil
}

// Reactivate はremoved状態のMembershipをactiveへ戻す(Userの再追加)。
func (m *Membership) Reactivate() (MembershipEventType, error) {
	if m.status != MembershipStatusRemoved {
		return "", errors.New("reactivateはremoved状態からのみ可能です")
	}

	m.status = MembershipStatusActive
	m.updatedAt = time.Now()
	return MembershipEventReactivated, nil
}
