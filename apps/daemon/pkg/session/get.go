// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package session

import (
	"errors"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
)

func (s *SessionService) Get(sessionId string) (*Session, error) {
	_, ok := s.sessions[sessionId]
	if !ok {
		return nil, common_errors.NewNotFoundError(errors.New("session not found"))
	}

	commands, err := s.getSessionCommands(sessionId)
	if err != nil {
		return nil, err
	}

	return &Session{
		SessionId: sessionId,
		Commands:  commands,
	}, nil
}
