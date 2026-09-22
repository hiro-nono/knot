package domain

import (
	"testing"

	"uuid"
)

func TestNewMembershipEvent(t *testing.T) {
	membershipID := uuid.New()

	if _, err := NewMembershipEvent(uuid.Nil(), MembershipEventCreated); err == nil {
		t.Error("NewMembershipEvent(Nil, ...) error = nil, want error")
	}
	if _, err := NewMembershipEvent(membershipID, MembershipEventType("invalid")); err == nil {
		t.Error(`NewMembershipEvent(..., "invalid") error = nil, want error`)
	}

	e, err := NewMembershipEvent(membershipID, MembershipEventCreated)
	if err != nil {
		t.Fatalf("NewMembershipEvent() error = %v", err)
	}
	if e.MembershipID() != membershipID {
		t.Errorf("MembershipID() = %v, want %v", e.MembershipID(), membershipID)
	}
	if e.EventType() != MembershipEventCreated {
		t.Errorf("EventType() = %v, want %v", e.EventType(), MembershipEventCreated)
	}
}
