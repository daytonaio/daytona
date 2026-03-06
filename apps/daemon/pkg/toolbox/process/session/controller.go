// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package session

import (
	"log/slog"

	"github.com/daytonaio/daemon/pkg/session"
)

type SessionController struct {
	logger         *slog.Logger
	configDir      string
	sessionService *session.SessionService
}

func NewSessionController(logger *slog.Logger, configDir, workDir string, sessionService *session.SessionService) *SessionController {
	return &SessionController{
		logger:         logger.With(slog.String("component", "session_controller")),
		configDir:      configDir,
		sessionService: sessionService,
	}
}
