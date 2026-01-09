// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package session

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
)

func (s *SessionService) getSessionCommands(sessionId string) ([]*Command, error) {
	session, ok := s.sessions[sessionId]
	if !ok {
		return nil, common_errors.NewNotFoundError(errors.New("session not found"))
	}

	commands := []*Command{}
	for _, command := range session.commands {
		cmd, err := s.GetSessionCommand(sessionId, command.Id)
		if err != nil {
			return nil, err
		}
		commands = append(commands, cmd)
	}

	return commands, nil
}

func (s *SessionService) GetSessionCommand(sessionId, cmdId string) (*Command, error) {
	session, ok := s.sessions[sessionId]
	if !ok {
		return nil, common_errors.NewNotFoundError(errors.New("session not found"))
	}

	command, ok := session.commands[cmdId]
	if !ok {
		return nil, common_errors.NewNotFoundError(errors.New("command not found"))
	}

	if command.ExitCode != nil {
		return command, nil
	}

	_, exitCodeFilePath := command.LogFilePath(session.Dir(s.configDir))
	exitCode, err := os.ReadFile(exitCodeFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return command, nil
		}
		return nil, fmt.Errorf("failed to read exit code file: %w", err)
	}

	exitCodeInt, err := strconv.Atoi(strings.TrimRight(string(exitCode), "\n"))
	if err != nil {
		return nil, fmt.Errorf("failed to convert exit code to int: %w", err)
	}

	command.ExitCode = &exitCodeInt

	return command, nil
}
