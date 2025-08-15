package database

import (
	"database/sql"
	"os"
	"strings"
	"testing"

	_ "github.com/tursodatabase/go-libsql"
)

// setupTestDB is a helper function to create a temporary database for testing.
// It returns a database connection and a cleanup function to close the DB and remove the file.
func setupTestDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()

	tmpfile, err := os.CreateTemp("", "test-db-*.db")
	if err != nil {
		t.Fatalf("Failed to create temporary file for test database: %v", err)
	}
	dbPath := "file:" + tmpfile.Name()

	db, err := sql.Open("libsql", dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database at %s: %v", dbPath, err)
	}

	cleanup := func() {
		db.Close()
		os.Remove(tmpfile.Name())
	}

	return db, cleanup
}

// TestApplyMigrations_Success verifies that migrations run successfully and create the expected schema.
func TestApplyMigrations_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// This is the function we are testing.
	err := ApplyMigrations(db, "migrations")
	if err != nil {
		t.Fatalf("ApplyMigrations failed unexpectedly: %v", err)
	}

	// 1. Verify all expected tables were created.
	expectedTables := []string{"services", "skus", "pricing_info", "pricing_updates", "schema_migrations"}
	for _, tableName := range expectedTables {
		var name string
		query := "SELECT name FROM sqlite_master WHERE type='table' AND name = ?"
		err := db.QueryRow(query, tableName).Scan(&name)
		if err != nil {
			if err == sql.ErrNoRows {
				t.Errorf("Verification failed: table '%s' was not created", tableName)
			} else {
				t.Errorf("Verification failed: error checking for table '%s': %v", tableName, err)
			}
		}
	}

	// 2. Verify the migration was tracked.
	var trackedVersion string
	err = db.QueryRow("SELECT version FROM schema_migrations WHERE version = ?", "0001_initial_schema.sql").Scan(&trackedVersion)
	if err != nil {
		t.Errorf("Verification failed: migration '0001_initial_schema.sql' was not tracked: %v", err)
	}

	// 3. Verify the new schema of the 'skus' table.
	rows, err := db.Query("PRAGMA table_info(skus)")
	if err != nil {
		t.Fatalf("Failed to get table info for skus: %v", err)
	}
	defer rows.Close()

	expectedColumns := map[string]string{
		"category":        "BLOB",
		"service_regions": "BLOB",
		"geo_taxonomy":    "BLOB",
		"sku_name":        "TEXT",
	}

	for rows.Next() {
		var (
			cid        int
			name       string
			dataType   string
			notnull    int
			dfltValue  interface{}
			pk         int
		)
		if err := rows.Scan(&cid, &name, &dataType, &notnull, &dfltValue, &pk); err != nil {
			t.Fatalf("Failed to scan table info row: %v", err)
		}

		if expectedType, ok := expectedColumns[name]; ok {
			if strings.ToUpper(dataType) != expectedType {
				t.Errorf("For column '%s', expected type '%s', but got '%s'", name, expectedType, dataType)
			}
			delete(expectedColumns, name) // Mark as found
		}
	}

	if len(expectedColumns) > 0 {
		for name := range expectedColumns {
			t.Errorf("Expected column '%s' was not found in 'skus' table", name)
		}
	}
}

// TestNotNullConstraints verifies that the NOT NULL constraints are enforced.
func TestNotNullConstraints(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// We must apply migrations before we can test constraints.
	err := ApplyMigrations(db, "migrations")
	if err != nil {
		t.Fatalf("ApplyMigrations failed unexpectedly: %v", err)
	}

	// Need to insert a valid service before testing skus that depend on it.
	_, err = db.Exec("INSERT INTO services (service_id, display_name) VALUES ('test-service', 'Test Service')")
	if err != nil {
		t.Fatalf("Failed to insert prerequisite service: %v", err)
	}

	testCases := []struct {
		name    string
		sql     string
		args    []interface{}
	}{
		{"services.display_name", "INSERT INTO services (service_id, display_name) VALUES (?, ?)", []interface{}{"s1", nil}},
		{"skus.service_id", "INSERT INTO skus (sku_id, service_id, sku_name, description) VALUES (?, ?, ?, ?)", []interface{}{"sku1", nil, "name", "desc"}},
		{"skus.sku_name", "INSERT INTO skus (sku_id, service_id, sku_name, description) VALUES (?, ?, ?, ?)", []interface{}{"sku1", "test-service", nil, "desc"}},
		{"skus.description", "INSERT INTO skus (sku_id, service_id, sku_name, description) VALUES (?, ?, ?, ?)", []interface{}{"sku1", "test-service", "name", nil}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := db.Exec(tc.sql, tc.args...)
			if err == nil {
				t.Errorf("Expected an error for NOT NULL constraint on %s, but got nil", tc.name)
			} else if !strings.Contains(strings.ToLower(err.Error()), "not null constraint failed") {
				t.Errorf("Expected a 'NOT NULL constraint failed' error, but got: %v", err)
			}
		})
	}
}
