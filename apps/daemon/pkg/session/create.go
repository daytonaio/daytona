// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package session

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/daytonaio/daemon/pkg/common"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
)

func (s *SessionService) Create(sessionId string, isLegacy bool) error {
	ctx, cancel := context.WithCancel(context.Background())

	cmd := exec.CommandContext(ctx, common.GetShell())
	cmd.Env = os.Environ()

	if isLegacy {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			cancel()
			return fmt.Errorf("failed to obtain user home directory while checking SDK version: %w", err)
		}

		cmd.Dir = homeDir
	}

	if _, ok := s.sessions[sessionId]; ok {
		cancel()
		return common_errors.NewConflictError(errors.New("session already exists"))
	}

	stdinWriter, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return err
	}

	err = cmd.Start()
	if err != nil {
		cancel()
		return err
	}

	session := &session{
		id:          sessionId,
		cmd:         cmd,
		stdinWriter: stdinWriter,
		commands:    map[string]*Command{},
		ctx:         ctx,
		cancel:      cancel,
	}
	s.sessions[sessionId] = session

	err = os.MkdirAll(session.Dir(s.configDir), 0755)
	if err != nil {
		return err
	}

	return nil
}
