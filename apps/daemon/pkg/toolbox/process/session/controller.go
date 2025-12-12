// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package session

import (
	"time"

	"github.com/daytonaio/daemon/pkg/session"
)

type SessionController struct {
	configDir      string
	sessionService *session.SessionService
}

func NewSessionController(configDir, workDir string, terminationGracePeriodSeconds, terminationCheckIntervalMilliseconds int) *SessionController {
	if terminationGracePeriodSeconds <= 0 {
		terminationGracePeriodSeconds = 5 // default to 5 seconds
	}

	if terminationCheckIntervalMilliseconds <= 0 {
		terminationCheckIntervalMilliseconds = 100 // default to 100 milliseconds
	}

	terminationGracePeriod := time.Duration(terminationGracePeriodSeconds) * time.Second
	terminationCheckInterval := time.Duration(terminationCheckIntervalMilliseconds) * time.Millisecond

	service := session.NewSessionService(configDir, terminationGracePeriod, terminationCheckInterval)

	return &SessionController{
		configDir:      configDir,
		sessionService: service,
	}
}
