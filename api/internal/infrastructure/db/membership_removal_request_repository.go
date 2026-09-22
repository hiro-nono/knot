package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"knot-api/internal/domain"

	"uuid"
)

// MembershipRemovalRequestRepository はPostgreSQLを用いたMembershipRemovalRequestの
// 永続化を行う。
type MembershipRemovalRequestRepository struct {
	conn *sql.DB
}

// NewMembershipRemovalRequestRepository はMembershipRemovalRequestRepositoryを生成する。
func NewMembershipRemovalRequestRepository(conn *sql.DB) *MembershipRemovalRequestRepository {
	return &MembershipRemovalRequestRepository{conn: conn}
}

// Create はMembershipRemovalRequestを新規保存する。
func (r *MembershipRemovalRequestRepository) Create(ctx context.Context, request *domain.MembershipRemovalRequest) error {
	_, err := executorFromContext(ctx, r.conn).ExecContext(ctx, `
		INSERT INTO membership_removal_requests (id, membership_id, requested_by_user_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`,
		request.ID().String(),
		request.MembershipID().String(),
		request.RequestedByUserID().String(),
		string(request.Status()),
		request.CreatedAt(),
		request.UpdatedAt(),
	)
	if err != nil {
		return fmt.Errorf("insert membership removal request: %w", err)
	}

	return nil
}

// Get はIDを指定してMembershipRemovalRequestを1件取得する。
func (r *MembershipRemovalRequestRepository) Get(ctx context.Context, id uuid.UUID) (*domain.MembershipRemovalRequest, error) {
	row := executorFromContext(ctx, r.conn).QueryRowContext(ctx, `
		SELECT id, membership_id, requested_by_user_id, status, created_at, updated_at
		FROM membership_removal_requests
		WHERE id = $1
	`, id.String())

	return scanMembershipRemovalRequest(row)
}

// GetPendingByMembershipID はmembershipIDに対応するpending状態の
// MembershipRemovalRequestを取得する。
func (r *MembershipRemovalRequestRepository) GetPendingByMembershipID(ctx context.Context, membershipID uuid.UUID) (*domain.MembershipRemovalRequest, error) {
	row := executorFromContext(ctx, r.conn).QueryRowContext(ctx, `
		SELECT id, membership_id, requested_by_user_id, status, created_at, updated_at
		FROM membership_removal_requests
		WHERE membership_id = $1 AND status = $2
	`, membershipID.String(), string(domain.MembershipRemovalRequestStatusPending))

	return scanMembershipRemovalRequest(row)
}

// ListPendingByAccountID はaccountIDに属するMembershipのうち、pending状態の
// MembershipRemovalRequestを一覧取得する。
func (r *MembershipRemovalRequestRepository) ListPendingByAccountID(ctx context.Context, accountID uuid.UUID) ([]*domain.MembershipRemovalRequest, error) {
	rows, err := executorFromContext(ctx, r.conn).QueryContext(ctx, `
		SELECT rr.id, rr.membership_id, rr.requested_by_user_id, rr.status, rr.created_at, rr.updated_at
		FROM membership_removal_requests rr
		JOIN memberships m ON m.id = rr.membership_id
		WHERE m.account_id = $1 AND rr.status = $2
		ORDER BY rr.created_at
	`, accountID.String(), string(domain.MembershipRemovalRequestStatusPending))
	if err != nil {
		return nil, fmt.Errorf("select membership removal requests: %w", err)
	}
	defer rows.Close()

	var requests []*domain.MembershipRemovalRequest
	for rows.Next() {
		request, err := scanMembershipRemovalRequestRow(rows)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate membership removal requests: %w", err)
	}

	return requests, nil
}

// Update はMembershipRemovalRequestのstatus等を更新する。
func (r *MembershipRemovalRequestRepository) Update(ctx context.Context, request *domain.MembershipRemovalRequest) error {
	result, err := executorFromContext(ctx, r.conn).ExecContext(ctx, `
		UPDATE membership_removal_requests
		SET status = $2, updated_at = $3
		WHERE id = $1
	`,
		request.ID().String(),
		string(request.Status()),
		request.UpdatedAt(),
	)
	if err != nil {
		return fmt.Errorf("update membership removal request: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check update membership removal request result: %w", err)
	}
	if affected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func scanMembershipRemovalRequest(row *sql.Row) (*domain.MembershipRemovalRequest, error) {
	request, err := scanMembershipRemovalRequestRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return request, nil
}

func scanMembershipRemovalRequestRow(scanner rowScanner) (*domain.MembershipRemovalRequest, error) {
	var (
		rawID                string
		rawMembershipID      string
		rawRequestedByUserID string
		status               string
		createdAt            time.Time
		updatedAt            time.Time
	)
	if err := scanner.Scan(&rawID, &rawMembershipID, &rawRequestedByUserID, &status, &createdAt, &updatedAt); err != nil {
		return nil, fmt.Errorf("scan membership removal request: %w", err)
	}

	id, err := uuid.Parse(rawID)
	if err != nil {
		return nil, fmt.Errorf("parse membership removal request id: %w", err)
	}
	membershipID, err := uuid.Parse(rawMembershipID)
	if err != nil {
		return nil, fmt.Errorf("parse membership id: %w", err)
	}
	requestedByUserID, err := uuid.Parse(rawRequestedByUserID)
	if err != nil {
		return nil, fmt.Errorf("parse requested_by_user_id: %w", err)
	}

	return domain.ReconstructMembershipRemovalRequest(id, membershipID, requestedByUserID, domain.MembershipRemovalRequestStatus(status), createdAt, updatedAt), nil
}
