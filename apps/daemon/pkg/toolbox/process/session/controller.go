// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package session

import (
	"log/slog"
	"sync"

	"github.com/daytonaio/daemon/pkg/session"
	"github.com/daytonaio/daemon/pkg/toolbox/process"
)

type trackerRegistration struct {
	pid   int
	token uint64
}

type SessionController struct {
	logger         *slog.Logger
	configDir      string
	sessionService *session.SessionService
	tracker        *process.ProcessTracker
	trackerMu      sync.Mutex
	trackerEntries map[string]trackerRegistration
}

func NewSessionController(logger *slog.Logger, configDir string, sessionService *session.SessionService, tracker *process.ProcessTracker) *SessionController {
	return &SessionController{
		logger:         logger.With(slog.String("component", "session_controller")),
		configDir:      configDir,
		sessionService: sessionService,
		tracker:        tracker,
		trackerEntries: make(map[string]trackerRegistration),
	}
}

func (s *SessionController) storeTrackerRegistration(sessionID string, pid int, token uint64) {
	s.trackerMu.Lock()
	defer s.trackerMu.Unlock()
	s.trackerEntries[sessionID] = trackerRegistration{pid: pid, token: token}
}

func (s *SessionController) deleteTrackerRegistration(sessionID string) {
	s.trackerMu.Lock()
	defer s.trackerMu.Unlock()
	delete(s.trackerEntries, sessionID)
}

func (s *SessionController) popTrackerRegistration(sessionID string) (trackerRegistration, bool) {
	s.trackerMu.Lock()
	defer s.trackerMu.Unlock()
	entry, ok := s.trackerEntries[sessionID]
	if ok {
		delete(s.trackerEntries, sessionID)
	}
	return entry, ok
}
