// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package session

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
} // @name SessionExecuteResponse
