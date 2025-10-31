package util

import (
	"log/slog"
	"strings"
)

// parseSlogLevel parses log level from environment variable
func ParseLogLevel(env string) slog.Level {
	switch strings.ToLower(env) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
