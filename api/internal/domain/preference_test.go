package domain

import (
	"testing"
	"time"

	"uuid"
)

func TestNewPreference(t *testing.T) {
	if _, err := NewPreference(uuid.Nil()); err == nil {
		t.Error("NewPreference(Nil) error = nil, want error")
	}

	userID := uuid.New()
	p, err := NewPreference(userID)
	if err != nil {
		t.Fatalf("NewPreference() error = %v", err)
	}
	if p.UserID() != userID {
		t.Errorf("UserID() = %v, want %v", p.UserID(), userID)
	}
	if len(p.Items()) != 0 {
		t.Errorf("Items() = %v, want empty", p.Items())
	}
}

func TestPreference_SetAndGet(t *testing.T) {
	p, err := NewPreference(uuid.New())
	if err != nil {
		t.Fatalf("NewPreference() error = %v", err)
	}

	if err := p.Set("", "value"); err == nil {
		t.Error(`Set("", ...) error = nil, want error`)
	}

	if err := p.Set("reading_level", "easy"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	v, ok := p.Get("reading_level")
	if !ok || v != "easy" {
		t.Errorf("Get(\"reading_level\") = (%q, %v), want (\"easy\", true)", v, ok)
	}

	if _, ok := p.Get("unknown"); ok {
		t.Error(`Get("unknown") ok = true, want false`)
	}
}

func TestPreference_Items_ReturnsCopy(t *testing.T) {
	p, err := NewPreference(uuid.New())
	if err != nil {
		t.Fatalf("NewPreference() error = %v", err)
	}
	if err := p.Set("language", "ja"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	items := p.Items()
	items["language"] = "en"

	v, _ := p.Get("language")
	if v != "ja" {
		t.Errorf("mutating Items() result affected internal state: Get(\"language\") = %q, want \"ja\"", v)
	}
}

func TestReconstructPreference_NilItems(t *testing.T) {
	p := ReconstructPreference(uuid.New(), uuid.New(), nil, time.Time{}, time.Time{})
	if p.Items() == nil {
		t.Error("Items() = nil, want empty map")
	}
}
