// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package options

import "time"

// RunCode holds optional parameters for [daytona.CodeInterpreterService.RunCode].
type RunCode struct {
	ContextID string            // Interpreter context ID for persistent state
	Env       map[string]string // Environment variables for code execution
	Timeout   *time.Duration    // Execution timeout
}

// WithCustomContext sets the interpreter context ID for code execution.
//
// Using a context allows you to maintain state (variables, imports, etc.)
// across multiple code executions. Create a context with CreateContext first.
//
// Example:
//
//	ctx, _ := sandbox.CodeInterpreter.CreateContext(ctx, nil)
//	channels, err := sandbox.CodeInterpreter.RunCode(ctx, "x = 42",
//	    options.WithCustomContext(ctx["id"].(string)),
//	)
func WithCustomContext(contextID string) func(*RunCode) {
	return func(opts *RunCode) {
		opts.ContextID = contextID
	}
}

// WithEnv sets environment variables for code execution.
//
// These variables are available to the code during execution.
//
// Example:
//
//	channels, err := sandbox.CodeInterpreter.RunCode(ctx, "import os; print(os.environ['API_KEY'])",
//	    options.WithEnv(map[string]string{"API_KEY": "secret"}),
//	)
func WithEnv(env map[string]string) func(*RunCode) {
	return func(opts *RunCode) {
		opts.Env = env
	}
}

// WithInterpreterTimeout sets the execution timeout for code.
//
// If the code doesn't complete within the timeout, execution is terminated.
//
// Example:
//
//	channels, err := sandbox.CodeInterpreter.RunCode(ctx, "import time; time.sleep(60)",
//	    options.WithInterpreterTimeout(5*time.Second),
//	)
func WithInterpreterTimeout(timeout time.Duration) func(*RunCode) {
	return func(opts *RunCode) {
		opts.Timeout = &timeout
	}
}
