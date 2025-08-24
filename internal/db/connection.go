package db

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tursodatabase/libsql-client-go/libsql"
	_ "modernc.org/sqlite"
)

// DefaultLocalDBPath is used when no DB_PATH is provided.
const DefaultLocalDBPath = "_data/db-spike-strip.sqlite3"

// NewConnection returns a *sql.DB connected either to Turso (when TURSO_DATABASE_URL is set)
// or to a local SQLite database file. You can override the local path with DB_PATH.
func NewConnection() (*sql.DB, error) {
	if os.Getenv("TURSO_DATABASE_URL") != "" {
		return NewTursoConnection()
	}
	path := os.Getenv("DB_PATH")
	if path == "" {
		path = DefaultLocalDBPath
	}
	return NewLocalSQLite(path)
}

// NewLocalSQLite opens a SQLite database using the pure-Go modernc driver.
func NewLocalSQLite(path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}
	// Enable foreign keys and WAL for better concurrency
	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(ON)&_pragma=journal_mode(WAL)", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// NewTursoConnection opens a connection to Turso/LibSQL using the official client connector.
func NewTursoConnection() (*sql.DB, error) {
	url := os.Getenv("TURSO_DATABASE_URL")
	token := os.Getenv("TURSO_AUTH_TOKEN")
	if url == "" {
		return nil, errors.New("TURSO_DATABASE_URL is required for Turso connection")
	}
	connector, err := libsql.NewConnector(url, libsql.WithAuthToken(token))
	if err != nil {
		return nil, err
	}
	return sql.OpenDB(connector), nil
}

// NewTestConnection creates a new in-memory SQLite connection for testing
func NewTestConnection() (*sql.DB, error) {
	dsn := ":memory:"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	// Run migrations for testing
	migrations := os.DirFS("../../db/migrations")
	if err := RunMigrations(db, migrations, ""); err != nil {
		// Try alternative path for when tests are run from project root
		migrations = os.DirFS("db/migrations")
		if err := RunMigrations(db, migrations, ""); err != nil {
			return nil, fmt.Errorf("failed to run test migrations: %w", err)
		}
	}

	return db, nil
}
