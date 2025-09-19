// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package session

type SessionController struct {
	configDir string
}

func NewSessionController(configDir, workDir string) *SessionController {
	return &SessionController{
		configDir: configDir,
	}
}
