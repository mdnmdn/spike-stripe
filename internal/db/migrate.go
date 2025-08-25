package db

import (
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
)

// RunMigrations applies all .sql files in the given directory in lexical order.
// It keeps track of executed scripts in a `_migrations` table and skips any that
// have already been applied (matched by filename).
func RunMigrations(db *sql.DB, migrationsFS fs.FS, dir string) error {
	// Ensure the migrations tracking table exists
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS _migrations (
			name TEXT PRIMARY KEY,
			applied_at TEXT NOT NULL DEFAULT (datetime('now'))
		);
	`); err != nil {
		return fmt.Errorf("create _migrations table: %w", err)
	}

	entries, err := fs.Glob(migrationsFS, filepath.ToSlash(filepath.Join(dir, "*.sql")))
	if err != nil {
		return fmt.Errorf("glob migrations: %w", err)
	}
	sort.Strings(entries)
	for _, name := range entries {
		base := filepath.Base(name)

		// Check if migration was already applied
		var exists int
		row := db.QueryRow("SELECT 1 FROM _migrations WHERE name = ? LIMIT 1", base)
		scanErr := row.Scan(&exists)
		if scanErr == nil {
			// already applied
			continue
		}
		if scanErr != nil && !errors.Is(scanErr, sql.ErrNoRows) {
			return fmt.Errorf("check migration %s: %w", base, scanErr)
		}

		// Read migration file
		b, err := fs.ReadFile(migrationsFS, name)
		if err != nil {
			return fmt.Errorf("read %s: %w", name, err)
		}

		// Apply inside a transaction so the insert into _migrations is atomic with the script
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("begin tx for %s: %w", base, err)
		}
		if _, err := tx.Exec(string(b)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("apply %s: %w", base, err)
		}
		if _, err := tx.Exec("INSERT INTO _migrations (name) VALUES (?)", base); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record migration %s: %w", base, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", base, err)
		}
	}
	return nil
}
