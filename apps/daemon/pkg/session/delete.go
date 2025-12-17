// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package session

import (
	"context"
	"errors"
	"os"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/daytonaio/daemon/pkg/common"

	log "github.com/sirupsen/logrus"
)

func (s *SessionService) Delete(ctx context.Context, sessionId string) error {
	session, ok := s.sessions[sessionId]
	if !ok {
		return common_errors.NewNotFoundError(errors.New("session not found"))
	}

	// Terminate process group first with signals (SIGTERM -> SIGKILL)
	err := common.TerminateProcessTreeGracefully(ctx, session.cmd.Process, s.terminationGracePeriod, s.terminationCheckInterval)
	if err != nil {
		log.Errorf("Failed to terminate session %s: %v", session.id, err)
		// Continue with cleanup even if termination fails
	}

	// Cancel context after termination
	session.cancel()

	// Clean up session directory
	err = os.RemoveAll(session.Dir(s.configDir))
	if err != nil {
		return common_errors.NewBadRequestError(err)
	}

	delete(s.sessions, session.id)
	return nil
}
