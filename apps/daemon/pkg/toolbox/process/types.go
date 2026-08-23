// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package process

const maxOutputBytes uint32 = 1<<31 - 1

type ExecuteRequest struct {
	Command string `json:"command" validate:"required"`
	// Timeout in seconds, defaults to 10 seconds
	Timeout *uint32 `json:"timeout,omitempty" validate:"optional"`
	// Maximum bytes retained from each output stream
	MaxOutputBytes *uint32 `json:"maxOutputBytes,omitempty" validate:"optional" minimum:"1" maximum:"2147483647"`
	// Current working directory
	Cwd *string `json:"cwd,omitempty" validate:"optional"`
	// Environment variables to set for the command
	Envs map[string]string `json:"envs,omitempty" validate:"optional"`
} //	@name	ExecuteRequest

// TODO: Set ExitCode as required once all sandboxes migrated to the new daemon
type ExecuteResponse struct {
	ExitCode        int    `json:"exitCode"`
	Result          string `json:"result" validate:"required"`
	Stdout          string `json:"stdout,omitempty" validate:"optional"`
	Stderr          string `json:"stderr,omitempty" validate:"optional"`
	TimedOut        bool   `json:"timedOut,omitempty" validate:"optional"`
	StdoutTruncated bool   `json:"stdoutTruncated,omitempty" validate:"optional"`
	StderrTruncated bool   `json:"stderrTruncated,omitempty" validate:"optional"`
} //	@name	ExecuteResponse
