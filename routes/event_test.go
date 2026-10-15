package routes

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	"github.com/alejandrolopezj/events-booking-rest-api/db"
	"github.com/gin-gonic/gin"
)

func TestGetEventsReturnsStoredEvents(t *testing.T) {
	setupRouteTest(t)
	insertTestEvent(t, 7)

	server := gin.New()
	server.GET("/events", getEvents)
	response := performRequest(server, http.MethodGet, "/events", "")
	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body)
	}

	var events []map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &events); err != nil {
		t.Fatalf("could not decode event response: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected one event, got %d", len(events))
	}
}

func TestGetEventRejectsInvalidIDAndReturnsEvent(t *testing.T) {
	setupRouteTest(t)
	eventID := insertTestEvent(t, 7)

	server := gin.New()
	server.GET("/events/:id", getEvent)

	invalidResponse := performRequest(server, http.MethodGet, "/events/not-an-id", "")
	if invalidResponse.Code != http.StatusBadRequest {
		t.Errorf("expected invalid ID status %d, got %d", http.StatusBadRequest, invalidResponse.Code)
	}

	response := performRequest(server, http.MethodGet, "/events/"+strconv.FormatInt(eventID, 10), "")
	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body)
	}
}

func TestCreateEventUsesAuthenticatedUser(t *testing.T) {
	setupRouteTest(t)

	server := gin.New()
	server.POST("/events", handlerWithUserID(createEvent, 42))
	response := performRequest(server, http.MethodPost, "/events",
		`{"title":"Go Meetup","description":"Meetup","location":"Online","date_time":"2026-10-14T18:00:00Z"}`)
	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body)
	}

	var userID int64
	if err := db.DB.QueryRow("SELECT user_id FROM events").Scan(&userID); err != nil {
		t.Fatalf("could not read persisted event owner: %v", err)
	}
	if userID != 42 {
		t.Errorf("expected event owner ID 42, got %d", userID)
	}
}

func TestUpdateEventRejectsNonOwnerAndUpdatesOwnerEvent(t *testing.T) {
	setupRouteTest(t)
	eventID := insertTestEvent(t, 7)

	server := gin.New()
	server.PUT("/events/:id", handlerWithUserID(updateEvent, 8))
	response := performRequest(server, http.MethodPut, "/events/"+strconv.FormatInt(eventID, 10),
		`{"title":"Updated","description":"Updated description","location":"Remote","date_time":"2026-10-15T18:00:00Z"}`)
	if response.Code != http.StatusForbidden {
		t.Fatalf("expected non-owner status %d, got %d", http.StatusForbidden, response.Code)
	}

	server = gin.New()
	server.PUT("/events/:id", handlerWithUserID(updateEvent, 7))
	response = performRequest(server, http.MethodPut, "/events/"+strconv.FormatInt(eventID, 10),
		`{"title":"Updated","description":"Updated description","location":"Remote","date_time":"2026-10-15T18:00:00Z"}`)
	if response.Code != http.StatusOK {
		t.Fatalf("expected owner status %d, got %d: %s", http.StatusOK, response.Code, response.Body)
	}

	var title string
	if err := db.DB.QueryRow("SELECT title FROM events WHERE id = ?", eventID).Scan(&title); err != nil {
		t.Fatalf("could not read updated event: %v", err)
	}
	if title != "Updated" {
		t.Errorf("expected updated title, got %q", title)
	}
}

func TestDeleteEventDeletesOwnerEvent(t *testing.T) {
	setupRouteTest(t)
	eventID := insertTestEvent(t, 7)

	server := gin.New()
	server.DELETE("/events/:id", handlerWithUserID(deleteEvent, 7))
	response := performRequest(server, http.MethodDelete, "/events/"+strconv.FormatInt(eventID, 10), "")
	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body)
	}

	var count int
	if err := db.DB.QueryRow("SELECT COUNT(*) FROM events WHERE id = ?", eventID).Scan(&count); err != nil {
		t.Fatalf("could not verify event deletion: %v", err)
	}
	if count != 0 {
		t.Errorf("expected event to be deleted, found %d rows", count)
	}
}

func TestEventHandlersReturnNotFoundForMissingEvent(t *testing.T) {
	setupRouteTest(t)

	server := gin.New()
	server.GET("/events/:id", getEvent)
	response := performRequest(server, http.MethodGet, "/events/999", "")
	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, response.Code)
	}
}

func TestGetEventsReturnsInternalErrorWhenDatabaseIsClosed(t *testing.T) {
	setupRouteTest(t)
	if err := db.DB.Close(); err != nil {
		t.Fatalf("could not close test database: %v", err)
	}

	server := gin.New()
	server.GET("/events", getEvents)
	response := performRequest(server, http.MethodGet, "/events", "")
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, response.Code)
	}
}

