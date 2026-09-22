package domain

import (
	"testing"

	"uuid"
)

func TestNewMembershipRemovalRequest(t *testing.T) {
	membershipID := uuid.New()
	requestedByUserID := uuid.New()

	if _, err := NewMembershipRemovalRequest(uuid.Nil(), requestedByUserID); err == nil {
		t.Error("NewMembershipRemovalRequest(Nil, ...) error = nil, want error")
	}
	if _, err := NewMembershipRemovalRequest(membershipID, uuid.Nil()); err == nil {
		t.Error("NewMembershipRemovalRequest(..., Nil) error = nil, want error")
	}

	r, err := NewMembershipRemovalRequest(membershipID, requestedByUserID)
	if err != nil {
		t.Fatalf("NewMembershipRemovalRequest() error = %v", err)
	}
	if !r.IsPending() {
		t.Error("IsPending() = false, want true")
	}
}

func TestMembershipRemovalRequest_Approve_Reject(t *testing.T) {
	r, err := NewMembershipRemovalRequest(uuid.New(), uuid.New())
	if err != nil {
		t.Fatalf("NewMembershipRemovalRequest() error = %v", err)
	}

	if err := r.Approve(); err != nil {
		t.Fatalf("Approve() error = %v", err)
	}
	if r.Status() != MembershipRemovalRequestStatusApproved {
		t.Errorf("Status() = %v, want %v", r.Status(), MembershipRemovalRequestStatusApproved)
	}
	if err := r.Approve(); err == nil {
		t.Error("Approve() on already-approved request error = nil, want error")
	}
	if err := r.Reject(); err == nil {
		t.Error("Reject() on already-approved request error = nil, want error")
	}
}

func TestMembershipRemovalRequest_Reject(t *testing.T) {
	r, err := NewMembershipRemovalRequest(uuid.New(), uuid.New())
	if err != nil {
		t.Fatalf("NewMembershipRemovalRequest() error = %v", err)
	}

	if err := r.Reject(); err != nil {
		t.Fatalf("Reject() error = %v", err)
	}
	if r.Status() != MembershipRemovalRequestStatusRejected {
		t.Errorf("Status() = %v, want %v", r.Status(), MembershipRemovalRequestStatusRejected)
	}
}
