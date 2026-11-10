package logger_test

import (
	"log/slog"
	"testing"

	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

func TestParseLevel(t *testing.T) {
	if logger.ParseLevel("debug") != slog.LevelDebug {
		t.Fatal("debug")
	}
	if logger.ParseLevel("warn") != slog.LevelWarn {
		t.Fatal("warn")
	}
	if logger.ParseLevel("warning") != slog.LevelWarn {
		t.Fatal("warning")
	}
	if logger.ParseLevel("error") != slog.LevelError {
		t.Fatal("error")
	}
	if logger.ParseLevel("unknown") != slog.LevelInfo {
		t.Fatal("default info")
	}
}
