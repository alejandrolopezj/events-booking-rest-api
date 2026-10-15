package utils

import (
	"errors"
	"log/slog"
	"os"
	"strings"
)

// Logger is the application-wide structured logger used across the project.
var Logger *slog.Logger

// InitLogger initializes the global logger using the LEVEL_LOG environment
// variable. When unset or invalid, it defaults to DEBUG.
func InitLogger() {
	level := parseLogLevel(os.Getenv("LEVEL_LOG"))
	Logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
}

// LogAndReturnError writes the technical error to the application log and
// returns a user-friendly message for the client.
func LogAndReturnError(err error, technicalMessage, userMessage string) error {
	if err == nil {
		return nil
	}
	Logger.Error(technicalMessage, "error", err)
	return errors.New(userMessage)
}

// parseLogLevel converts a log level name into a slog.Level. Unrecognized or
// empty values default to DEBUG.
func parseLogLevel(value string) slog.Level {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN", "WARNING":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelDebug
	}
}
