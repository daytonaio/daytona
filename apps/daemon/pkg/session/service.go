// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package session

import (
	"time"

	cmap "github.com/orcaman/concurrent-map/v2"
)

type SessionService struct {
	configDir                string
	sessions                 cmap.ConcurrentMap[string, *session]
	terminationGracePeriod   time.Duration
	terminationCheckInterval time.Duration
}

func NewSessionService(configDir string, terminationGracePeriod, terminationCheckInterval time.Duration) *SessionService {
	return &SessionService{
		configDir:                configDir,
		sessions:                 cmap.New[*session](),
		terminationGracePeriod:   terminationGracePeriod,
		terminationCheckInterval: terminationCheckInterval,
	}
}
