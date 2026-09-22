package usecase

import (
	"time"

	"knot-api/internal/domain"
)

// MembershipView はMembershipを外部に返却するための表現。
type MembershipView struct {
	ID        string    `json:"id"`
	AccountID string    `json:"account_id"`
	UserID    string    `json:"user_id"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func newMembershipView(membership *domain.Membership) *MembershipView {
	return &MembershipView{
		ID:        membership.ID().String(),
		AccountID: membership.AccountID().String(),
		UserID:    membership.UserID().String(),
		Role:      string(membership.Role()),
		Status:    string(membership.Status()),
		CreatedAt: membership.CreatedAt(),
		UpdatedAt: membership.UpdatedAt(),
	}
}

// MembershipRemovalRequestView はMembershipRemovalRequestを外部に返却するための表現。
type MembershipRemovalRequestView struct {
	ID                string    `json:"id"`
	MembershipID      string    `json:"membership_id"`
	RequestedByUserID string    `json:"requested_by_user_id"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func newMembershipRemovalRequestView(request *domain.MembershipRemovalRequest) *MembershipRemovalRequestView {
	return &MembershipRemovalRequestView{
		ID:                request.ID().String(),
		MembershipID:      request.MembershipID().String(),
		RequestedByUserID: request.RequestedByUserID().String(),
		Status:            string(request.Status()),
		CreatedAt:         request.CreatedAt(),
		UpdatedAt:         request.UpdatedAt(),
	}
}
