// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package util

import (
	log "github.com/sirupsen/logrus"
)

type DebugLogWriter struct{}

func (w *DebugLogWriter) Write(p []byte) (n int, err error) {
	log.Debug(string(p))
	return len(p), nil
}

type InfoLogWriter struct{}

func (w *InfoLogWriter) Write(p []byte) (n int, err error) {
	log.Info(string(p))
	return len(p), nil
}

type TraceLogWriter struct{}

func (w *TraceLogWriter) Write(p []byte) (n int, err error) {
	log.Trace(string(p))
	return len(p), nil
}

type ErrorLogWriter struct{}

func (w *ErrorLogWriter) Write(p []byte) (n int, err error) {
	log.Error(string(p))
	return len(p), nil
}
