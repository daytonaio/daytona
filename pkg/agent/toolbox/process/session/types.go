// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"io"
	"os/exec"
	"path/filepath"
)

type CreateSessionRequest struct {
	SessionId string `json:"sessionId" validate:"required"`
} // @name CreateSessionRequest

type SessionExecuteRequest struct {
	Command string `json:"command" validate:"required"`
	Async   bool   `json:"async" validate:"optional"`
} // @name SessionExecuteRequest

type SessionExecuteResponse struct {
	CommandId *string `json:"cmdId" validate:"optional"`
	Output    *string `json:"output" validate:"optional"`
	ExitCode  *int    `json:"exitCode" validate:"optional"`
} // @name SessionExecuteResponse

type Session struct {
	SessionId string     `json:"sessionId" validate:"required"`
	Commands  []*Command `json:"commands" validate:"required"`
} // @name Session

type session struct {
	id          string
	cmd         *exec.Cmd
	stdinWriter io.Writer
	commands    map[string]*Command
}

func (s *session) Dir(configDir string) string {
	return filepath.Join(configDir, "sessions", s.id)
}

type Command struct {
	Id       string `json:"id" validate:"required"`
	Command  string `json:"command" validate:"required"`
	ExitCode *int   `json:"exitCode,omitempty" validate:"optional"`
} // @name Command

func (c *Command) LogFilePath(sessionDir string) string {
	return filepath.Join(sessionDir, c.Id, "output.log")
}
