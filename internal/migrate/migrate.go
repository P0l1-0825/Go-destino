package migrate

import (
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"sort"
	"strings"
)

// Run executes all SQL migration files from the given filesystem in lexicographic order.
// It tracks applied migrations in a schema_migrations table to avoid re-running them.
func Run(db *sql.DB, migrationsFS fs.FS) error {
	// Create tracking table if not exists
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMPTZ DEFAULT NOW()
		)
	`); err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	// Read applied migrations
	applied := make(map[string]bool)
	rows, err := db.Query("SELECT version FROM schema_migrations")
	if err != nil {
		return fmt.Errorf("query schema_migrations: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return fmt.Errorf("scan version: %w", err)
		}
		applied[v] = true
	}

	// Discover migration files
	entries, err := fs.ReadDir(migrationsFS, ".")
	if err != nil {
		return fmt.Errorf("read migrations directory: %w", err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	// Run pending migrations
	for _, file := range files {
		if applied[file] {
			continue
		}

		data, err := fs.ReadFile(migrationsFS, file)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", file, err)
		}

		log.Printf("Applying migration: %s", file)

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("begin tx for %s: %w", file, err)
		}

		if _, err := tx.Exec(string(data)); err != nil {
			tx.Rollback()
			return fmt.Errorf("execute migration %s: %w", file, err)
		}

		if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", file); err != nil {
			tx.Rollback()
			return fmt.Errorf("record migration %s: %w", file, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", file, err)
		}

		log.Printf("Migration %s applied successfully", file)
	}

	log.Printf("Migrations complete (%d total, %d already applied)", len(files), len(applied))
	return nil
}
