package middlewares

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alejandrolopezj/events-booking-rest-api/utils"
	"github.com/gin-gonic/gin"
)

func TestAuthenticateRejectsMissingOrInvalidToken(t *testing.T) {
	setTestGinMode(t)

	tests := []struct {
		name  string
		token string
	}{
		{name: "missing token"},
		{name: "invalid token", token: "not-a-valid-token"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router := gin.New()
			nextHandlerCalled := false
			router.GET("/", Authenticate, func(context *gin.Context) {
				nextHandlerCalled = true
				context.Status(http.StatusOK)
			})

			request := httptest.NewRequest(http.MethodGet, "/", nil)
			if test.token != "" {
				request.Header.Set("Authorization", test.token)
			}
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusUnauthorized {
				t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, recorder.Code)
			}
			if nextHandlerCalled {
				t.Fatal("next handler was called for an unauthorized request")
			}

			var response map[string]string
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Fatalf("could not decode response body: %v", err)
			}
			if response["error"] != errorMessage {
				t.Errorf("expected error %q, got %q", errorMessage, response["error"])
			}
		})
	}
}

func TestAuthenticateAcceptsValidTokenAndSetsClaims(t *testing.T) {
	setTestGinMode(t)

	const userID int64 = 42
	const email = "user@example.com"
	token, err := utils.GenerateJWT(userID, email)
	if err != nil {
		t.Fatalf("could not generate JWT: %v", err)
	}

	router := gin.New()
	nextHandlerCalled := false
	router.GET("/", Authenticate, func(context *gin.Context) {
		nextHandlerCalled = true
		value, exists := context.Get("claims")
		if !exists {
			t.Error("expected claims to be set in the Gin context")
			context.Status(http.StatusInternalServerError)
			return
		}

		claims, ok := value.(*utils.JWTClaims)
		if !ok {
			t.Errorf("expected claims of type *utils.JWTClaims, got %T", value)
			context.Status(http.StatusInternalServerError)
			return
		}
		if claims.UserID != userID {
			t.Errorf("expected user ID %d, got %d", userID, claims.UserID)
		}
		if claims.Email != email {
			t.Errorf("expected email %q, got %q", email, claims.Email)
		}
		context.Status(http.StatusOK)
	})

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Authorization", token)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if !nextHandlerCalled {
		t.Fatal("next handler was not called for an authorized request")
	}
}

func setTestGinMode(t *testing.T) {
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
}
