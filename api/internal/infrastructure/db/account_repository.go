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

// AccountRepository はPostgreSQLを用いたAccountの永続化を行う。
type AccountRepository struct {
	conn *sql.DB
}

// NewAccountRepository はAccountRepositoryを生成する。
func NewAccountRepository(conn *sql.DB) *AccountRepository {
	return &AccountRepository{conn: conn}
}

// Create はAccountを新規保存する。
func (r *AccountRepository) Create(ctx context.Context, account *domain.Account) error {
	_, err := executorFromContext(ctx, r.conn).ExecContext(ctx, `
		INSERT INTO accounts (id, provider_id, account_type, name, role, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`,
		account.ID().String(),
		account.ProviderID(),
		string(account.AccountType()),
		account.Name(),
		string(account.Role()),
		string(account.Status()),
		account.CreatedAt(),
		account.UpdatedAt(),
	)
	if err != nil {
		return fmt.Errorf("insert account: %w", err)
	}

	return nil
}

// Get はIDを指定してAccountを1件取得する。
func (r *AccountRepository) Get(ctx context.Context, id uuid.UUID) (*domain.Account, error) {
	row := executorFromContext(ctx, r.conn).QueryRowContext(ctx, `
		SELECT id, provider_id, account_type, name, role, status, created_at, updated_at
		FROM accounts
		WHERE id = $1
	`, id.String())

	return scanAccount(row)
}

// GetByProviderID はProviderIDを指定してAccountを1件取得する。
func (r *AccountRepository) GetByProviderID(ctx context.Context, providerID string) (*domain.Account, error) {
	row := executorFromContext(ctx, r.conn).QueryRowContext(ctx, `
		SELECT id, provider_id, account_type, name, role, status, created_at, updated_at
		FROM accounts
		WHERE provider_id = $1
	`, providerID)

	return scanAccount(row)
}

// Update はAccountの内容(role, status等)を更新する。
func (r *AccountRepository) Update(ctx context.Context, account *domain.Account) error {
	result, err := executorFromContext(ctx, r.conn).ExecContext(ctx, `
		UPDATE accounts
		SET role = $2, status = $3, updated_at = $4
		WHERE id = $1
	`,
		account.ID().String(),
		string(account.Role()),
		string(account.Status()),
		account.UpdatedAt(),
	)
	if err != nil {
		return fmt.Errorf("update account: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check update account result: %w", err)
	}
	if affected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// ListByStatus は指定したstatusのAccountを一覧取得する。
func (r *AccountRepository) ListByStatus(ctx context.Context, status domain.AccountStatus) ([]*domain.Account, error) {
	rows, err := executorFromContext(ctx, r.conn).QueryContext(ctx, `
		SELECT id, provider_id, account_type, name, role, status, created_at, updated_at
		FROM accounts
		WHERE status = $1
		ORDER BY created_at
	`, string(status))
	if err != nil {
		return nil, fmt.Errorf("select accounts: %w", err)
	}
	defer rows.Close()

	var accounts []*domain.Account
	for rows.Next() {
		account, err := scanAccountRow(rows)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate accounts: %w", err)
	}

	return accounts, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanAccount(row *sql.Row) (*domain.Account, error) {
	account, err := scanAccountRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return account, nil
}

func scanAccountRow(scanner rowScanner) (*domain.Account, error) {
	var (
		rawID       string
		providerID  string
		accountType string
		name        sql.NullString
		role        string
		status      string
		createdAt   time.Time
		updatedAt   time.Time
	)
	if err := scanner.Scan(&rawID, &providerID, &accountType, &name, &role, &status, &createdAt, &updatedAt); err != nil {
		return nil, fmt.Errorf("scan account: %w", err)
	}

	id, err := uuid.Parse(rawID)
	if err != nil {
		return nil, fmt.Errorf("parse account id: %w", err)
	}

	var namePtr *string
	if name.Valid {
		namePtr = &name.String
	}

	return domain.ReconstructAccount(id, providerID, domain.AccountType(accountType), namePtr, domain.AccountRole(role), domain.AccountStatus(status), createdAt, updatedAt), nil
}
