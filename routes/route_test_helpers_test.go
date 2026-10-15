package routes

import (
	"database/sql"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alejandrolopezj/events-booking-rest-api/db"
	"github.com/alejandrolopezj/events-booking-rest-api/utils"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

func setupRouteTest(t *testing.T) {
	t.Helper()

	previousMode := gin.Mode()
	gin.SetMode(gin.TestMode)
	t.Cleanup(func() {
		gin.SetMode(previousMode)
	})

	previousLogger := utils.Logger
	utils.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	t.Cleanup(func() {
		utils.Logger = previousLogger
	})

	database, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("could not open in-memory database: %v", err)
	}
	database.SetMaxOpenConns(1)

	_, err = database.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL
		);
		CREATE TABLE events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT NOT NULL,
			location TEXT NOT NULL,
			date_time DATETIME NOT NULL,
			user_id INTEGER NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);
		CREATE TABLE registrations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			event_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			FOREIGN KEY (event_id) REFERENCES events(id),
			FOREIGN KEY (user_id) REFERENCES users(id)
		);
	`)
	if err != nil {
		database.Close()
		t.Fatalf("could not create test tables: %v", err)
	}

	previousDB := db.DB
	db.DB = database
	t.Cleanup(func() {
		db.DB = previousDB
		if err := database.Close(); err != nil {
			t.Errorf("could not close test database: %v", err)
		}
	})
}

func performRequest(handler http.Handler, method, target, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, target, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}

func handlerWithUserID(handler gin.HandlerFunc, userID int64) gin.HandlerFunc {
	return func(context *gin.Context) {
		context.Set("claims", &utils.JWTClaims{UserID: userID})
		handler(context)
	}
}

func insertTestEvent(t *testing.T, userID int64) int64 {
	t.Helper()

	result, err := db.DB.Exec(`
		INSERT INTO events (title, description, location, date_time, user_id)
		VALUES (?, ?, ?, ?, ?)
	`, "Test event", "Test description", "Online", "2026-10-14T18:00:00Z", userID)
	if err != nil {
		t.Fatalf("could not insert test event: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("could not get test event ID: %v", err)
	}
	return id
}
