// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package session

import "github.com/daytonaio/daemon/pkg/session"

type CreateSessionRequest struct {
	SessionId string `json:"sessionId" validate:"required"`
} // @name CreateSessionRequest

type SessionExecuteRequest struct {
	Command  string `json:"command" validate:"required"`
	RunAsync bool   `json:"runAsync" validate:"optional"`
	Async    bool   `json:"async" validate:"optional"`
} // @name SessionExecuteRequest

type SessionSendInputRequest struct {
	Data string `json:"data" validate:"required"`
} // @name SessionSendInputRequest

type SessionExecuteResponse struct {
	CommandId *string `json:"cmdId" validate:"optional"`
	Output    *string `json:"output" validate:"optional"`
	Stdout    *string `json:"stdout" validate:"optional"`
	Stderr    *string `json:"stderr" validate:"optional"`
	ExitCode  *int    `json:"exitCode" validate:"optional"`
} // @name SessionExecuteResponse

type SessionCommandLogsResponse struct {
	Stdout string `json:"stdout" validate:"required"`
	Stderr string `json:"stderr" validate:"required"`
} // @name SessionCommandLogsResponse

type CommandDTO struct {
	Id       string `json:"id" validate:"required"`
	Command  string `json:"command" validate:"required"`
	ExitCode *int   `json:"exitCode,omitempty" validate:"optional"`
} // @name CommandDTO

type SessionDTO struct {
	SessionId string        `json:"sessionId" validate:"required"`
	Commands  []*CommandDTO `json:"commands" validate:"required"`
} // @name SessionDTO

func CommandToDTO(c *session.Command) *CommandDTO {
	return &CommandDTO{
		Id:       c.Id,
		Command:  c.Command,
		ExitCode: c.ExitCode,
	}
}

func SessionToDTO(s *session.Session) *SessionDTO {
	commands := make([]*CommandDTO, 0, len(s.Commands))
	for _, cmd := range s.Commands {
		commands = append(commands, CommandToDTO(cmd))
	}

	return &SessionDTO{
		SessionId: s.SessionId,
		Commands:  commands,
	}
}
