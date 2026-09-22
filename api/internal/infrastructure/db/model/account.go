package model

import (
	"time"
	"uuid"
)

// AccountType はaccountsテーブルのaccount_typeカラムの値を表す。
type AccountType string

const (
	AccountTypePersonal     AccountType = "personal"
	AccountTypeOrganization AccountType = "organization"
)

// AccountRole はaccountsテーブルのroleカラムの値を表す。
type AccountRole string

const (
	AccountRoleAdmin AccountRole = "admin"
	AccountRoleUser  AccountRole = "user"
)

// AccountStatus はaccountsテーブルのstatusカラムの値を表す。
type AccountStatus string

const (
	AccountStatusActive    AccountStatus = "active"
	AccountStatusFrozen    AccountStatus = "frozen"
	AccountStatusSuspended AccountStatus = "suspended"
	AccountStatusWithdrawn AccountStatus = "withdrawn"
	AccountStatusBanned    AccountStatus = "banned"
)

// Account はaccountsテーブルの行を表す。
// 認証は外部プロバイダ(Supabase)に委譲しており、プロバイダ自体は未確定のため、
// 現時点ではプロバイダ側の不透明なユーザーIDのみを保持する。
// 氏名・言語などのプロフィール情報はUserが別途保持する。
// NameはOrganizationの名称(personalの場合はnil)。
type Account struct {
	ID          uuid.UUID
	ProviderID  string
	AccountType AccountType
	Name        *string
	Role        AccountRole
	Status      AccountStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// User はusersテーブルの行を表す。
// 認証されたユーザー自身のプロフィール情報を1件保持する。Accountとの所属関係・
// 権限はMembershipが持つため、account_idは持たない。
type User struct {
	ID        uuid.UUID
	LastName  string
	FirstName string
	Language  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// MembershipRole はmembershipsテーブルのroleカラムの値を表す。
type MembershipRole string

const (
	MembershipRoleOwner  MembershipRole = "owner"
	MembershipRoleAdmin  MembershipRole = "admin"
	MembershipRoleMember MembershipRole = "member"
)

// MembershipStatus はmembershipsテーブルのstatusカラムの値を表す。
type MembershipStatus string

const (
	MembershipStatusActive  MembershipStatus = "active"
	MembershipStatusRemoved MembershipStatus = "removed"
)

// Membership はmembershipsテーブルの行を表す。
// UserがどのAccount(Organization)に所属し、その中でどんな権限(role)・
// 状態(status)を持つかを1件保持する。同一account_id・user_idの組み合わせは
// 常に1件のみ存在する。
type Membership struct {
	ID        uuid.UUID
	AccountID uuid.UUID
	UserID    uuid.UUID
	Role      MembershipRole
	Status    MembershipStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

// MembershipEventType はmembership_eventsテーブルのevent_typeカラムの値を表す。
type MembershipEventType string

const (
	MembershipEventCreated     MembershipEventType = "created"
	MembershipEventRoleChanged MembershipEventType = "role_changed"
	MembershipEventRemoved     MembershipEventType = "removed"
	MembershipEventReactivated MembershipEventType = "reactivated"
)

// MembershipEvent はmembership_eventsテーブルの行を表す。
// Membershipに対して何が起きたかを1件保持する(履歴、追記のみ)。
type MembershipEvent struct {
	ID           uuid.UUID
	MembershipID uuid.UUID
	EventType    MembershipEventType
	CreatedAt    time.Time
}

// MembershipRemovalRequestStatus はmembership_removal_requestsテーブルの
// statusカラムの値を表す。
type MembershipRemovalRequestStatus string

const (
	MembershipRemovalRequestStatusPending  MembershipRemovalRequestStatus = "pending"
	MembershipRemovalRequestStatusApproved MembershipRemovalRequestStatus = "approved"
	MembershipRemovalRequestStatusRejected MembershipRemovalRequestStatus = "rejected"
)

// MembershipRemovalRequest はmembership_removal_requestsテーブルの行を表す。
// AdminによるMembership除外の申請と、Ownerによる承認待ち状態を1件保持する。
type MembershipRemovalRequest struct {
	ID                uuid.UUID
	MembershipID      uuid.UUID
	RequestedByUserID uuid.UUID
	Status            MembershipRemovalRequestStatus
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// AccountStatusEventType はaccount_status_eventsテーブルのevent_typeカラムの値を表す。
type AccountStatusEventType string

const (
	AccountStatusEventCreated     AccountStatusEventType = "created"
	AccountStatusEventFrozen      AccountStatusEventType = "frozen"
	AccountStatusEventSuspended   AccountStatusEventType = "suspended"
	AccountStatusEventWithdrawn   AccountStatusEventType = "withdrawn"
	AccountStatusEventBanned      AccountStatusEventType = "banned"
	AccountStatusEventUnfrozen    AccountStatusEventType = "unfrozen"
	AccountStatusEventReactivated AccountStatusEventType = "reactivated"
)

// AccountStatusEvent はaccount_status_eventsテーブルの行を表す。
// Accountのステータス更新イベントを1件保持する(履歴、追記のみ)。
type AccountStatusEvent struct {
	ID        uuid.UUID
	AccountID uuid.UUID
	EventType AccountStatusEventType
	CreatedAt time.Time
}
