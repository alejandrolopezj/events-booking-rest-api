package db

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestInitDBCreatesTables(t *testing.T) {
	t.Chdir(t.TempDir())

	if err := InitDB(); err != nil {
		t.Fatalf("InitDB() returned an error: %v", err)
	}
	if DB == nil {
		t.Fatal("InitDB() did not initialize DB")
	}

	t.Cleanup(func() {
		DB.Close()
		DB = nil
	})

	assertTablesExist(t, DB)
}

func TestInitDBReturnsPingError(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.Mkdir("api.db", 0o755); err != nil {
		t.Fatalf("could not create directory at database path: %v", err)
	}

	if err := InitDB(); err == nil {
		t.Fatal("InitDB() returned nil when the database path could not be opened")
	}
}

func TestCreateTablesIsIdempotent(t *testing.T) {
	database, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("could not open in-memory database: %v", err)
	}
	defer database.Close()
	database.SetMaxOpenConns(1)

	if err := createTables(database); err != nil {
		t.Fatalf("first createTables() call returned an error: %v", err)
	}
	if err := createTables(database); err != nil {
		t.Fatalf("second createTables() call returned an error: %v", err)
	}

	assertTablesExist(t, database)
}

func TestCreateTablesReturnsExecutionError(t *testing.T) {
	database, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("could not open in-memory database: %v", err)
	}
	if err := database.Close(); err != nil {
		t.Fatalf("could not close database: %v", err)
	}

	if err := createTables(database); err == nil {
		t.Fatal("createTables() returned nil for a closed database")
	}
}

func assertTablesExist(t *testing.T, database *sql.DB) {
	t.Helper()

	rows, err := database.Query(`
		SELECT name
		FROM sqlite_master
		WHERE type = 'table'
			AND name IN ('users', 'events', 'registrations')
	`)
	if err != nil {
		t.Fatalf("could not query database tables: %v", err)
	}
	defer rows.Close()

	found := make(map[string]bool)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("could not scan table name: %v", err)
		}
		found[name] = true
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("error iterating over database tables: %v", err)
	}

	for _, name := range []string{"users", "events", "registrations"} {
		if !found[name] {
			t.Errorf("expected table %q to exist", name)
		}
	}
}
