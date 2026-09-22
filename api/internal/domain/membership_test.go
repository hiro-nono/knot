package domain

import (
	"testing"

	"uuid"
)

func TestNewMembership(t *testing.T) {
	accountID := uuid.New()
	userID := uuid.New()

	if _, err := NewMembership(uuid.Nil(), userID, MembershipRoleOwner); err == nil {
		t.Error("NewMembership(Nil, ...) error = nil, want error")
	}
	if _, err := NewMembership(accountID, uuid.Nil(), MembershipRoleOwner); err == nil {
		t.Error("NewMembership(..., Nil, ...) error = nil, want error")
	}
	if _, err := NewMembership(accountID, userID, MembershipRole("invalid")); err == nil {
		t.Error(`NewMembership(..., "invalid") error = nil, want error`)
	}

	m, err := NewMembership(accountID, userID, MembershipRoleOwner)
	if err != nil {
		t.Fatalf("NewMembership() error = %v", err)
	}
	if m.AccountID() != accountID {
		t.Errorf("AccountID() = %v, want %v", m.AccountID(), accountID)
	}
	if m.UserID() != userID {
		t.Errorf("UserID() = %v, want %v", m.UserID(), userID)
	}
	if !m.IsOwner() {
		t.Error("IsOwner() = false, want true")
	}
	if !m.IsActive() {
		t.Error("IsActive() = false, want true")
	}
	if m.Status() != MembershipStatusActive {
		t.Errorf("Status() = %v, want %v", m.Status(), MembershipStatusActive)
	}
}

func TestMembership_ChangeRole(t *testing.T) {
	m, err := NewMembership(uuid.New(), uuid.New(), MembershipRoleMember)
	if err != nil {
		t.Fatalf("NewMembership() error = %v", err)
	}

	if _, err := m.ChangeRole(MembershipRole("invalid")); err == nil {
		t.Error(`ChangeRole("invalid") error = nil, want error`)
	}
	if _, err := m.ChangeRole(MembershipRoleMember); err == nil {
		t.Error("ChangeRole(same role) error = nil, want error")
	}
	if _, err := m.ChangeRole(MembershipRoleOwner); err == nil {
		t.Error("ChangeRole(owner) error = nil, want error (owner changes are not allowed via this method)")
	}

	eventType, err := m.ChangeRole(MembershipRoleAdmin)
	if err != nil {
		t.Fatalf("ChangeRole() error = %v", err)
	}
	if eventType != MembershipEventRoleChanged {
		t.Errorf("eventType = %v, want %v", eventType, MembershipEventRoleChanged)
	}
	if m.Role() != MembershipRoleAdmin {
		t.Errorf("Role() = %v, want %v", m.Role(), MembershipRoleAdmin)
	}

	owner, err := NewMembership(uuid.New(), uuid.New(), MembershipRoleOwner)
	if err != nil {
		t.Fatalf("NewMembership() error = %v", err)
	}
	if _, err := owner.ChangeRole(MembershipRoleAdmin); err == nil {
		t.Error("ChangeRole() on owner membership error = nil, want error")
	}
}

func TestMembership_Remove_Reactivate(t *testing.T) {
	m, err := NewMembership(uuid.New(), uuid.New(), MembershipRoleMember)
	if err != nil {
		t.Fatalf("NewMembership() error = %v", err)
	}

	if _, err := m.Reactivate(); err == nil {
		t.Error("Reactivate() on active membership error = nil, want error")
	}

	eventType, err := m.Remove()
	if err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
	if eventType != MembershipEventRemoved {
		t.Errorf("eventType = %v, want %v", eventType, MembershipEventRemoved)
	}
	if m.IsActive() {
		t.Error("IsActive() = true, want false after Remove()")
	}

	if _, err := m.Remove(); err == nil {
		t.Error("Remove() on already-removed membership error = nil, want error")
	}

	eventType, err = m.Reactivate()
	if err != nil {
		t.Fatalf("Reactivate() error = %v", err)
	}
	if eventType != MembershipEventReactivated {
		t.Errorf("eventType = %v, want %v", eventType, MembershipEventReactivated)
	}
	if !m.IsActive() {
		t.Error("IsActive() = false, want true after Reactivate()")
	}
}

func TestMembership_Remove_OwnerNotAllowed(t *testing.T) {
	owner, err := NewMembership(uuid.New(), uuid.New(), MembershipRoleOwner)
	if err != nil {
		t.Fatalf("NewMembership() error = %v", err)
	}
	if _, err := owner.Remove(); err == nil {
		t.Error("Remove() on owner membership error = nil, want error")
	}
}
