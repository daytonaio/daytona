/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
	"github.com/sirupsen/logrus"
)

var secretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(api[_-]?key|access[_-]?key|secret[_-]?key|secret|token|password|passwd|authorization)[:=]\s*([^\s,;]+)`),
	regexp.MustCompile(`(?i)(bearer\s+)[A-Za-z0-9._\-~+/=]+`),
	regexp.MustCompile(`(?i)(AKIA[0-9A-Z]{16})`),
	regexp.MustCompile(`(?i)(gh[pousr]_[A-Za-z0-9]{20,})`),
	regexp.MustCompile(`(?i)(sk_live_[A-Za-z0-9]{16,})`),
}

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
		attrs = append(attrs, slog.Any(k, redactValue(v)))
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

func redactValue(value any) any {
	switch typed := value.(type) {
	case string:
		return redactString(typed)
	case fmt.Stringer:
		return redactString(typed.String())
	case []byte:
		return redactString(string(typed))
	default:
		return value
	}
}

func redactString(input string) string {
	redacted := input
	for _, pattern := range secretPatterns {
		redacted = pattern.ReplaceAllString(redacted, "$1[REDACTED]")
	}
	return redacted
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
