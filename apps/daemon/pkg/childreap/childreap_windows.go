//go:build windows

// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

// Package childreap is a thin os/exec wrapper on Windows. The Linux build
// wraps go-reaper to recover the exit status of orphaned children when the
// daemon runs as PID 1; Windows has no PID-1 reparenting and no SIGCHLD, so
// no reaper is installed here. The helpers still honor the (exitCode, err)
// contract shared with Linux: a non-zero exit folds into the returned exit
// code with a nil error, and a non-nil error is returned only when cmd
// couldn't be started or no exit status could be recovered. Callers must
// check exitCode != 0 to detect command failure.
package childreap

import (
	"bytes"
	"errors"
	"os/exec"
)

func Start() {}

// Wait waits for a started cmd and returns its exit code. See the Linux
// implementation for the full (exitCode, err) contract.
func Wait(cmd *exec.Cmd) (int, error) {
	if cmd == nil || cmd.Process == nil {
		return -1, errors.New("childreap.Wait: cmd not started")
	}
	return foldExitError(cmd.Wait())
}

// Reap is equivalent to Wait on Windows: with no PID-1 reaper to race
// against, cmd.Wait is always the sole status consumer.
func Reap(cmd *exec.Cmd) (int, error) {
	return Wait(cmd)
}

// Run starts cmd and waits for it to finish. Like the Linux implementation,
// non-zero exits are reported via the exit code, not the error.
func Run(cmd *exec.Cmd) (int, error) {
	if cmd == nil {
		return -1, errors.New("childreap.Run: nil cmd")
	}
	if err := cmd.Start(); err != nil {
		return -1, err
	}
	return Wait(cmd)
}

// CombinedOutput runs cmd with stdout and stderr both captured into a
// single buffer. Returns (output, exitCode, err); err being nil does NOT
// mean the command succeeded — check exitCode for that.
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

// Output runs cmd and captures stdout into a buffer. Stderr is left as-is.
// Returns (stdout, exitCode, err); see Run for the (exitCode, err) contract.
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

// foldExitError mirrors the Linux contract: *exec.ExitError carries a real
// exit status, so it folds into (code, nil); any other error means the
// status could not be recovered.
func foldExitError(waitErr error) (int, error) {
	if waitErr == nil {
		return 0, nil
	}
	var ee *exec.ExitError
	if errors.As(waitErr, &ee) {
		return ee.ExitCode(), nil
	}
	return -1, waitErr
}
