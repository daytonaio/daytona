// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package childreap

import (
	"bytes"
	"errors"
	"os/exec"
)

// Run starts cmd and waits for it to finish. The reaper-safe analog of
// cmd.Run().
//
// Unlike cmd.Run(), this returns (exitCode, err) instead of folding
// non-zero exits into err — that's intentional: cmd.Run()'s behavior of
// returning *exec.ExitError only when cmd.Wait() recovers a real
// ProcessState breaks under PID-1 reaping (cmd.Wait() loses the race
// and returns ECHILD-wrapped SyscallError, which callers using a
// type-switch to detect non-zero exits misclassify as "command
// failed"). Callers who want the same semantic should check
// `exitCode != 0` explicitly.
//
// Returns a non-nil error only when cmd couldn't be started or its exit
// status couldn't be recovered (see Wait for details).
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
// single buffer. The reaper-safe analog of cmd.CombinedOutput().
//
// Returns (output, exitCode, err); see Run for the (exitCode, err)
// contract. err being nil does NOT mean the command succeeded — check
// exitCode for that.
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

// Output runs cmd and captures stdout into a buffer. The reaper-safe
// analog of cmd.Output(). Stderr is left as-is (typically discarded by
// Go's exec when unset).
//
// Returns (stdout, exitCode, err); see Run for the (exitCode, err)
// contract.
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
