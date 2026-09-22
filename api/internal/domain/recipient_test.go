package domain

import (
	"testing"

	"uuid"
)

func TestNewRecipient(t *testing.T) {
	informationID := uuid.New()
	userID := uuid.New()

	if _, err := NewRecipient(uuid.Nil(), userID); err == nil {
		t.Error("NewRecipient(Nil, ...) error = nil, want error")
	}
	if _, err := NewRecipient(informationID, uuid.Nil()); err == nil {
		t.Error("NewRecipient(..., Nil) error = nil, want error")
	}

	r, err := NewRecipient(informationID, userID)
	if err != nil {
		t.Fatalf("NewRecipient() error = %v", err)
	}
	if r.InformationID() != informationID {
		t.Errorf("InformationID() = %v, want %v", r.InformationID(), informationID)
	}
	if r.UserID() != userID {
		t.Errorf("UserID() = %v, want %v", r.UserID(), userID)
	}
}
