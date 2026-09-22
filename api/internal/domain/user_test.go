package domain

import (
	"testing"
)

func TestNewUser(t *testing.T) {
	if _, err := NewUser("", "太郎", "ja"); err == nil {
		t.Error(`NewUser("", ...) error = nil, want error`)
	}
	if _, err := NewUser("山田", "", "ja"); err == nil {
		t.Error(`NewUser(..., "", ...) error = nil, want error`)
	}
	if _, err := NewUser("山田", "太郎", ""); err == nil {
		t.Error(`NewUser(..., "") error = nil, want error`)
	}

	u, err := NewUser("山田", "太郎", "ja")
	if err != nil {
		t.Fatalf("NewUser() error = %v", err)
	}
	if u.LastName() != "山田" || u.FirstName() != "太郎" || u.Language() != "ja" {
		t.Errorf("u = %+v, unexpected", u)
	}
}

func TestUser_UpdateProfile(t *testing.T) {
	u, err := NewUser("山田", "太郎", "ja")
	if err != nil {
		t.Fatalf("NewUser() error = %v", err)
	}

	if err := u.UpdateProfile("", "太郎", "ja"); err == nil {
		t.Error(`UpdateProfile("", ...) error = nil, want error`)
	}

	if err := u.UpdateProfile("鈴木", "花子", "en"); err != nil {
		t.Fatalf("UpdateProfile() error = %v", err)
	}
	if u.LastName() != "鈴木" || u.FirstName() != "花子" || u.Language() != "en" {
		t.Errorf("u = %+v, unexpected after update", u)
	}
}
