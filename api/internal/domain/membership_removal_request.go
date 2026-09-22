package domain

import (
	"errors"
	"time"

	"uuid"
)

// MembershipRemovalRequestStatus はMembershipRemovalRequestの状態を表す。
type MembershipRemovalRequestStatus string

const (
	MembershipRemovalRequestStatusPending  MembershipRemovalRequestStatus = "pending"
	MembershipRemovalRequestStatusApproved MembershipRemovalRequestStatus = "approved"
	MembershipRemovalRequestStatusRejected MembershipRemovalRequestStatus = "rejected"
)

// MembershipRemovalRequest は、AdminがMembershipの除外(status変更)を
// 申請したことを表すEntity。
//
// Adminによる除外操作はOwnerの承認が必要であるため、Admin操作は即座に
// Membershipのstatusを変更せず、承認待ち(pending)のリクエストとして記録する。
// Ownerが承認(Approve)して初めてMembershipが実際にremoved状態になる。
// (Owner自身による除外操作はこのリクエストを経由せず、即座に反映される。)
type MembershipRemovalRequest struct {
	id                uuid.UUID
	membershipID      uuid.UUID
	requestedByUserID uuid.UUID
	status            MembershipRemovalRequestStatus
	createdAt         time.Time
	updatedAt         time.Time
}

// NewMembershipRemovalRequest は新しいMembershipRemovalRequestを生成する。
// 常にpending状態で作成される。
func NewMembershipRemovalRequest(membershipID, requestedByUserID uuid.UUID) (*MembershipRemovalRequest, error) {
	if membershipID == uuid.Nil() {
		return nil, errors.New("membershipID は必須です")
	}
	if requestedByUserID == uuid.Nil() {
		return nil, errors.New("requestedByUserID は必須です")
	}

	now := time.Now()
	return &MembershipRemovalRequest{
		id:                uuid.New(),
		membershipID:      membershipID,
		requestedByUserID: requestedByUserID,
		status:            MembershipRemovalRequestStatusPending,
		createdAt:         now,
		updatedAt:         now,
	}, nil
}

// ReconstructMembershipRemovalRequest は永続化されたデータからMembershipRemovalRequestを
// 復元する。バリデーションは行わない。
func ReconstructMembershipRemovalRequest(id, membershipID, requestedByUserID uuid.UUID, status MembershipRemovalRequestStatus, createdAt, updatedAt time.Time) *MembershipRemovalRequest {
	return &MembershipRemovalRequest{
		id:                id,
		membershipID:      membershipID,
		requestedByUserID: requestedByUserID,
		status:            status,
		createdAt:         createdAt,
		updatedAt:         updatedAt,
	}
}

func (r *MembershipRemovalRequest) ID() uuid.UUID {
	return r.id
}

func (r *MembershipRemovalRequest) MembershipID() uuid.UUID {
	return r.membershipID
}

func (r *MembershipRemovalRequest) RequestedByUserID() uuid.UUID {
	return r.requestedByUserID
}

func (r *MembershipRemovalRequest) Status() MembershipRemovalRequestStatus {
	return r.status
}

func (r *MembershipRemovalRequest) IsPending() bool {
	return r.status == MembershipRemovalRequestStatusPending
}

func (r *MembershipRemovalRequest) CreatedAt() time.Time {
	return r.createdAt
}

func (r *MembershipRemovalRequest) UpdatedAt() time.Time {
	return r.updatedAt
}

// Approve はpending状態のリクエストを承認済みにする(Ownerが実行する)。
func (r *MembershipRemovalRequest) Approve() error {
	if !r.IsPending() {
		return errors.New("pending状態のリクエストのみ承認できます")
	}
	r.status = MembershipRemovalRequestStatusApproved
	r.updatedAt = time.Now()
	return nil
}

// Reject はpending状態のリクエストを却下済みにする(Ownerが実行する)。
func (r *MembershipRemovalRequest) Reject() error {
	if !r.IsPending() {
		return errors.New("pending状態のリクエストのみ却下できます")
	}
	r.status = MembershipRemovalRequestStatusRejected
	r.updatedAt = time.Now()
	return nil
}
