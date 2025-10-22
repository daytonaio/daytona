// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package session

import (
	"log/slog"
	"time"

	cmap "github.com/orcaman/concurrent-map/v2"
)

type SessionService struct {
	logger                   *slog.Logger
	configDir                string
	sessions                 cmap.ConcurrentMap[string, *session]
	terminationGracePeriod   time.Duration
	terminationCheckInterval time.Duration
}

func NewSessionService(logger *slog.Logger, configDir string, terminationGracePeriod, terminationCheckInterval time.Duration) *SessionService {
	return &SessionService{
		logger:                   logger.With(slog.String("component", "session_service")),
		configDir:                configDir,
		sessions:                 cmap.New[*session](),
		terminationGracePeriod:   terminationGracePeriod,
		terminationCheckInterval: terminationCheckInterval,
	}
}
