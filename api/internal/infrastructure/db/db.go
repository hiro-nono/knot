// Package db はPostgreSQLへの接続とマイグレーション実行を提供する。
package db

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// New は指定されたDSNでPostgreSQLの接続プールを開き、
// pingによって疎通を確認する。
func New(dsn string) (*sql.DB, error) {
	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return conn, nil
}

// Migrate はmigrationsPath配下にある未適用の.sqlマイグレーションを
// すべてconnに適用する。
func Migrate(conn *sql.DB, migrationsPath string) error {
	driver, err := postgres.WithInstance(conn, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://"+migrationsPath, "postgres", driver)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("run migrations: %w", err)
	}

	return nil
}
