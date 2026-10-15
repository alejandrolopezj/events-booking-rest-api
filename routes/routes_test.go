package routes

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegisterRoutesRegistersEndpoints(t *testing.T) {
	setupRouteTest(t)

	server := gin.New()
	RegisterRoutes(server)

	registered := make(map[string]bool)
	for _, route := range server.Routes() {
		registered[route.Method+" "+route.Path] = true
	}

	expected := []string{
		"GET /events",
		"GET /events/:id",
		"POST /events",
		"PUT /events/:id",
		"DELETE /events/:id",
		"POST /events/:id/registration",
		"DELETE /events/:id/registration",
		"POST /signup",
		"POST /login",
	}
	for _, route := range expected {
		if !registered[route] {
			t.Errorf("expected route %q to be registered", route)
		}
	}
}

func TestRegisterRoutesProtectsEventCreation(t *testing.T) {
	setupRouteTest(t)

	server := gin.New()
	RegisterRoutes(server)

	response := performRequest(server, http.MethodPost, "/events", `{"title":"Event"}`)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}
