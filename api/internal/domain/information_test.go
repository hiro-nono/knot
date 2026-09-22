package domain

import (
	"testing"

	"uuid"
)

func TestNewInformation(t *testing.T) {
	accountID := uuid.New()
	createdByUserID := uuid.New()

	if _, err := NewInformation(uuid.Nil(), createdByUserID, "旅行のお知らせ", InformationAccessTypePublic, InformationResponsePolicyAnonymous); err == nil {
		t.Error("NewInformation(Nil, ...) error = nil, want error")
	}
	if _, err := NewInformation(accountID, uuid.Nil(), "旅行のお知らせ", InformationAccessTypePublic, InformationResponsePolicyAnonymous); err == nil {
		t.Error("NewInformation(..., Nil, ...) error = nil, want error")
	}
	if _, err := NewInformation(accountID, createdByUserID, "", InformationAccessTypePublic, InformationResponsePolicyAnonymous); err == nil {
		t.Error(`NewInformation(..., "", ...) error = nil, want error`)
	}
	if _, err := NewInformation(accountID, createdByUserID, "旅行のお知らせ", InformationAccessType("invalid"), InformationResponsePolicyAnonymous); err == nil {
		t.Error(`NewInformation(..., "invalid", ...) error = nil, want error`)
	}
	if _, err := NewInformation(accountID, createdByUserID, "旅行のお知らせ", InformationAccessTypePublic, InformationResponsePolicy("invalid")); err == nil {
		t.Error(`NewInformation(..., "invalid") error = nil, want error`)
	}

	info, err := NewInformation(accountID, createdByUserID, "旅行のお知らせ", InformationAccessTypeRestricted, InformationResponsePolicyAuthenticated)
	if err != nil {
		t.Fatalf("NewInformation() error = %v", err)
	}
	if info.AccountID() != accountID {
		t.Errorf("AccountID() = %v, want %v", info.AccountID(), accountID)
	}
	if info.CreatedByUserID() != createdByUserID {
		t.Errorf("CreatedByUserID() = %v, want %v", info.CreatedByUserID(), createdByUserID)
	}
	if info.AccessType() != InformationAccessTypeRestricted {
		t.Errorf("AccessType() = %v, want %v", info.AccessType(), InformationAccessTypeRestricted)
	}
	if info.IsPublic() {
		t.Error("IsPublic() = true, want false")
	}

	pub, err := NewInformation(accountID, createdByUserID, "旅行のお知らせ", InformationAccessTypePublic, InformationResponsePolicyAnonymous)
	if err != nil {
		t.Fatalf("NewInformation() error = %v", err)
	}
	if !pub.IsPublic() {
		t.Error("IsPublic() = false, want true")
	}
}

func TestInformation_RequiresAuthenticatedResponder_And_RequiresRecipientForResponse(t *testing.T) {
	accountID := uuid.New()
	createdByUserID := uuid.New()

	publicAnonymous, err := NewInformation(accountID, createdByUserID, "旅行のお知らせ", InformationAccessTypePublic, InformationResponsePolicyAnonymous)
	if err != nil {
		t.Fatalf("NewInformation() error = %v", err)
	}
	if publicAnonymous.RequiresAuthenticatedResponder() {
		t.Error("RequiresAuthenticatedResponder() = true, want false for public+anonymous")
	}
	if publicAnonymous.RequiresRecipientForResponse() {
		t.Error("RequiresRecipientForResponse() = true, want false for public+anonymous")
	}

	publicAuthenticated, err := NewInformation(accountID, createdByUserID, "旅行のお知らせ", InformationAccessTypePublic, InformationResponsePolicyAuthenticated)
	if err != nil {
		t.Fatalf("NewInformation() error = %v", err)
	}
	if !publicAuthenticated.RequiresAuthenticatedResponder() {
		t.Error("RequiresAuthenticatedResponder() = false, want true for public+authenticated")
	}
	if publicAuthenticated.RequiresRecipientForResponse() {
		t.Error("RequiresRecipientForResponse() = true, want false for public+authenticated")
	}

	// restrictedはresponse_policyに関わらず常に認証済みUser・Recipientの両方が必須。
	restricted, err := NewInformation(accountID, createdByUserID, "旅行のお知らせ", InformationAccessTypeRestricted, InformationResponsePolicyAnonymous)
	if err != nil {
		t.Fatalf("NewInformation() error = %v", err)
	}
	if !restricted.RequiresAuthenticatedResponder() {
		t.Error("RequiresAuthenticatedResponder() = false, want true for restricted")
	}
	if !restricted.RequiresRecipientForResponse() {
		t.Error("RequiresRecipientForResponse() = false, want true for restricted")
	}
}

func TestInformation_AddSource(t *testing.T) {
	info, err := NewInformation(uuid.New(), uuid.New(), "旅行のお知らせ", InformationAccessTypePublic, InformationResponsePolicyAnonymous)
	if err != nil {
		t.Fatalf("NewInformation() error = %v", err)
	}

	source, err := NewSource(info.ID(), SourceTypeSchedule, "date", "2026-10-01", SourceStatusConfirmed, nil)
	if err != nil {
		t.Fatalf("NewSource() error = %v", err)
	}
	if err := info.AddSource(*source); err != nil {
		t.Fatalf("AddSource() error = %v", err)
	}
	if len(info.Sources()) != 1 {
		t.Fatalf("len(Sources()) = %d, want 1", len(info.Sources()))
	}

	otherSource, err := NewSource(uuid.New(), SourceTypeSchedule, "date", "2026-10-01", SourceStatusConfirmed, nil)
	if err != nil {
		t.Fatalf("NewSource() error = %v", err)
	}
	if err := info.AddSource(*otherSource); err == nil {
		t.Error("AddSource() with mismatched informationID error = nil, want error")
	}
}
