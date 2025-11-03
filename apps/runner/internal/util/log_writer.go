// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package util

import (
	"log/slog"
)

type DebugLogWriter struct{}

func (w *DebugLogWriter) Write(p []byte) (n int, err error) {
	slog.Debug(string(p))
	return len(p), nil
}

type InfoLogWriter struct{}

func (w *InfoLogWriter) Write(p []byte) (n int, err error) {
	slog.Info(string(p))
	return len(p), nil
}

type ErrorLogWriter struct{}

func (w *ErrorLogWriter) Write(p []byte) (n int, err error) {
	slog.Error(string(p))
	return len(p), nil
}
