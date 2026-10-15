package routes

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSignupCreatesUserAndLoginReturnsToken(t *testing.T) {
	setupRouteTest(t)

	server := gin.New()
	server.POST("/signup", signup)
	server.POST("/login", login)

	signupResponse := performRequest(server, http.MethodPost, "/signup",
		`{"email":"user@example.com","password":"secure-password"}`)
	if signupResponse.Code != http.StatusCreated {
		t.Fatalf("expected signup status %d, got %d: %s", http.StatusCreated, signupResponse.Code, signupResponse.Body)
	}

	loginResponse := performRequest(server, http.MethodPost, "/login",
		`{"email":"user@example.com","password":"secure-password"}`)
	if loginResponse.Code != http.StatusOK {
		t.Fatalf("expected login status %d, got %d: %s", http.StatusOK, loginResponse.Code, loginResponse.Body)
	}

	var result map[string]string
	if err := json.Unmarshal(loginResponse.Body.Bytes(), &result); err != nil {
		t.Fatalf("could not decode login response: %v", err)
	}
	if result["token"] == "" {
		t.Fatal("expected login response to include a token")
	}
}

func TestSignupRejectsInvalidJSONAndLoginRejectsInvalidCredentials(t *testing.T) {
	setupRouteTest(t)

	server := gin.New()
	server.POST("/signup", signup)
	server.POST("/login", login)

	signupResponse := performRequest(server, http.MethodPost, "/signup", `{`)
	if signupResponse.Code != http.StatusBadRequest {
		t.Errorf("expected invalid signup status %d, got %d", http.StatusBadRequest, signupResponse.Code)
	}

	loginResponse := performRequest(server, http.MethodPost, "/login",
		`{"email":"missing@example.com","password":"wrong-password"}`)
	if loginResponse.Code != http.StatusUnauthorized {
		t.Errorf("expected invalid login status %d, got %d", http.StatusUnauthorized, loginResponse.Code)
	}

	invalidLoginJSON := performRequest(server, http.MethodPost, "/login", `{`)
	if invalidLoginJSON.Code != http.StatusBadRequest {
		t.Errorf("expected invalid login JSON status %d, got %d", http.StatusBadRequest, invalidLoginJSON.Code)
	}
}

func TestSignupReturnsErrorForDuplicateEmail(t *testing.T) {
	setupRouteTest(t)

	server := gin.New()
	server.POST("/signup", signup)
	payload := `{"email":"duplicate@example.com","password":"secure-password"}`

	firstResponse := performRequest(server, http.MethodPost, "/signup", payload)
	if firstResponse.Code != http.StatusCreated {
		t.Fatalf("expected first signup status %d, got %d", http.StatusCreated, firstResponse.Code)
	}
	duplicateResponse := performRequest(server, http.MethodPost, "/signup", payload)
	if duplicateResponse.Code != http.StatusInternalServerError {
		t.Fatalf("expected duplicate signup status %d, got %d", http.StatusInternalServerError, duplicateResponse.Code)
	}
}
