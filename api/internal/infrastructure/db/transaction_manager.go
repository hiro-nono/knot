package db

import (
	"context"
	"database/sql"
	"fmt"
)

type txContextKey struct{}

// executor はsql.DBとsql.Txの両方が満たす、クエリ実行に必要な最小インターフェース。
// 各Repositoryはこれを介してDBにアクセスすることで、
// 通常の接続でもトランザクション内でも同じコードで動作する。
type executor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// executorFromContext はctxにトランザクションが積まれていればそれを、
// なければdefaultConnを返す。
func executorFromContext(ctx context.Context, defaultConn *sql.DB) executor {
	if tx, ok := ctx.Value(txContextKey{}).(*sql.Tx); ok {
		return tx
	}
	return defaultConn
}

// TransactionManager はPostgreSQLのトランザクションを管理する。
type TransactionManager struct {
	conn *sql.DB
}

// NewTransactionManager はTransactionManagerを生成する。
func NewTransactionManager(conn *sql.DB) *TransactionManager {
	return &TransactionManager{conn: conn}
}

// WithinTransaction はfnを1つのトランザクション内で実行する。
// fnがエラーを返す、またはpanicした場合はロールバックし、
// そうでなければコミットする。
func (m *TransactionManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) (err error) {
	tx, err := m.conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(context.WithValue(ctx, txContextKey{}, tx)); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("rollback transaction: %v (original error: %w)", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
