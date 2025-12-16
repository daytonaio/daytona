// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package session

import (
	"errors"
	"fmt"
	"os"
	"strings"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// SendInput sends data to the session's stdin for a specific running command
// This enables interactive command input for sessions
func (s *SessionService) SendInput(sessionId, commandId string, data string) error {
	session, ok := s.sessions[sessionId]
	if !ok {
		return common_errors.NewNotFoundError(errors.New("session not found"))
	}

	// Check if the session process is still active
	if session.cmd.ProcessState != nil && session.cmd.ProcessState.Exited() {
		return common_errors.NewGoneError(errors.New("session process has exited"))
	}

	// Verify the command exists
	command, ok := session.commands[commandId]
	if !ok {
		return common_errors.NewNotFoundError(errors.New("command not found"))
	}

	// Check if the command is still running (exit code not set means still running)
	if command.ExitCode != nil {
		return common_errors.NewGoneError(fmt.Errorf("command has already completed with exit code %d", *command.ExitCode))
	}

	inputFilePath := command.InputFilePath(session.Dir(s.configDir))
	f, err := os.OpenFile(inputFilePath, os.O_WRONLY, 0600)
	if err != nil {
		return common_errors.NewInternalServerError(fmt.Errorf("failed to open input pipe: %w", err))
	}
	defer f.Close()

	// Ensure newline for commands like `read`
	if !strings.HasSuffix(data, "\n") {
		data += "\n"
	}

	// Write to input pipe
	if _, err := f.Write([]byte(data)); err != nil {
		return common_errors.NewInternalServerError(fmt.Errorf("failed to write to input pipe: %w", err))
	}

	// Also echo input to log file for visibility (appears as stdout)
	logFilePath, _ := command.LogFilePath(session.Dir(s.configDir))
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		log.Debugf("failed to open log file to echo input: %v", err)
	} else {
		defer logFile.Close()
		// Write with STDOUT prefix to maintain log format consistency
		dataWithPrefix := append(STDOUT_PREFIX, []byte(data)...)
		logFile.Write(dataWithPrefix)
	}

	return nil
}
