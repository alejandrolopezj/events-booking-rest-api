package main

import (
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/alejandrolopezj/events-booking-rest-api/utils"
	"github.com/gin-gonic/gin"
)

func TestRunInitializesDatabaseAndRegistersRoutes(t *testing.T) {
	setupMainTest(t)

	databaseInitialized := false
	serverStarted := false
	err := run(
		func() error {
			databaseInitialized = true
			return nil
		},
		func(server *gin.Engine) error {
			serverStarted = true
			if !databaseInitialized {
				t.Error("server started before database initialization")
			}

			registered := make(map[string]bool)
			for _, route := range server.Routes() {
				registered[route.Method+" "+route.Path] = true
			}
			for _, route := range []string{"GET /events", "POST /signup", "POST /events"} {
				if !registered[route] {
					t.Errorf("expected route %q to be registered", route)
				}
			}
			return nil
		},
	)

	if err != nil {
		t.Fatalf("run() returned an unexpected error: %v", err)
	}
	if !databaseInitialized {
		t.Error("database initialization was not called")
	}
	if !serverStarted {
		t.Error("server callback was not called")
	}
}

func TestRunReturnsDatabaseInitializationErrorWithoutStartingServer(t *testing.T) {
	setupMainTest(t)

	wantErr := errors.New("database unavailable")
	serverStarted := false
	err := run(
		func() error {
			return wantErr
		},
		func(*gin.Engine) error {
			serverStarted = true
			return nil
		},
	)

	if !errors.Is(err, wantErr) {
		t.Errorf("expected database initialization error %v, got %v", wantErr, err)
	}
	if serverStarted {
		t.Error("server callback was called after database initialization failed")
	}
}

func TestRunReturnsServerError(t *testing.T) {
	setupMainTest(t)

	wantErr := errors.New("server failed")
	err := run(
		func() error { return nil },
		func(*gin.Engine) error { return wantErr },
	)
	if !errors.Is(err, wantErr) {
		t.Errorf("expected server error %v, got %v", wantErr, err)
	}
}

func TestRunMainReturnsZeroWhenServerStopsWithoutError(t *testing.T) {
	setupMainTest(t)

	exitCode := runMain(
		func() error { return nil },
		func(*gin.Engine) error { return nil },
	)
	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}
}

func TestRunMainReturnsOneWhenDatabaseInitializationFails(t *testing.T) {
	setupMainTest(t)

	exitCode := runMain(
		func() error { return errors.New("database unavailable") },
		func(*gin.Engine) error {
			t.Error("server callback was called after database initialization failed")
			return nil
		},
	)
	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}
}

func TestRunMainReturnsOneWhenServerFails(t *testing.T) {
	setupMainTest(t)

	exitCode := runMain(
		func() error { return nil },
		func(*gin.Engine) error { return errors.New("server failed") },
	)
	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}
}

func setupMainTest(t *testing.T) {
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
