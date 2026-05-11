// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package logger

import (
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
)

func NewLogger() *slog.Logger {
	log := slog.New(tint.NewHandler(os.Stdout, &tint.Options{
		NoColor:    !isatty.IsTerminal(os.Stdout.Fd()),
		TimeFormat: time.RFC3339,
		Level:      parseLogLevel(os.Getenv("SESSION_DAEMON_LOG_LEVEL")),
	}))
	slog.SetDefault(log)
	return log
}

func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
