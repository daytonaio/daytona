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

func NewSessionService(configDir string, terminationGracePeriodSeconds, terminationCheckIntervalMilliseconds int) *SessionService {
	if terminationGracePeriodSeconds <= 0 {
		terminationGracePeriodSeconds = 5 // default to 5 seconds
	}

	if terminationCheckIntervalMilliseconds <= 0 {
		terminationCheckIntervalMilliseconds = 100 // default to 100 milliseconds
	}

	terminationGracePeriod := time.Duration(terminationGracePeriodSeconds) * time.Second
	terminationCheckInterval := time.Duration(terminationCheckIntervalMilliseconds) * time.Millisecond

	return &SessionService{
		configDir:                configDir,
		sessions:                 make(map[string]*session),
		terminationGracePeriod:   terminationGracePeriod,
		terminationCheckInterval: terminationCheckInterval,
	}
}
