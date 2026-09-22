package domain

import (
	"testing"
	"time"

	"uuid"
)

func TestNewResponse(t *testing.T) {
	informationID := uuid.New()
	userID := uuid.New()

	if _, err := NewResponse(uuid.Nil(), &userID); err == nil {
		t.Error("NewResponse(Nil, ...) error = nil, want error")
	}
	nilUUID := uuid.Nil()
	if _, err := NewResponse(informationID, &nilUUID); err == nil {
		t.Error("NewResponse(..., &Nil) error = nil, want error")
	}

	r, err := NewResponse(informationID, &userID)
	if err != nil {
		t.Fatalf("NewResponse() error = %v", err)
	}
	if r.InformationID() != informationID {
		t.Errorf("InformationID() = %v, want %v", r.InformationID(), informationID)
	}
	if r.UserID() == nil || *r.UserID() != userID {
		t.Errorf("UserID() = %v, want %v", r.UserID(), userID)
	}
	if len(r.Items()) != 0 {
		t.Errorf("Items() = %v, want empty", r.Items())
	}

	// userIDはnil(匿名回答)を許容する。
	anon, err := NewResponse(informationID, nil)
	if err != nil {
		t.Fatalf("NewResponse(nil) error = %v", err)
	}
	if anon.UserID() != nil {
		t.Errorf("UserID() = %v, want nil", anon.UserID())
	}
}

func TestResponse_AddItem(t *testing.T) {
	r, err := NewResponse(uuid.New(), nil)
	if err != nil {
		t.Fatalf("NewResponse() error = %v", err)
	}

	value := "はい"
	item, err := NewResponseItem(r.ID(), uuid.New(), nil, &value)
	if err != nil {
		t.Fatalf("NewResponseItem() error = %v", err)
	}
	if err := r.AddItem(*item); err != nil {
		t.Fatalf("AddItem() error = %v", err)
	}
	if len(r.Items()) != 1 {
		t.Fatalf("len(Items()) = %d, want 1", len(r.Items()))
	}

	otherItem, err := NewResponseItem(uuid.New(), uuid.New(), nil, &value)
	if err != nil {
		t.Fatalf("NewResponseItem() error = %v", err)
	}
	if err := r.AddItem(*otherItem); err == nil {
		t.Error("AddItem() with mismatched responseID error = nil, want error")
	}
}

func TestReconstructResponse_NilItems(t *testing.T) {
	r := ReconstructResponse(uuid.New(), uuid.New(), nil, nil, time.Now(), time.Now())
	if r.Items() == nil {
		t.Error("Items() = nil, want empty slice")
	}
	if r.UserID() != nil {
		t.Errorf("UserID() = %v, want nil", r.UserID())
	}
}
