// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package session

import (
	"context"
	"io"
	"os/exec"
	"path/filepath"
)

// Stream prefixes for multiplexing stdout/stderr in logs
var (
	STDOUT_PREFIX = []byte{0x01, 0x01, 0x01}
	STDERR_PREFIX = []byte{0x02, 0x02, 0x02}
)

type session struct {
	id          string
	cmd         *exec.Cmd
	stdinWriter io.Writer
	commands    map[string]*Command
	ctx         context.Context
	cancel      context.CancelFunc
}

func (s *session) Dir(configDir string) string {
	return filepath.Join(configDir, "sessions", s.id)
}

type Command struct {
	Id       string `json:"id" validate:"required"`
	Command  string `json:"command" validate:"required"`
	ExitCode *int   `json:"exitCode,omitempty" validate:"optional"`
}

func (c *Command) LogFilePath(sessionDir string) (string, string) {
	return filepath.Join(sessionDir, c.Id, "output.log"), filepath.Join(sessionDir, c.Id, "exit_code")
}

func (c *Command) InputFilePath(sessionDir string) string {
	return filepath.Join(sessionDir, c.Id, "input.pipe")
}

type Session struct {
	SessionId string     `json:"sessionId" validate:"required"`
	Commands  []*Command `json:"commands" validate:"required"`
}

type SessionExecute struct {
	CommandId *string `json:"cmdId" validate:"optional"`
	Output    *string `json:"output" validate:"optional"`
	Stdout    *string `json:"stdout" validate:"optional"`
	Stderr    *string `json:"stderr" validate:"optional"`
	ExitCode  *int    `json:"exitCode" validate:"optional"`
}
