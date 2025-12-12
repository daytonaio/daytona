// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package session

func (s *SessionService) List() ([]Session, error) {
	sessions := []Session{}

	for sessionId := range s.sessions {
		commands, err := s.getSessionCommands(sessionId)
		if err != nil {
			return nil, err
		}

		sessions = append(sessions, Session{
			SessionId: sessionId,
			Commands:  commands,
		})
	}

	return sessions, nil
}
