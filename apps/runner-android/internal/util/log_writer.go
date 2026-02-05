// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package util

import (
	"os"

	log "github.com/sirupsen/logrus"
)

// DebugLogWriter writes to stdout only when log level is debug
type DebugLogWriter struct{}

func (w *DebugLogWriter) Write(p []byte) (n int, err error) {
	if log.GetLevel() >= log.DebugLevel {
		return os.Stdout.Write(p)
	}
	return len(p), nil
}
