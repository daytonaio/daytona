// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package session

import (
	"errors"
	"log/slog"
	"time"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
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

func (s *SessionService) GetSessionPID(sessionId string) (int, error) {
	session, ok := s.sessions.Get(sessionId)
	if !ok {
		return 0, common_errors.NewNotFoundError(errors.New("session not found"))
	}

	if session.cmd == nil || session.cmd.Process == nil {
		return 0, errors.New("session process not available")
	}

	return session.cmd.Process.Pid, nil
}
