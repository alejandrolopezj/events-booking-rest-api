package models

import (
	"database/sql"
	"io"
	"log/slog"
	"testing"

	"github.com/alejandrolopezj/events-booking-rest-api/db"
	"github.com/alejandrolopezj/events-booking-rest-api/utils"
	_ "github.com/mattn/go-sqlite3"
)

func TestUserSaveStoresHashedPasswordAndGeneratedID(t *testing.T) {
	database := newUserTestDB(t)
	user := &User{
		Email:    "user@example.com",
		Password: "correct-horse-battery-staple",
	}

	if err := user.Save(); err != nil {
		t.Fatalf("Save() returned an error: %v", err)
	}
	if user.ID == 0 {
		t.Fatal("Save() did not set the generated user ID")
	}

	var storedPassword string
	err := database.QueryRow(
		"SELECT password FROM users WHERE id = ?",
		user.ID,
	).Scan(&storedPassword)
	if err != nil {
		t.Fatalf("could not read saved user password: %v", err)
	}
	if storedPassword == user.Password {
		t.Fatal("Save() stored the password in plain text")
	}
	if !utils.CheckPasswordHash(user.Password, storedPassword) {
		t.Error("stored password hash does not match the original password")
	}
}

func TestUserValidateCredentials(t *testing.T) {
	newUserTestDB(t)
	user := &User{
		Email:    "user@example.com",
		Password: "correct-password",
	}
	if err := user.Save(); err != nil {
		t.Fatalf("Save() returned an error: %v", err)
	}

	t.Run("accepts correct password", func(t *testing.T) {
		candidate := &User{Email: user.Email}
		if err := candidate.ValidateCredentials(user.Password); err != nil {
			t.Fatalf("ValidateCredentials() returned an error: %v", err)
		}
		if candidate.ID != user.ID {
			t.Errorf("expected user ID %d, got %d", user.ID, candidate.ID)
		}
	})

	t.Run("rejects incorrect password", func(t *testing.T) {
		candidate := &User{Email: user.Email}
		err := candidate.ValidateCredentials("wrong-password")
		if err == nil {
			t.Fatal("ValidateCredentials() accepted an incorrect password")
		}
		if err.Error() != "Invalid user credentials" {
			t.Errorf("expected a user-friendly credential error, got %q", err)
		}
	})

	t.Run("rejects unknown email", func(t *testing.T) {
		candidate := &User{Email: "missing@example.com"}
		err := candidate.ValidateCredentials("any-password")
		if err == nil {
			t.Fatal("ValidateCredentials() accepted an unknown email")
		}
		if err.Error() != "Invalid user credentials" {
			t.Errorf("expected a user-friendly credential error, got %q", err)
		}
	})
}

func TestUserSaveReturnsErrors(t *testing.T) {
	database := newUserTestDB(t)

	user := &User{Email: "user@example.com", Password: "password"}
	if err := user.Save(); err != nil {
		t.Fatalf("initial Save() returned an error: %v", err)
	}
	if err := (&User{Email: user.Email, Password: "another-password"}).Save(); err == nil || err.Error() != "Error saving the user" {
		t.Errorf("expected duplicate email error, got %v", err)
	}

	if err := (&User{
		Email:    "long-password@example.com",
		Password: string(make([]byte, 73)),
	}).Save(); err == nil || err.Error() != "Error saving the user" {
		t.Errorf("expected password hashing error, got %v", err)
	}

	if err := database.Close(); err != nil {
		t.Fatalf("could not close test database: %v", err)
	}
	if err := (&User{Email: "closed@example.com", Password: "password"}).Save(); err == nil || err.Error() != "Error saving the user" {
		t.Errorf("expected closed database error, got %v", err)
	}
}

func newUserTestDB(t *testing.T) *sql.DB {
	t.Helper()

	database, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("could not open in-memory database: %v", err)
	}
	database.SetMaxOpenConns(1)

	if _, err := database.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL
		)
	`); err != nil {
		database.Close()
		t.Fatalf("could not create users table: %v", err)
	}

	previousDB := db.DB
	db.DB = database
	t.Cleanup(func() {
		db.DB = previousDB
		if err := database.Close(); err != nil {
			t.Errorf("could not close test database: %v", err)
		}
	})

	previousLogger := utils.Logger
	utils.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	t.Cleanup(func() {
		utils.Logger = previousLogger
	})

	return database
}
