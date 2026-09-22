package domain

import (
	"errors"
	"fmt"
	"time"

	"uuid"
)

// AccountRole はAccountの権限区分を表す。
type AccountRole string

const (
	AccountRoleAdmin AccountRole = "admin"
	AccountRoleUser  AccountRole = "user"
)

func (r AccountRole) valid() bool {
	switch r {
	case AccountRoleAdmin, AccountRoleUser:
		return true
	default:
		return false
	}
}

// AccountType はAccountの利用形態を表す。
type AccountType string

const (
	// AccountTypePersonal は個人利用のAccountを表す。
	AccountTypePersonal AccountType = "personal"
	// AccountTypeOrganization は企業などの組織利用のAccountを表す。
	AccountTypeOrganization AccountType = "organization"
)

func (t AccountType) valid() bool {
	switch t {
	case AccountTypePersonal, AccountTypeOrganization:
		return true
	default:
		return false
	}
}

// AccountStatus はAccountの状態を表す。
type AccountStatus string

const (
	// AccountStatusActive は通常利用可能な状態を表す。
	AccountStatusActive AccountStatus = "active"
	// AccountStatusFrozen は凍結された状態を表す。Unfreezeでのみactiveに戻せる。
	AccountStatusFrozen AccountStatus = "frozen"
	// AccountStatusSuspended は一時停止された状態を表す。Reactivateでのみactiveに戻せる。
	AccountStatusSuspended AccountStatus = "suspended"
	// AccountStatusWithdrawn は退会(論理削除)された状態を表す。
	AccountStatusWithdrawn AccountStatus = "withdrawn"
	// AccountStatusBanned はアカウント停止(admin判断による停止)された状態を表す。
	// Reactivateでのみactiveに戻せる。
	AccountStatusBanned AccountStatus = "banned"
)

func (s AccountStatus) valid() bool {
	switch s {
	case AccountStatusActive, AccountStatusFrozen, AccountStatusSuspended, AccountStatusWithdrawn, AccountStatusBanned:
		return true
	default:
		return false
	}
}

// Account は認証されたIdentityを表すEntity。
// 認証は外部プロバイダ(Supabase)に委譲しており、プロバイダ側の不透明なユーザーIDのみを
// 保持する。氏名・言語などのプロフィール情報はUserが別途保持する。
// accountTypeにより、個人利用(personal)か組織利用(organization)かを区別する。
// 組織利用のAccountは複数のUserを持つことができる(1:N)。
// nameはOrganizationの名称を表す。personalなAccountはnameを持たず(nil)、
// organizationなAccountは必ずnameを持つ。
type Account struct {
	id          uuid.UUID
	providerID  string
	accountType AccountType
	name        *string
	role        AccountRole
	status      AccountStatus
	createdAt   time.Time
	updatedAt   time.Time
}

