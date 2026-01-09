// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package session

import "time"

type SessionService struct {
	configDir                string
	sessions                 map[string]*session
	terminationGracePeriod   time.Duration
	terminationCheckInterval time.Duration
}

func NewSessionService(configDir string, terminationGracePeriod, terminationCheckInterval time.Duration) *SessionService {
	return &SessionService{
		configDir:                configDir,
		sessions:                 make(map[string]*session),
		terminationGracePeriod:   terminationGracePeriod,
		terminationCheckInterval: terminationCheckInterval,
	}
}
