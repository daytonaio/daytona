// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package options

import (
	"time"

	"github.com/daytonaio/daytona/libs/sdk-go/pkg/types"
)

// ExecuteCommand holds optional parameters for [daytona.ProcessService.ExecuteCommand].
type ExecuteCommand struct {
	Cwd     *string           // Working directory for command execution
	Env     map[string]string // Environment variables
	Timeout *time.Duration    // Command execution timeout. 0 means wait indefinitely. Defaults to 6 minutes.
}

// WithCwd sets the working directory for command execution.
//
// Example:
//
//	result, err := sandbox.Process.ExecuteCommand(ctx, "ls -la",
//	    options.WithCwd("/home/user/project"),
//	)
func WithCwd(cwd string) func(*ExecuteCommand) {
	return func(opts *ExecuteCommand) {
		opts.Cwd = &cwd
	}
}

// WithCommandEnv sets environment variables for the command.
//
// These variables are added to the command's environment in addition to
// the sandbox's default environment.
//
// Example:
//
//	result, err := sandbox.Process.ExecuteCommand(ctx, "echo $MY_VAR",
//	    options.WithCommandEnv(map[string]string{"MY_VAR": "hello"}),
//	)
func WithCommandEnv(env map[string]string) func(*ExecuteCommand) {
	return func(opts *ExecuteCommand) {
		opts.Env = env
	}
}

// WithExecuteTimeout sets the timeout for command execution.
//
// If the command doesn't complete within the timeout, it will be terminated.
//
// Example:
//
//	result, err := sandbox.Process.ExecuteCommand(ctx, "sleep 60",
//	    options.WithExecuteTimeout(5*time.Second),
//	)
func WithExecuteTimeout(timeout time.Duration) func(*ExecuteCommand) {
	return func(opts *ExecuteCommand) {
		opts.Timeout = &timeout
	}
}

// CodeRun holds optional parameters for [daytona.ProcessService.CodeRun].
type CodeRun struct {
	Params  *types.CodeRunParams // Code execution parameters
	Timeout *time.Duration       // Execution timeout. 0 means wait indefinitely. Defaults to 6 minutes.
}

// WithCodeRunParams sets the code execution parameters.
//
// Example:
//
//	result, err := sandbox.Process.CodeRun(ctx, code,
//	    options.WithCodeRunParams(types.CodeRunParams{Language: "python"}),
//	)
func WithCodeRunParams(params types.CodeRunParams) func(*CodeRun) {
	return func(opts *CodeRun) {
		opts.Params = &params
	}
}

// WithCodeRunTimeout sets the timeout for code execution.
//
// Example:
//
//	result, err := sandbox.Process.CodeRun(ctx, code,
//	    options.WithCodeRunTimeout(30*time.Second),
//	)
func WithCodeRunTimeout(timeout time.Duration) func(*CodeRun) {
	return func(opts *CodeRun) {
		opts.Timeout = &timeout
	}
}

// PtySession holds optional parameters for [daytona.ProcessService.CreatePtySession].
type PtySession struct {
	PtySize *types.PtySize    // Terminal dimensions (rows and columns)
	Env     map[string]string // Environment variables for the PTY session
}

// WithPtySize sets the PTY terminal dimensions.
//
// Example:
//
//	session, err := sandbox.Process.CreatePtySession(ctx, "my-session",
//	    options.WithPtySize(types.PtySize{Rows: 24, Cols: 80}),
//	)
func WithPtySize(size types.PtySize) func(*PtySession) {
	return func(opts *PtySession) {
		opts.PtySize = &size
	}
}

// WithPtyEnv sets environment variables for the PTY session.
//
// Example:
//
//	session, err := sandbox.Process.CreatePtySession(ctx, "my-session",
//	    options.WithPtyEnv(map[string]string{"TERM": "xterm-256color"}),
//	)
func WithPtyEnv(env map[string]string) func(*PtySession) {
	return func(opts *PtySession) {
		opts.Env = env
	}
}

// CreatePty holds optional parameters for [daytona.ProcessService.CreatePty].
type CreatePty struct {
	PtySize *types.PtySize    // Terminal dimensions (rows and columns)
	Env     map[string]string // Environment variables for the PTY session
}

// WithCreatePtySize sets the PTY terminal dimensions for CreatePty.
//
// Example:
//
//	handle, err := sandbox.Process.CreatePty(ctx, "my-pty",
//	    options.WithCreatePtySize(types.PtySize{Rows: 24, Cols: 80}),
//	)
func WithCreatePtySize(ptySize types.PtySize) func(*CreatePty) {
	return func(opts *CreatePty) {
		opts.PtySize = &ptySize
	}
}

// WithCreatePtyEnv sets environment variables for CreatePty.
//
// Example:
//
//	handle, err := sandbox.Process.CreatePty(ctx, "my-pty",
//	    options.WithCreatePtyEnv(map[string]string{"TERM": "xterm-256color"}),
//	)
func WithCreatePtyEnv(env map[string]string) func(*CreatePty) {
	return func(opts *CreatePty) {
		opts.Env = env
	}
}
