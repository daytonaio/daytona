/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
	"github.com/sirupsen/logrus"
)

func NewLogger() *slog.Logger {
	log := slog.New(tint.NewHandler(os.Stdout, &tint.Options{
		NoColor:    !isatty.IsTerminal(os.Stdout.Fd()),
		TimeFormat: time.RFC3339,
		Level:      parseLogLevel(os.Getenv("LOG_LEVEL")),
	}))
	slog.SetDefault(log)

	// Redirect logrus (used by distribution registry) to slog
	logrus.SetOutput(io.Discard)
	logrus.AddHook(&slogHook{logger: log})

	return log
}

// slogHook redirects logrus logs to slog
type slogHook struct {
	logger *slog.Logger
}

func (h *slogHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *slogHook) Fire(entry *logrus.Entry) error {
	attrs := make([]slog.Attr, 0, len(entry.Data))
	for k, v := range entry.Data {
		attrs = append(attrs, slog.Any(k, v))
	}

	level := slog.LevelInfo
	switch entry.Level {
	case logrus.TraceLevel, logrus.DebugLevel:
		level = slog.LevelDebug
	case logrus.InfoLevel:
		level = slog.LevelInfo
	case logrus.WarnLevel:
		level = slog.LevelWarn
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		level = slog.LevelError
	}

	h.logger.LogAttrs(context.Background(), level, entry.Message, attrs...)
	return nil
}

// parseLogLevel converts a string log level to slog.Level
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