// NewAccount は新しいAccountを生成する。
// roleは常にuserとして作成される(adminへの昇格は本Entityでは扱わない)。
// nameは、accountTypeがpersonalの場合は必ずnil、organizationの場合は
// 必ず非空文字列である必要がある。
func NewAccount(providerID string, accountType AccountType, name *string) (*Account, error) {
	if providerID == "" {
		return nil, errors.New("providerID は必須です")
	}
	if !accountType.valid() {
		return nil, fmt.Errorf("account_type の値が不正です: %q", accountType)
	}
	if err := validateAccountName(accountType, name); err != nil {
		return nil, err
	}

	now := time.Now()
	return &Account{
		id:          uuid.New(),
		providerID:  providerID,
		accountType: accountType,
		name:        name,
		role:        AccountRoleUser,
		status:      AccountStatusActive,
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

func validateAccountName(accountType AccountType, name *string) error {
	switch accountType {
	case AccountTypePersonal:
		if name != nil {
			return errors.New("personalなaccountにnameを設定することはできません")
		}
	case AccountTypeOrganization:
		if name == nil || *name == "" {
			return errors.New("organizationなaccountにはnameが必須です")
		}
	}
	return nil
}

// ReconstructAccount は永続化されたデータからAccountを復元する。
// 呼び出し元(repository)がすでに正当なデータであることを保証するため、
// バリデーションは行わない。
func ReconstructAccount(id uuid.UUID, providerID string, accountType AccountType, name *string, role AccountRole, status AccountStatus, createdAt, updatedAt time.Time) *Account {
	return &Account{
		id:          id,
		providerID:  providerID,
		accountType: accountType,
		name:        name,
		role:        role,
		status:      status,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}
}

func (a *Account) ID() uuid.UUID {
	return a.id
}

func (a *Account) ProviderID() string {
	return a.providerID
}

func (a *Account) AccountType() AccountType {
	return a.accountType
}

// Name はOrganizationの名称を返す。personalなAccountの場合はnilを返す。
func (a *Account) Name() *string {
	return a.name
}

func (a *Account) Role() AccountRole {
	return a.role
}

func (a *Account) Status() AccountStatus {
	return a.status
}

func (a *Account) IsAdmin() bool {
	return a.role == AccountRoleAdmin
}

func (a *Account) CreatedAt() time.Time {
	return a.createdAt
}

func (a *Account) UpdatedAt() time.Time {
	return a.updatedAt
}

// Freeze はactiveなAccountを凍結する。
func (a *Account) Freeze() (AccountStatusEventType, error) {
	if a.status != AccountStatusActive {
		return "", fmt.Errorf("freeze is only allowed from active status, current status is %q", a.status)
	}
	a.status = AccountStatusFrozen
	a.updatedAt = time.Now()
	return AccountStatusEventFrozen, nil
}

// Suspend はactiveなAccountを一時停止する。
func (a *Account) Suspend() (AccountStatusEventType, error) {
	if a.status != AccountStatusActive {
		return "", fmt.Errorf("suspend is only allowed from active status, current status is %q", a.status)
	}
	a.status = AccountStatusSuspended
	a.updatedAt = time.Now()
	return AccountStatusEventSuspended, nil
}

// Ban はwithdrawn以外のAccountをアカウント停止する。
func (a *Account) Ban() (AccountStatusEventType, error) {
	if a.status == AccountStatusWithdrawn {
		return "", errors.New("banning a withdrawn account is not allowed")
	}
	if a.status == AccountStatusBanned {
		return "", errors.New("account is already banned")
	}
	a.status = AccountStatusBanned
	a.updatedAt = time.Now()
	return AccountStatusEventBanned, nil
}

// Withdraw はAccountを退会(論理削除)する。
func (a *Account) Withdraw() (AccountStatusEventType, error) {
	if a.status == AccountStatusWithdrawn {
		return "", errors.New("account is already withdrawn")
	}
	a.status = AccountStatusWithdrawn
	a.updatedAt = time.Now()
	return AccountStatusEventWithdrawn, nil
}

// Unfreeze はfrozenなAccountをactiveに戻す。
func (a *Account) Unfreeze() (AccountStatusEventType, error) {
	if a.status != AccountStatusFrozen {
		return "", fmt.Errorf("unfreeze is only allowed from frozen status, current status is %q", a.status)
	}
	a.status = AccountStatusActive
	a.updatedAt = time.Now()
	return AccountStatusEventUnfrozen, nil
}

// Reactivate はsuspendedまたはbannedなAccountをactiveに戻す。
func (a *Account) Reactivate() (AccountStatusEventType, error) {
	if a.status != AccountStatusSuspended && a.status != AccountStatusBanned {
		return "", fmt.Errorf("reactivate is only allowed from suspended or banned status, current status is %q", a.status)
	}
	a.status = AccountStatusActive
	a.updatedAt = time.Now()
	return AccountStatusEventReactivated, nil
}
