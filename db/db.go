// Package db centralizes the initialization and access to the API's SQLite
// database.
package db

import (
	"database/sql"
	"embed"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema/*.sql
var schemaFiles embed.FS

// DB is the global SQLite database connection used by the API.
var DB *sql.DB

// InitDB opens the database connection, verifies connectivity to SQLite, and
// creates the required tables for the application.
func InitDB() error {
	database, err := sql.Open("sqlite3", "api.db")
	if err != nil {
		return fmt.Errorf("Opening database: %w", err)
	}

	database.SetMaxOpenConns(10)
	database.SetMaxIdleConns(5)

	if err := database.Ping(); err != nil {
		database.Close()
		return fmt.Errorf("Connecting to database: %w", err)
	}

	if err := createTables(database); err != nil {
		database.Close()
		return err
	}

	DB = database
	return nil
}

// createTables creates the core application tables.
// It is executed during database initialization.
func createTables(database *sql.DB) error {
	files := []string{
		"users.sql",
		"events.sql",
		"registrations.sql",
	}

	for _, file := range files {
		statement, err := schemaFiles.ReadFile("schema/" + file)
		if err != nil {
			return fmt.Errorf("Reading database schema %q: %w", file, err)
		}
		if _, err := database.Exec(string(statement)); err != nil {
			return fmt.Errorf("Executing database schema %q: %w", file, err)
		}
	}

	return nil
}
