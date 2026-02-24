// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package session

import (
	"log/slog"
	"time"

	"github.com/daytonaio/daemon/pkg/session"
)

type SessionController struct {
	logger         *slog.Logger
	configDir      string
	sessionService *session.SessionService
}

func NewSessionController(logger *slog.Logger, configDir, workDir string, terminationGracePeriodSeconds, terminationCheckIntervalMilliseconds int) *SessionController {
	if terminationGracePeriodSeconds <= 0 {
		terminationGracePeriodSeconds = 5 // default to 5 seconds
	}

	if terminationCheckIntervalMilliseconds <= 0 {
		terminationCheckIntervalMilliseconds = 100 // default to 100 milliseconds
	}

	terminationGracePeriod := time.Duration(terminationGracePeriodSeconds) * time.Second
	terminationCheckInterval := time.Duration(terminationCheckIntervalMilliseconds) * time.Millisecond

	service := session.NewSessionService(logger, configDir, terminationGracePeriod, terminationCheckInterval)

	return &SessionController{
		logger:         logger.With(slog.String("component", "session_controller")),
		configDir:      configDir,
		sessionService: service,
	}
}