func TestCreateEventRejectsInvalidJSONAndReportsDatabaseError(t *testing.T) {
	setupRouteTest(t)
	server := gin.New()
	server.POST("/events", handlerWithUserID(createEvent, 42))

	invalidResponse := performRequest(server, http.MethodPost, "/events", `{}`)
	if invalidResponse.Code != http.StatusBadRequest {
		t.Errorf("expected invalid payload status %d, got %d", http.StatusBadRequest, invalidResponse.Code)
	}

	if err := db.DB.Close(); err != nil {
		t.Fatalf("could not close test database: %v", err)
	}
	response := performRequest(server, http.MethodPost, "/events",
		`{"title":"Event","description":"Description","location":"Online","date_time":"2026-10-14T18:00:00Z"}`)
	if response.Code != http.StatusInternalServerError {
		t.Errorf("expected database error status %d, got %d", http.StatusInternalServerError, response.Code)
	}
}

func TestUpdateEventHandlesInvalidIDMissingEventAndInvalidPayload(t *testing.T) {
	setupRouteTest(t)
	server := gin.New()
	server.PUT("/events/:id", handlerWithUserID(updateEvent, 7))

	invalidID := performRequest(server, http.MethodPut, "/events/invalid", `{}`)
	if invalidID.Code != http.StatusBadRequest {
		t.Errorf("expected invalid ID status %d, got %d", http.StatusBadRequest, invalidID.Code)
	}

	missing := performRequest(server, http.MethodPut, "/events/999", `{}`)
	if missing.Code != http.StatusNotFound {
		t.Errorf("expected missing event status %d, got %d", http.StatusNotFound, missing.Code)
	}

	eventID := insertTestEvent(t, 7)
	invalidPayload := performRequest(server, http.MethodPut, "/events/"+strconv.FormatInt(eventID, 10), `{}`)
	if invalidPayload.Code != http.StatusBadRequest {
		t.Errorf("expected invalid payload status %d, got %d", http.StatusBadRequest, invalidPayload.Code)
	}

	if _, err := db.DB.Exec(`
		CREATE TRIGGER reject_event_update BEFORE UPDATE ON events
		BEGIN SELECT RAISE(ABORT, 'update rejected'); END;
	`); err != nil {
		t.Fatalf("could not create update trigger: %v", err)
	}
	updateError := performRequest(server, http.MethodPut, "/events/"+strconv.FormatInt(eventID, 10),
		`{"title":"Updated","description":"Description","location":"Online","date_time":"2026-10-15T18:00:00Z"}`)
	if updateError.Code != http.StatusInternalServerError {
		t.Errorf("expected update error status %d, got %d", http.StatusInternalServerError, updateError.Code)
	}
}

func TestDeleteEventHandlesInvalidIDMissingEventAndDatabaseError(t *testing.T) {
	setupRouteTest(t)
	server := gin.New()
	server.DELETE("/events/:id", handlerWithUserID(deleteEvent, 7))

	invalidID := performRequest(server, http.MethodDelete, "/events/invalid", "")
	if invalidID.Code != http.StatusBadRequest {
		t.Errorf("expected invalid ID status %d, got %d", http.StatusBadRequest, invalidID.Code)
	}
	missing := performRequest(server, http.MethodDelete, "/events/999", "")
	if missing.Code != http.StatusNotFound {
		t.Errorf("expected missing event status %d, got %d", http.StatusNotFound, missing.Code)
	}

	eventID := insertTestEvent(t, 7)
	nonOwnerServer := gin.New()
	nonOwnerServer.DELETE("/events/:id", handlerWithUserID(deleteEvent, 8))
	forbidden := performRequest(nonOwnerServer, http.MethodDelete, "/events/"+strconv.FormatInt(eventID, 10), "")
	if forbidden.Code != http.StatusForbidden {
		t.Errorf("expected non-owner status %d, got %d", http.StatusForbidden, forbidden.Code)
	}

	if _, err := db.DB.Exec(`
		CREATE TRIGGER reject_event_delete BEFORE DELETE ON events
		BEGIN SELECT RAISE(ABORT, 'delete rejected'); END;
	`); err != nil {
		t.Fatalf("could not create delete trigger: %v", err)
	}
	response := performRequest(server, http.MethodDelete, "/events/"+strconv.FormatInt(eventID, 10), "")
	if response.Code != http.StatusInternalServerError {
		t.Errorf("expected delete error status %d, got %d", http.StatusInternalServerError, response.Code)
	}
}
