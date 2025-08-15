package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ApplyMigrations discovers and executes all pending .sql migrations from a
// specified directory in a transactional manner. It tracks applied migrations
// in a dedicated `schema_migrations` table to ensure each migration is only
// run once.
func ApplyMigrations(db *sql.DB, migrationsDir string) error {
	// First, ensure the tracking table exists. This statement is idempotent.
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT NOT NULL PRIMARY KEY)`)
	if err != nil {
		return fmt.Errorf("failed to create or verify schema_migrations table: %w", err)
	}

	// Get a set of all previously applied migration versions.
	applied, err := getAppliedMigrations(db)
	if err != nil {
		return err
	}

	// Discover all available migration files on disk.
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory '%s': %w", migrationsDir, err)
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	// Begin a transaction to apply all pending migrations.
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin migration transaction: %w", err)
	}
	// Defer a rollback. If the transaction is successfully committed, this is a no-op.
	defer tx.Rollback()

	// Loop through migration files, applying any that are new.
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".sql" {
			continue
		}

		version := file.Name()
		if applied[version] {
			continue // Skip migration that has already been applied.
		}

		// Read the migration file.
		filePath := filepath.Join(migrationsDir, version)
		sqlBytes, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read migration file '%s': %w", filePath, err)
		}

		// Split the script into individual statements and execute them.
		statements := strings.Split(string(sqlBytes), ";")
		for _, stmt := range statements {
			trimmedStmt := strings.TrimSpace(stmt)
			if trimmedStmt == "" {
				continue
			}
			if _, err := tx.Exec(trimmedStmt); err != nil {
				return fmt.Errorf("failed to execute statement from migration '%s': %w", version, err)
			}
		}

		// Record the successful migration in the tracking table within the same transaction.
		_, err = tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", version)
		if err != nil {
			return fmt.Errorf("failed to record migration version '%s': %w", version, err)
		}
	}

	// All pending migrations were applied successfully, commit the transaction.
	return tx.Commit()
}

// getAppliedMigrations fetches a set of all migration versions that have been
// previously applied, according to the schema_migrations table.
func getAppliedMigrations(db *sql.DB) (map[string]bool, error) {
	rows, err := db.Query("SELECT version FROM schema_migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to query for applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("failed to scan applied migration version: %w", err)
		}
		applied[version] = true
	}
	return applied, rows.Err()
}
