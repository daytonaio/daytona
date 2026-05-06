// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package process

type ExecuteRequest struct {
	Command string `json:"command" validate:"required"`
	// Timeout in seconds, defaults to 10 seconds
	Timeout *uint32 `json:"timeout,omitempty" validate:"optional"`
	// Current working directory
	Cwd *string `json:"cwd,omitempty" validate:"optional"`
	// Environment variables to set for the command
	Envs map[string]string `json:"envs,omitempty" validate:"optional"`
	// If true, run the command fully detached: it is launched via `setsid -f`
	// in a new session with stdin/stdout/stderr redirected to /dev/null, so
	// the request returns as soon as the launcher forks. Use this for
	// long-running daemons (e.g. starting a tmux server, a background worker)
	// where the caller does not need the command's output and must not be
	// blocked by it. ExitCode reflects whether the launcher started cleanly,
	// not the eventual exit status of the detached process; Result is empty.
	RunDetached bool `json:"runDetached,omitempty" validate:"optional"`
} //	@name	ExecuteRequest

// TODO: Set ExitCode as required once all sandboxes migrated to the new daemon
type ExecuteResponse struct {
	ExitCode int    `json:"exitCode"`
	Result   string `json:"result" validate:"required"`
} //	@name	ExecuteResponse
