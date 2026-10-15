package models

import (
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/alejandrolopezj/events-booking-rest-api/db"
	"github.com/alejandrolopezj/events-booking-rest-api/utils"
	_ "github.com/mattn/go-sqlite3"
)

func TestEventSavePersistsDynamicallyBuiltValues(t *testing.T) {
	database, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("could not open in-memory database: %v", err)
	}
	database.SetMaxOpenConns(1)
	defer database.Close()

	_, err = database.Exec(`
		CREATE TABLE events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT NOT NULL,
			location TEXT NOT NULL,
			date_time DATETIME NOT NULL,
			user_id INTEGER NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("could not create events table: %v", err)
	}

	previousDB := db.DB
	db.DB = database
	t.Cleanup(func() {
		db.DB = previousDB
	})

	eventTime := time.Date(2026, time.October, 13, 22, 0, 0, 0, time.UTC)
	event := &Event{
		ID:          99,
		Title:       "Go Meetup",
		Description: "A community event",
		Location:    "Online",
		DateTime:    eventTime,
		UserID:      12,
	}

	if err := event.Save(); err != nil {
		t.Fatalf("Save() returned an error: %v", err)
	}
	if event.ID == 0 || event.ID == 99 {
		t.Errorf("expected Save() to set a generated ID, got %d", event.ID)
	}

	var title, description, location string
	var savedDate time.Time
	var userID int64
	err = database.QueryRow(`
		SELECT title, description, location, date_time, user_id
		FROM events
		WHERE id = ?
	`, event.ID).Scan(&title, &description, &location, &savedDate, &userID)
	if err != nil {
		t.Fatalf("could not read saved event: %v", err)
	}
	if title != event.Title || description != event.Description || location != event.Location {
		t.Errorf("saved event text fields do not match: got %q, %q, %q", title, description, location)
	}
	if !savedDate.Equal(eventTime) {
		t.Errorf("expected saved date %v, got %v", eventTime, savedDate)
	}
	if userID != event.UserID {
		t.Errorf("expected user ID %d, got %d", event.UserID, userID)
	}
}

func TestEventPersistenceMethodsReturnUserFacingErrors(t *testing.T) {
	setupEventTestDB(t)

	event := &Event{
		ID:          insertModelTestEvent(t, 1),
		Title:       "Updated title",
		Description: "Description",
		Location:    "Online",
		DateTime:    time.Now().UTC(),
		UserID:      1,
	}
	if _, err := db.DB.Exec("INSERT INTO registrations (event_id, user_id) VALUES (?, ?)", event.ID, 2); err != nil {
		t.Fatalf("could not insert registration fixture: %v", err)
	}

	if _, err := db.DB.Exec(`
		CREATE TRIGGER reject_event_update BEFORE UPDATE ON events
		BEGIN SELECT RAISE(ABORT, 'update rejected'); END;
		CREATE TRIGGER reject_event_delete BEFORE DELETE ON events
		BEGIN SELECT RAISE(ABORT, 'delete rejected'); END;
		CREATE TRIGGER reject_registration_delete BEFORE DELETE ON registrations
		BEGIN SELECT RAISE(ABORT, 'unregister rejected'); END;
		CREATE TRIGGER reject_registration_insert BEFORE INSERT ON registrations
		BEGIN SELECT RAISE(ABORT, 'register rejected'); END;
	`); err != nil {
		t.Fatalf("could not create test triggers: %v", err)
	}

	tests := []struct {
		name string
		run  func() error
		want string
	}{
		{
			name: "save execution failure",
			run: func() error {
				_, err := db.DB.Exec("CREATE TRIGGER reject_event_insert BEFORE INSERT ON events BEGIN SELECT RAISE(ABORT, 'insert rejected'); END")
				if err != nil {
					return err
				}
				return (&Event{Title: "Rejected", Description: "Description", Location: "Online", DateTime: time.Now(), UserID: 1}).Save()
			},
			want: "Error saving the event",
		},
		{
			name: "update execution failure",
			run:  event.Update,
			want: "Error updating the event",
		},
		{
			name: "delete execution failure",
			run:  event.Delete,
			want: "Error deleting the event",
		},
		{
			name: "register execution failure",
			run:  func() error { return event.Register(2) },
			want: "Error registering for the event",
		},
		{
			name: "unregister execution failure",
			run:  func() error { return event.Unregister(2) },
			want: "Error unregistering from the event",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.run()
			if err == nil {
				t.Fatal("expected persistence error")
			}
			if err.Error() != test.want {
				t.Errorf("expected error %q, got %q", test.want, err)
			}
		})
	}
}

func TestEventReadMethodsHandleDatabaseErrors(t *testing.T) {
	database := setupEventTestDB(t)
	if _, err := database.Exec(`
		INSERT INTO events (title, description, location, date_time, user_id)
		VALUES ('Bad user', 'Description', 'Online', '2026-10-14T18:00:00Z', 'not-an-int')
	`); err != nil {
		t.Fatalf("could not insert malformed event: %v", err)
	}

	if _, err := GetAllEvents(); err == nil || err.Error() != "Error fetching events" {
		t.Errorf("expected scan error for malformed event, got %v", err)
	}
	if _, err := GetEventByID(1); err == nil || err.Error() != "Error fetching the event" {
		t.Errorf("expected scan error for malformed event, got %v", err)
	}
	if _, err := GetEventByID(999); !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected sql.ErrNoRows for a missing event, got %v", err)
	}

	if err := database.Close(); err != nil {
		t.Fatalf("could not close test database: %v", err)
	}
	if _, err := GetAllEvents(); err == nil || err.Error() != "Error fetching events" {
		t.Errorf("expected query error for closed database, got %v", err)
	}
	if _, err := GetEventByID(1); err == nil || err.Error() != "Error fetching the event" {
		t.Errorf("expected query error for closed database, got %v", err)
	}
}

func TestEventPersistenceMethodsHandleClosedDatabase(t *testing.T) {
	database := setupEventTestDB(t)
	if err := database.Close(); err != nil {
		t.Fatalf("could not close test database: %v", err)
	}

	event := &Event{
		ID:          1,
		Title:       "Event",
		Description: "Description",
		Location:    "Online",
		DateTime:    time.Now().UTC(),
		UserID:      1,
	}
	tests := []struct {
		name string
		run  func() error
		want string
	}{
		{
			name: "save prepare failure",
			run:  event.Save,
			want: "Error saving the event",
		},
		{
			name: "update prepare failure",
			run:  event.Update,
			want: "Error updating the event",
		},
		{
			name: "delete prepare failure",
			run:  event.Delete,
			want: "Error deleting the event",
		},
		{
			name: "register prepare failure",
			run:  func() error { return event.Register(2) },
			want: "Error registering for the event",
		},
		{
			name: "unregister prepare failure",
			run:  func() error { return event.Unregister(2) },
			want: "Error unregistering from the event",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.run()
			if err == nil || err.Error() != test.want {
				t.Errorf("expected error %q, got %v", test.want, err)
			}
		})
	}
}

func setupEventTestDB(t *testing.T) *sql.DB {
	t.Helper()

	database, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("could not open in-memory database: %v", err)
	}
	database.SetMaxOpenConns(1)

	_, err = database.Exec(`
		CREATE TABLE users (id INTEGER PRIMARY KEY);
		INSERT INTO users (id) VALUES (1);
		CREATE TABLE events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT NOT NULL,
			location TEXT NOT NULL,
			date_time DATETIME NOT NULL,
			user_id INTEGER NOT NULL
		);
		CREATE TABLE registrations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			event_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL
		);
	`)
	if err != nil {
		database.Close()
		t.Fatalf("could not create test schema: %v", err)
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

func insertModelTestEvent(t *testing.T, userID int64) int64 {
	t.Helper()

	result, err := db.DB.Exec(`
		INSERT INTO events (title, description, location, date_time, user_id)
		VALUES ('Existing', 'Description', 'Online', ?, ?)
	`, time.Now().UTC(), userID)
	if err != nil {
		t.Fatalf("could not insert test event: %v", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("could not get event ID: %v", err)
	}
	return id
}
