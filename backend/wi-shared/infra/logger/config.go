package logger

import (
	"log/slog"
	"strings"
)

// Options configures a service logger.
type Options struct {
	Service string
	Env     string
	Level   string
}

// ParseLevel maps config strings to slog levels. Unknown values default to info.
func ParseLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
