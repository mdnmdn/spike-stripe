package db

import (
	"database/sql"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
)

// RunMigrations applies all .sql files in the given directory in lexical order.
func RunMigrations(db *sql.DB, migrationsFS fs.FS, dir string) error {
	entries, err := fs.Glob(migrationsFS, filepath.ToSlash(filepath.Join(dir, "*.sql")))
	if err != nil {
		return fmt.Errorf("glob migrations: %w", err)
	}
	sort.Strings(entries)
	for _, name := range entries {
		b, err := fs.ReadFile(migrationsFS, name)
		if err != nil {
			return fmt.Errorf("read %s: %w", name, err)
		}
		if _, err := db.Exec(string(b)); err != nil {
			return fmt.Errorf("apply %s: %w", filepath.Base(name), err)
		}
	}
	return nil
}
