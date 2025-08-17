package sync

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"mcp-server/internal/database"
)

var (
	testDB   *sql.DB
	testRepo database.Repository
)

func TestMain(m *testing.M) {
	var err error
	testDB, err = database.Connect("file:memdb1?mode=memory&cache=shared")
	if err != nil {
		panic(err)
	}
	if err := database.Migrate(testDB); err != nil {
		panic(err)
	}
	testRepo = database.NewRepository(testDB)
	code := m.Run()
	testDB.Close()
	os.Exit(code)
}

func cleanDB(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	stmts := []string{
		"DELETE FROM pricing_info",
		"DELETE FROM pricing_updates",
		"DELETE FROM skus",
		"DELETE FROM services",
	}
	for _, stmt := range stmts {
		if _, err := testDB.ExecContext(ctx, stmt); err != nil {
			t.Fatalf("cleanup %s: %v", stmt, err)
		}
	}
}
