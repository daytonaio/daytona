// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"bufio"
	"io"
	"os/exec"
)

type CreateSessionRequest struct {
	SessionId string  `json:"sessionId" validate:"required"`
	Alias     *string `json:"alias,omitempty" validate:"optional"`
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
	Alias     *string    `json:"alias,omitempty" validate:"optional"`
	Commands  []*Command `json:"commands" validate:"required"`
} // @name Session

type session struct {
	cmd         *exec.Cmd
	alias       *string
	outReader   *bufio.Scanner
	stdinWriter io.Writer
	commands    map[string]*Command
}

type Command struct {
	Id       string `json:"id" validate:"required"`
	Command  string `json:"command" validate:"required"`
	ExitCode *int   `json:"exitCode,omitempty" validate:"optional"`
} // @name Command
