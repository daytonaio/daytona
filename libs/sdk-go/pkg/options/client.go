// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package options

import "time"

// CreateSandbox holds optional parameters for [daytona.Client.Create].
type CreateSandbox struct {
	Timeout      *time.Duration // Maximum time to wait for sandbox creation
	WaitForStart bool           // Whether to wait for the sandbox to reach started state
	LogChannel   chan string    // Channel for receiving build logs during image builds
}

// WithTimeout sets the maximum duration to wait for sandbox creation to complete.
//
// If the timeout is exceeded before the sandbox is ready, Create returns an error.
// The default timeout is 60 seconds.
//
// Example:
//
//	sandbox, err := client.Create(ctx, params,
//	    options.WithTimeout(5*time.Minute),
//	)
func WithTimeout(timeout time.Duration) func(*CreateSandbox) {
	return func(opts *CreateSandbox) {
		opts.Timeout = &timeout
	}
}

// WithWaitForStart controls whether [daytona.Client.Create] waits for the sandbox
// to reach the started state before returning.
//
// When true (the default), Create blocks until the sandbox is fully started and ready
// for use. When false, Create returns immediately after the sandbox is created,
// which may be in a pending or building state.
//
// Example:
//
//	// Return immediately without waiting for the sandbox to start
//	sandbox, err := client.Create(ctx, params,
//	    options.WithWaitForStart(false),
//	)
func WithWaitForStart(waitForStart bool) func(*CreateSandbox) {
	return func(opts *CreateSandbox) {
		opts.WaitForStart = waitForStart
	}
}

// WithLogChannel provides a channel for receiving build logs during sandbox creation.
//
// When creating a sandbox from a custom image that requires building, build logs
// are streamed to the provided channel. The channel is closed when streaming completes.
// If no build is required, no logs are sent and the channel remains unused.
//
// Example:
//
//	logChan := make(chan string)
//	go func() {
//	    for log := range logChan {
//	        fmt.Println(log)
//	    }
//	}()
//	sandbox, err := client.Create(ctx, params,
//	    options.WithLogChannel(logChan),
//	)
func WithLogChannel(logChannel chan string) func(*CreateSandbox) {
	return func(opts *CreateSandbox) {
		opts.LogChannel = logChannel
	}
}
