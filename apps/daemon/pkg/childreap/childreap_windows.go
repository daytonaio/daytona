// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

//go:build windows

package childreap

import (
	"bytes"
	"errors"
	"os/exec"
)

// Start is a no-op on Windows since Windows does not have PID 1 zombie reaping.
func Start() {}

// Wait waits for cmd to finish, drains I/O, and returns the exit code.
func Wait(cmd *exec.Cmd) (int, error) {
	if cmd == nil || cmd.Process == nil {
		return -1, errors.New("childreap.Wait: cmd not started")
	}
	err := cmd.Wait()
	if err == nil {
		return 0, nil
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return ee.ExitCode(), nil
	}
	return -1, err
}

// Reap waits for cmd to exit and returns its exit code.
func Reap(cmd *exec.Cmd) (int, error) {
	return Wait(cmd)
}

// Run starts cmd and waits for it to finish.
func Run(cmd *exec.Cmd) (int, error) {
	if cmd == nil {
		return -1, errors.New("childreap.Run: nil cmd")
	}
	if err := cmd.Start(); err != nil {
		return -1, err
	}
	return Wait(cmd)
}

// CombinedOutput runs cmd with stdout and stderr captured into a buffer.
func CombinedOutput(cmd *exec.Cmd) ([]byte, int, error) {
	if cmd == nil {
		return nil, -1, errors.New("childreap.CombinedOutput: nil cmd")
	}
	if cmd.Stdout != nil {
		return nil, -1, errors.New("childreap.CombinedOutput: cmd.Stdout already set")
	}
	if cmd.Stderr != nil {
		return nil, -1, errors.New("childreap.CombinedOutput: cmd.Stderr already set")
	}
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	code, err := Run(cmd)
	return buf.Bytes(), code, err
}

// Output runs cmd and captures stdout into a buffer.
func Output(cmd *exec.Cmd) ([]byte, int, error) {
	if cmd == nil {
		return nil, -1, errors.New("childreap.Output: nil cmd")
	}
	if cmd.Stdout != nil {
		return nil, -1, errors.New("childreap.Output: cmd.Stdout already set")
	}
	var buf bytes.Buffer
	cmd.Stdout = &buf
	code, err := Run(cmd)
	return buf.Bytes(), code, err
}
