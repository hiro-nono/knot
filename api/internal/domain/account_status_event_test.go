package domain

import (
	"testing"

	"uuid"
)

func TestNewAccountStatusEvent(t *testing.T) {
	if _, err := NewAccountStatusEvent(uuid.Nil(), AccountStatusEventFrozen); err == nil {
		t.Error("NewAccountStatusEvent(Nil, ...) error = nil, want error")
	}

	accountID := uuid.New()
	if _, err := NewAccountStatusEvent(accountID, AccountStatusEventType("bogus")); err == nil {
		t.Error(`NewAccountStatusEvent(..., "bogus") error = nil, want error`)
	}

	e, err := NewAccountStatusEvent(accountID, AccountStatusEventFrozen)
	if err != nil {
		t.Fatalf("NewAccountStatusEvent() error = %v", err)
	}
	if e.AccountID() != accountID {
		t.Errorf("AccountID() = %v, want %v", e.AccountID(), accountID)
	}
	if e.EventType() != AccountStatusEventFrozen {
		t.Errorf("EventType() = %v, want %v", e.EventType(), AccountStatusEventFrozen)
	}
}
