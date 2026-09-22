package domain

import "testing"

func strPtrAccount(s string) *string { return &s }

func TestNewAccount(t *testing.T) {
	if _, err := NewAccount("", AccountTypePersonal, nil); err == nil {
		t.Error(`NewAccount("", ...) error = nil, want error`)
	}
	if _, err := NewAccount("supabase-user-1", AccountType("invalid"), nil); err == nil {
		t.Error(`NewAccount(..., "invalid") error = nil, want error`)
	}
	if _, err := NewAccount("supabase-user-1", AccountTypePersonal, strPtrAccount("nope")); err == nil {
		t.Error("NewAccount(personal, name) error = nil, want error")
	}
	if _, err := NewAccount("supabase-org-1", AccountTypeOrganization, nil); err == nil {
		t.Error("NewAccount(organization, nil) error = nil, want error")
	}
	if _, err := NewAccount("supabase-org-1", AccountTypeOrganization, strPtrAccount("")); err == nil {
		t.Error(`NewAccount(organization, "") error = nil, want error`)
	}

	a, err := NewAccount("supabase-user-1", AccountTypePersonal, nil)
	if err != nil {
		t.Fatalf("NewAccount() error = %v", err)
	}
	if a.AccountType() != AccountTypePersonal {
		t.Errorf("AccountType() = %v, want %v", a.AccountType(), AccountTypePersonal)
	}
	if a.Name() != nil {
		t.Errorf("Name() = %v, want nil", a.Name())
	}
	if a.Role() != AccountRoleUser {
		t.Errorf("Role() = %v, want %v", a.Role(), AccountRoleUser)
	}
	if a.Status() != AccountStatusActive {
		t.Errorf("Status() = %v, want %v", a.Status(), AccountStatusActive)
	}
	if a.IsAdmin() {
		t.Error("IsAdmin() = true, want false")
	}

	org, err := NewAccount("supabase-org-1", AccountTypeOrganization, strPtrAccount("Acme, Inc."))
	if err != nil {
		t.Fatalf("NewAccount() error = %v", err)
	}
	if org.AccountType() != AccountTypeOrganization {
		t.Errorf("AccountType() = %v, want %v", org.AccountType(), AccountTypeOrganization)
	}
	if org.Name() == nil || *org.Name() != "Acme, Inc." {
		t.Errorf("Name() = %v, want %q", org.Name(), "Acme, Inc.")
	}
}

func TestAccount_Freeze_Unfreeze(t *testing.T) {
	a, _ := NewAccount("p1", AccountTypePersonal, nil)

	eventType, err := a.Freeze()
	if err != nil {
		t.Fatalf("Freeze() error = %v", err)
	}
	if eventType != AccountStatusEventFrozen {
		t.Errorf("eventType = %v, want %v", eventType, AccountStatusEventFrozen)
	}
	if a.Status() != AccountStatusFrozen {
		t.Errorf("Status() = %v, want %v", a.Status(), AccountStatusFrozen)
	}

	if _, err := a.Freeze(); err == nil {
		t.Error("Freeze() on already-frozen account error = nil, want error")
	}
	if _, err := a.Reactivate(); err == nil {
		t.Error("Reactivate() on frozen account error = nil, want error")
	}

	eventType, err = a.Unfreeze()
	if err != nil {
		t.Fatalf("Unfreeze() error = %v", err)
	}
	if eventType != AccountStatusEventUnfrozen {
		t.Errorf("eventType = %v, want %v", eventType, AccountStatusEventUnfrozen)
	}
	if a.Status() != AccountStatusActive {
		t.Errorf("Status() = %v, want %v", a.Status(), AccountStatusActive)
	}

	if _, err := a.Unfreeze(); err == nil {
		t.Error("Unfreeze() on active account error = nil, want error")
	}
}

func TestAccount_Suspend_Reactivate(t *testing.T) {
	a, _ := NewAccount("p1", AccountTypePersonal, nil)

	if _, err := a.Suspend(); err != nil {
		t.Fatalf("Suspend() error = %v", err)
	}
	if a.Status() != AccountStatusSuspended {
		t.Errorf("Status() = %v, want %v", a.Status(), AccountStatusSuspended)
	}

	eventType, err := a.Reactivate()
	if err != nil {
		t.Fatalf("Reactivate() error = %v", err)
	}
	if eventType != AccountStatusEventReactivated {
		t.Errorf("eventType = %v, want %v", eventType, AccountStatusEventReactivated)
	}
	if a.Status() != AccountStatusActive {
		t.Errorf("Status() = %v, want %v", a.Status(), AccountStatusActive)
	}
}

func TestAccount_Ban_Reactivate(t *testing.T) {
	a, _ := NewAccount("p1", AccountTypePersonal, nil)

	if _, err := a.Ban(); err != nil {
		t.Fatalf("Ban() error = %v", err)
	}
	if a.Status() != AccountStatusBanned {
		t.Errorf("Status() = %v, want %v", a.Status(), AccountStatusBanned)
	}
	if _, err := a.Ban(); err == nil {
		t.Error("Ban() on already-banned account error = nil, want error")
	}

	if _, err := a.Reactivate(); err != nil {
		t.Fatalf("Reactivate() error = %v", err)
	}
	if a.Status() != AccountStatusActive {
		t.Errorf("Status() = %v, want %v", a.Status(), AccountStatusActive)
	}
}

func TestAccount_Withdraw(t *testing.T) {
	a, _ := NewAccount("p1", AccountTypePersonal, nil)

	eventType, err := a.Withdraw()
	if err != nil {
		t.Fatalf("Withdraw() error = %v", err)
	}
	if eventType != AccountStatusEventWithdrawn {
		t.Errorf("eventType = %v, want %v", eventType, AccountStatusEventWithdrawn)
	}
	if a.Status() != AccountStatusWithdrawn {
		t.Errorf("Status() = %v, want %v", a.Status(), AccountStatusWithdrawn)
	}

	if _, err := a.Withdraw(); err == nil {
		t.Error("Withdraw() on already-withdrawn account error = nil, want error")
	}
	if _, err := a.Ban(); err == nil {
		t.Error("Ban() on withdrawn account error = nil, want error")
	}
}
