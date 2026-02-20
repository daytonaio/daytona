// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package execute

type ExecuteRequest struct {
	Command string `json:"command" validate:"required"`
	// Timeout in seconds, defaults to 360 seconds (6 minutes)
	Timeout *uint32 `json:"timeout,omitempty" validate:"optional"`
	// Current working directory
	Cwd *string `json:"cwd,omitempty" validate:"optional"`
} // @name ExecuteRequest

type ExecuteResponse struct {
	ExitCode int    `json:"exitCode" validate:"required"`
	Result   string `json:"result" validate:"required"`
} // @name ExecuteResponse
