package migrate

import (
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"sort"
	"strings"
)

// isAlreadyExistsError checks if a PostgreSQL error indicates the object already exists.
// This handles cases where migrations were applied externally (e.g., docker-entrypoint-initdb.d).
func isAlreadyExistsError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "already exists") ||
		strings.Contains(msg, "42P07") || // duplicate_table
		strings.Contains(msg, "42710") // duplicate_object
}

// Run executes all SQL migration files from the given filesystem in lexicographic order.
// It tracks applied migrations in a schema_migrations table to avoid re-running them.
// If a migration fails with "already exists", it is marked as applied and execution continues.
// This makes it safe to use alongside docker-entrypoint-initdb.d or other migration systems.
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

	// Detect if database was bootstrapped externally (e.g., docker-entrypoint-initdb.d)
	// If core tables exist but no migrations are tracked, backfill schema_migrations
	if len(applied) == 0 {
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_name = 'tenants')").Scan(&exists)
		if err == nil && exists {
			log.Printf("Database tables exist but no migrations tracked — backfilling schema_migrations")
			backfillExisting(db, migrationsFS, applied)
		}
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
	newlyApplied := 0
	skippedExisting := 0
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

			// If tables/indexes already exist (e.g., from docker-entrypoint-initdb.d),
			// mark the migration as applied and continue
			if isAlreadyExistsError(err) {
				log.Printf("Migration %s: objects already exist, marking as applied", file)
				if _, markErr := db.Exec("INSERT INTO schema_migrations (version) VALUES ($1) ON CONFLICT DO NOTHING", file); markErr != nil {
					log.Printf("WARNING: could not mark %s as applied: %v", file, markErr)
				}
				skippedExisting++
				continue
			}

			return fmt.Errorf("execute migration %s: %w", file, err)
		}

		if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", file); err != nil {
			tx.Rollback()
			return fmt.Errorf("record migration %s: %w", file, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", file, err)
		}

		newlyApplied++
		log.Printf("Migration %s applied successfully", file)
	}

	log.Printf("Migrations complete: %d total files, %d already tracked, %d newly applied, %d marked as existing",
		len(files), len(applied), newlyApplied, skippedExisting)
	return nil
}

// backfillExisting marks all known migration files as applied when the database
// was bootstrapped externally (e.g., via docker-entrypoint-initdb.d).
// It detects which migrations are already applied by checking for their side effects.
func backfillExisting(db *sql.DB, migrationsFS fs.FS, applied map[string]bool) {
	entries, err := fs.ReadDir(migrationsFS, ".")
	if err != nil {
		return
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	// Check for evidence of each migration being applied
	// We use a heuristic: check if a table/column created by each migration exists
	checks := map[string]string{
		"001_init.sql":                "SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_name = 'tenants')",
		"002_v2_modules.sql":          "SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_name = 'vouchers')",
		"003_v3_robustify.sql":        "SELECT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_name = 'shifts' AND column_name = 'seller_id')",
		"004_v4_kiosk_ux.sql":         "SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_name = 'kiosk_sessions')",
		"005_v5_kiosk_monitoring.sql":  "SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_name = 'kiosk_diagnostics')",
		"006_v6_stabilization.sql":    "SELECT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'failed_attempts')",
	}

	for _, file := range files {
		if applied[file] {
			continue
		}

		// Check if migration was already applied externally
		checkSQL, hasCheck := checks[file]
		if hasCheck {
			var exists bool
			if err := db.QueryRow(checkSQL).Scan(&exists); err != nil || !exists {
				continue // Not applied yet
			}
		} else {
			continue // Unknown migration, don't assume it was applied
		}

		// Mark as applied
		if _, err := db.Exec("INSERT INTO schema_migrations (version) VALUES ($1) ON CONFLICT DO NOTHING", file); err != nil {
			log.Printf("WARNING: could not backfill %s: %v", file, err)
			continue
		}
		applied[file] = true
		log.Printf("Backfilled migration: %s (already applied externally)", file)
	}
}
