// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package session

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
)

func (s *SessionService) TerminateCommand(sessionId, commandId string) error {
	session, ok := s.sessions.Get(sessionId)
	if !ok {
		return common_errors.NewNotFoundError(errors.New("session not found"))
	}

	command, ok := session.commands.Get(commandId)
	if !ok {
		return common_errors.NewNotFoundError(errors.New("command not found"))
	}

	if command.ExitCode != nil {
		return common_errors.NewGoneError(fmt.Errorf("command has already completed with exit code %d", *command.ExitCode))
	}

	return s.killCommand(session, command)
}

func (s *SessionService) readCommandPid(session *session, command *Command) (int, error) {
	pidFilePath := command.PidFilePath(session.Dir(s.configDir))

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		data, err := os.ReadFile(pidFilePath)
		if err == nil {
			pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
			if err != nil {
				return 0, common_errors.NewInternalServerError(fmt.Errorf("invalid pid file content: %w", err))
			}
			return pid, nil
		}
		if !os.IsNotExist(err) {
			return 0, common_errors.NewInternalServerError(fmt.Errorf("failed to read pid file: %w", err))
		}
		time.Sleep(50 * time.Millisecond)
	}

	return 0, common_errors.NewInternalServerError(errors.New("command pid file not available yet"))
}

func (s *SessionService) killCommand(session *session, command *Command) error {
	pid, err := s.readCommandPid(session, command)
	if err != nil {
		return err
	}

	return syscall.Kill(-pid, syscall.SIGKILL)
}
