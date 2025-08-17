package database

import (
	"database/sql"
	"os"
	"testing"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	var err error
	testDB, err = Connect("file:memdb1?mode=memory&cache=shared")
	if err != nil {
		panic(err)
	}
	if err := Migrate(testDB); err != nil {
		panic(err)
	}
	code := m.Run()
	testDB.Close()
	os.Exit(code)
}

func setupTestRepo(t *testing.T) Repository {
	t.Helper()
	stmts := []string{
		"DELETE FROM pricing_info",
		"DELETE FROM pricing_updates",
		"DELETE FROM skus",
		"DELETE FROM services",
	}
	for _, stmt := range stmts {
		if _, err := testDB.Exec(stmt); err != nil {
			t.Fatalf("cleanup %s: %v", stmt, err)
		}
	}
	return NewRepository(testDB)
}
