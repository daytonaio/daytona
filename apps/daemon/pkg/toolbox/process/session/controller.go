// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package session

import (
	"github.com/daytonaio/daemon/pkg/session"
)

type SessionController struct {
	configDir      string
	sessionService *session.SessionService
}

func NewSessionController(configDir, workDir string, sessionService *session.SessionService) *SessionController {
	return &SessionController{
		configDir:      configDir,
		sessionService: sessionService,
	}
}
