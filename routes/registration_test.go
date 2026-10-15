package routes

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/alejandrolopezj/events-booking-rest-api/db"
	"github.com/gin-gonic/gin"
)

func TestRegisterAndUnregisterForEvent(t *testing.T) {
	setupRouteTest(t)
	eventID := insertTestEvent(t, 42)

	server := gin.New()
	server.POST("/events/:id/registration", handlerWithUserID(registerForEvent, 42))
	server.DELETE("/events/:id/registration", handlerWithUserID(unregisterFromEvent, 42))
	path := "/events/" + strconv.FormatInt(eventID, 10) + "/registration"

	registerResponse := performRequest(server, http.MethodPost, path, "")
	if registerResponse.Code != http.StatusCreated {
		t.Fatalf("expected registration status %d, got %d: %s", http.StatusCreated, registerResponse.Code, registerResponse.Body)
	}

	var count int
	if err := db.DB.QueryRow(
		"SELECT COUNT(*) FROM registrations WHERE event_id = ? AND user_id = ?",
		eventID, 42,
	).Scan(&count); err != nil {
		t.Fatalf("could not verify registration: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected one registration, got %d", count)
	}

	unregisterResponse := performRequest(server, http.MethodDelete, path, "")
	if unregisterResponse.Code != http.StatusOK {
		t.Fatalf("expected unregister status %d, got %d: %s", http.StatusOK, unregisterResponse.Code, unregisterResponse.Body)
	}
}

func TestRegistrationHandlersRejectInvalidIDAndNonOwner(t *testing.T) {
	setupRouteTest(t)
	eventID := insertTestEvent(t, 7)

	server := gin.New()
	server.POST("/events/:id/registration", handlerWithUserID(registerForEvent, 42))
	server.DELETE("/events/:id/registration", handlerWithUserID(unregisterFromEvent, 42))

	invalidResponse := performRequest(server, http.MethodPost, "/events/not-an-id/registration", "")
	if invalidResponse.Code != http.StatusBadRequest {
		t.Errorf("expected invalid ID status %d, got %d", http.StatusBadRequest, invalidResponse.Code)
	}

	path := "/events/" + strconv.FormatInt(eventID, 10) + "/registration"
	registerResponse := performRequest(server, http.MethodPost, path, "")
	if registerResponse.Code != http.StatusBadRequest {
		t.Errorf("expected non-owner registration status %d, got %d", http.StatusBadRequest, registerResponse.Code)
	}

	unregisterResponse := performRequest(server, http.MethodDelete, path, "")
	if unregisterResponse.Code != http.StatusBadRequest {
		t.Errorf("expected non-owner unregistration status %d, got %d", http.StatusBadRequest, unregisterResponse.Code)
	}
}

func TestRegistrationHandlersReturnNotFoundForMissingEvent(t *testing.T) {
	setupRouteTest(t)

	server := gin.New()
	server.POST("/events/:id/registration", handlerWithUserID(registerForEvent, 42))
	server.DELETE("/events/:id/registration", handlerWithUserID(unregisterFromEvent, 42))

	for _, method := range []string{http.MethodPost, http.MethodDelete} {
		response := performRequest(server, method, "/events/999/registration", "")
		if response.Code != http.StatusNotFound {
			t.Errorf("%s: expected missing event status %d, got %d", method, http.StatusNotFound, response.Code)
		}
	}
}

func TestRegistrationHandlersReportPersistenceErrors(t *testing.T) {
	setupRouteTest(t)
	eventID := insertTestEvent(t, 42)
	if _, err := db.DB.Exec(`
		CREATE TRIGGER reject_registration_insert BEFORE INSERT ON registrations
		BEGIN SELECT RAISE(ABORT, 'registration rejected'); END;
	`); err != nil {
		t.Fatalf("could not create registration trigger: %v", err)
	}

	server := gin.New()
	server.POST("/events/:id/registration", handlerWithUserID(registerForEvent, 42))
	server.DELETE("/events/:id/registration", handlerWithUserID(unregisterFromEvent, 42))
	path := "/events/" + strconv.FormatInt(eventID, 10) + "/registration"

	registerResponse := performRequest(server, http.MethodPost, path, "")
	if registerResponse.Code != http.StatusInternalServerError {
		t.Errorf("expected registration error status %d, got %d", http.StatusInternalServerError, registerResponse.Code)
	}

	if _, err := db.DB.Exec("DROP TRIGGER reject_registration_insert"); err != nil {
		t.Fatalf("could not drop registration trigger: %v", err)
	}
	if _, err := db.DB.Exec("INSERT INTO registrations (event_id, user_id) VALUES (?, ?)", eventID, 42); err != nil {
		t.Fatalf("could not insert registration fixture: %v", err)
	}
	if _, err := db.DB.Exec(`
		CREATE TRIGGER reject_registration_delete BEFORE DELETE ON registrations
		BEGIN SELECT RAISE(ABORT, 'unregistration rejected'); END;
	`); err != nil {
		t.Fatalf("could not create unregistration trigger: %v", err)
	}
	unregisterResponse := performRequest(server, http.MethodDelete, path, "")
	if unregisterResponse.Code != http.StatusInternalServerError {
		t.Errorf("expected unregistration error status %d, got %d", http.StatusInternalServerError, unregisterResponse.Code)
	}
}
