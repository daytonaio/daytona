// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package session

import (
	"context"
	"errors"
	"os"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/daytonaio/daemon/pkg/common"
)

func (s *SessionService) Delete(ctx context.Context, sessionId string) error {
	session, ok := s.sessions.Get(sessionId)
	if !ok {
		return common_errors.NewNotFoundError(errors.New("session not found"))
	}

	// Terminate process group first with signals (SIGTERM -> SIGKILL).
	// Use context.Background() so a disconnected HTTP client does not cancel
	// the grace period and force an immediate SIGKILL.
	if session.cmd != nil {
		err := common.TerminateProcessTreeGracefully(context.Background(), s.logger, session.cmd.Process, s.terminationGracePeriod, s.terminationCheckInterval)
		if err != nil {
			s.logger.ErrorContext(ctx, "Failed to terminate session", "sessionId", session.id, "error", err)
			// Continue with cleanup even if termination fails
		}
	}

	// Cancel context after termination
	session.cancel()

	// Clean up session directory
	err := os.RemoveAll(session.Dir(s.configDir))
	if err != nil {
		return common_errors.NewBadRequestError(err)
	}

	s.sessions.Remove(session.id)
	return nil
}
