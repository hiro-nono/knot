package domain

import (
	"testing"

	"uuid"
)

func TestNewResponseItem(t *testing.T) {
	responseID := uuid.New()
	sourceID := uuid.New()
	optionID := uuid.New()
	value := "はい"

	if _, err := NewResponseItem(uuid.Nil(), sourceID, &optionID, nil); err == nil {
		t.Error("NewResponseItem(Nil, ...) error = nil, want error")
	}
	if _, err := NewResponseItem(responseID, uuid.Nil(), &optionID, nil); err == nil {
		t.Error("NewResponseItem(..., Nil, ...) error = nil, want error")
	}
	if _, err := NewResponseItem(responseID, sourceID, nil, nil); err == nil {
		t.Error("NewResponseItem(..., nil, nil) error = nil, want error (both unset)")
	}
	if _, err := NewResponseItem(responseID, sourceID, &optionID, &value); err == nil {
		t.Error("NewResponseItem(..., optionID, value) error = nil, want error (both set)")
	}

	optionItem, err := NewResponseItem(responseID, sourceID, &optionID, nil)
	if err != nil {
		t.Fatalf("NewResponseItem() error = %v", err)
	}
	if optionItem.OptionID() == nil || *optionItem.OptionID() != optionID {
		t.Errorf("OptionID() = %v, want %v", optionItem.OptionID(), optionID)
	}
	if optionItem.Value() != nil {
		t.Errorf("Value() = %v, want nil", optionItem.Value())
	}

	valueItem, err := NewResponseItem(responseID, sourceID, nil, &value)
	if err != nil {
		t.Fatalf("NewResponseItem() error = %v", err)
	}
	if valueItem.Value() == nil || *valueItem.Value() != value {
		t.Errorf("Value() = %v, want %v", valueItem.Value(), value)
	}
	if valueItem.OptionID() != nil {
		t.Errorf("OptionID() = %v, want nil", valueItem.OptionID())
	}
}
