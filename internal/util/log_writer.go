// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

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
