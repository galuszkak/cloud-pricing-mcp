package database

import (
	"database/sql"
	"fmt"
	"testing"
)

func TestMigrateCreatesTables(t *testing.T) {
	db, err := Connect("file:memdb1?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()

	if err := Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	tables := []string{"services", "skus", "pricing_info", "pricing_updates"}
	for _, tbl := range tables {
		var name string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", tbl).Scan(&name)
		if err != nil {
			t.Fatalf("table %s not created: %v", tbl, err)
		}
	}
}

func TestMigrateColumnTypes(t *testing.T) {
	db, err := Connect("file:memdb1?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()

	if err := Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	checks := map[string]map[string]string{
		"skus": {
			"sku_name":        "TEXT",
			"category":        "BLOB",
			"service_regions": "BLOB",
			"geo_taxonomy":    "BLOB",
		},
		"pricing_info": {
			"tiered_rates": "BLOB",
		},
	}
	for table, cols := range checks {
		rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
		if err != nil {
			t.Fatalf("pragma table_info %s: %v", table, err)
		}
		defer rows.Close()
		types := map[string]string{}
		for rows.Next() {
			var cid int
			var name, ctype string
			var notnull, pk int
			var dflt sql.NullString
			if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
				t.Fatalf("scan %s info: %v", table, err)
			}
			types[name] = ctype
		}
		if err := rows.Err(); err != nil {
			t.Fatalf("iterate %s info: %v", table, err)
		}
		for col, want := range cols {
			got, ok := types[col]
			if !ok {
				t.Errorf("table %s missing column %s", table, col)
				continue
			}
			if got != want {
				t.Errorf("table %s column %s type %s, want %s", table, col, got, want)
			}
		}
	}
}
