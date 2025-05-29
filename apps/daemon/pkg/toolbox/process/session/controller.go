// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package session

type SessionController struct {
	configDir  string
	projectDir string
}

func NewSessionController(configDir, projectDir string) *SessionController {
	return &SessionController{
		configDir:  configDir,
		projectDir: projectDir,
	}
}
