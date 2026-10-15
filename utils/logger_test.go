package utils

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
)

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input string
		want  slog.Level
	}{
		{input: "DEBUG", want: slog.LevelDebug},
		{input: "info", want: slog.LevelInfo},
		{input: " WARN ", want: slog.LevelWarn},
		{input: "WARNING", want: slog.LevelWarn},
		{input: "ERROR", want: slog.LevelError},
		{input: "", want: slog.LevelDebug},
		{input: "invalid", want: slog.LevelDebug},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			if got := parseLogLevel(test.input); got != test.want {
				t.Errorf("parseLogLevel(%q) = %v, want %v", test.input, got, test.want)
			}
		})
	}
}

func TestInitLoggerUsesConfiguredLevel(t *testing.T) {
	t.Setenv("LEVEL_LOG", "ERROR")

	previousLogger := Logger
	t.Cleanup(func() {
		Logger = previousLogger
	})

	InitLogger()
	if Logger == nil {
		t.Fatal("InitLogger() left Logger nil")
	}
	if Logger.Enabled(context.Background(), slog.LevelInfo) {
		t.Error("INFO should be disabled when LEVEL_LOG is ERROR")
	}
	if !Logger.Enabled(context.Background(), slog.LevelError) {
		t.Error("ERROR should be enabled when LEVEL_LOG is ERROR")
	}
}

func TestLogAndReturnError(t *testing.T) {
	var output bytes.Buffer
	previousLogger := Logger
	Logger = slog.New(slog.NewTextHandler(&output, nil))
	t.Cleanup(func() {
		Logger = previousLogger
	})

	technicalErr := errors.New("database unavailable")
	got := LogAndReturnError(technicalErr, "failed to fetch records", "Could not load records")
	if got == nil || got.Error() != "Could not load records" {
		t.Fatalf("expected user-facing error, got %v", got)
	}
	if !strings.Contains(output.String(), "failed to fetch records") {
		t.Errorf("log did not include the technical message: %q", output.String())
	}
	if !strings.Contains(output.String(), technicalErr.Error()) {
		t.Errorf("log did not include the technical error: %q", output.String())
	}
}

func TestLogAndReturnErrorWithNil(t *testing.T) {
	if err := LogAndReturnError(nil, "technical message", "user message"); err != nil {
		t.Errorf("expected nil for a nil input error, got %v", err)
	}
}
