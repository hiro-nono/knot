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

// MembershipRepository はPostgreSQLを用いたMembershipの永続化を行う。
type MembershipRepository struct {
	conn *sql.DB
}

// NewMembershipRepository はMembershipRepositoryを生成する。
func NewMembershipRepository(conn *sql.DB) *MembershipRepository {
	return &MembershipRepository{conn: conn}
}

// Create はMembershipを新規保存する。
func (r *MembershipRepository) Create(ctx context.Context, membership *domain.Membership) error {
	_, err := executorFromContext(ctx, r.conn).ExecContext(ctx, `
		INSERT INTO memberships (id, account_id, user_id, role, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`,
		membership.ID().String(),
		membership.AccountID().String(),
		membership.UserID().String(),
		string(membership.Role()),
		string(membership.Status()),
		membership.CreatedAt(),
		membership.UpdatedAt(),
	)
	if err != nil {
		return fmt.Errorf("insert membership: %w", err)
	}

	return nil
}

// Get はIDを指定してMembershipを1件取得する。
func (r *MembershipRepository) Get(ctx context.Context, id uuid.UUID) (*domain.Membership, error) {
	row := executorFromContext(ctx, r.conn).QueryRowContext(ctx, `
		SELECT id, account_id, user_id, role, status, created_at, updated_at
		FROM memberships
		WHERE id = $1
	`, id.String())

	return scanMembership(row)
}

// GetByAccountAndUser はaccountID・userIDに対応するMembershipを取得する
// (statusを問わない)。
func (r *MembershipRepository) GetByAccountAndUser(ctx context.Context, accountID, userID uuid.UUID) (*domain.Membership, error) {
	row := executorFromContext(ctx, r.conn).QueryRowContext(ctx, `
		SELECT id, account_id, user_id, role, status, created_at, updated_at
		FROM memberships
		WHERE account_id = $1 AND user_id = $2
	`, accountID.String(), userID.String())

	return scanMembership(row)
}

// GetOwnerByAccountID はaccountIDに対応するowner roleのMembershipを取得する。
func (r *MembershipRepository) GetOwnerByAccountID(ctx context.Context, accountID uuid.UUID) (*domain.Membership, error) {
	row := executorFromContext(ctx, r.conn).QueryRowContext(ctx, `
		SELECT id, account_id, user_id, role, status, created_at, updated_at
		FROM memberships
		WHERE account_id = $1 AND role = $2
	`, accountID.String(), string(domain.MembershipRoleOwner))

	return scanMembership(row)
}

// ListByAccountID はaccountIDに対応するMembershipを一覧取得する。
func (r *MembershipRepository) ListByAccountID(ctx context.Context, accountID uuid.UUID) ([]*domain.Membership, error) {
	rows, err := executorFromContext(ctx, r.conn).QueryContext(ctx, `
		SELECT id, account_id, user_id, role, status, created_at, updated_at
		FROM memberships
		WHERE account_id = $1
		ORDER BY created_at
	`, accountID.String())
	if err != nil {
		return nil, fmt.Errorf("select memberships: %w", err)
	}
	defer rows.Close()

	var memberships []*domain.Membership
	for rows.Next() {
		membership, err := scanMembershipRow(rows)
		if err != nil {
			return nil, err
		}
		memberships = append(memberships, membership)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate memberships: %w", err)
	}

	return memberships, nil
}

// Update はMembershipのrole・status等を更新する。
func (r *MembershipRepository) Update(ctx context.Context, membership *domain.Membership) error {
	result, err := executorFromContext(ctx, r.conn).ExecContext(ctx, `
		UPDATE memberships
		SET role = $2, status = $3, updated_at = $4
		WHERE id = $1
	`,
		membership.ID().String(),
		string(membership.Role()),
		string(membership.Status()),
		membership.UpdatedAt(),
	)
	if err != nil {
		return fmt.Errorf("update membership: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check update membership result: %w", err)
	}
	if affected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func scanMembership(row *sql.Row) (*domain.Membership, error) {
	membership, err := scanMembershipRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return membership, nil
}

func scanMembershipRow(scanner rowScanner) (*domain.Membership, error) {
	var (
		rawID        string
		rawAccountID string
		rawUserID    string
		role         string
		status       string
		createdAt    time.Time
		updatedAt    time.Time
	)
	if err := scanner.Scan(&rawID, &rawAccountID, &rawUserID, &role, &status, &createdAt, &updatedAt); err != nil {
		return nil, fmt.Errorf("scan membership: %w", err)
	}

	id, err := uuid.Parse(rawID)
	if err != nil {
		return nil, fmt.Errorf("parse membership id: %w", err)
	}
	accountID, err := uuid.Parse(rawAccountID)
	if err != nil {
		return nil, fmt.Errorf("parse account id: %w", err)
	}
	userID, err := uuid.Parse(rawUserID)
	if err != nil {
		return nil, fmt.Errorf("parse user id: %w", err)
	}

	return domain.ReconstructMembership(id, accountID, userID, domain.MembershipRole(role), domain.MembershipStatus(status), createdAt, updatedAt), nil
}
