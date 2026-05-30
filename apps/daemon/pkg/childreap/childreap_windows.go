//go:build windows

// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

// Package childreap is a no-op on Windows. The Linux build wraps go-reaper to
// recover the exit status of orphaned children when the daemon runs as PID 1.
// Windows containers do not use PID-1 reparenting and have no SIGCHLD, so all
// helpers here delegate directly to os/exec semantics.
package childreap

import "os/exec"

func Start() {}

func Wait(cmd *exec.Cmd) (int, error) {
	if cmd == nil || cmd.Process == nil {
		return -1, exec.ErrNotFound
	}
	err := cmd.Wait()
	if cmd.ProcessState != nil {
		return cmd.ProcessState.ExitCode(), err
	}
	return -1, err
}

func Reap(cmd *exec.Cmd) (int, error) {
	return Wait(cmd)
}

func Run(cmd *exec.Cmd) (int, error) {
	if cmd == nil {
		return -1, exec.ErrNotFound
	}
	err := cmd.Run()
	if cmd.ProcessState != nil {
		return cmd.ProcessState.ExitCode(), err
	}
	return -1, err
}

func CombinedOutput(cmd *exec.Cmd) ([]byte, int, error) {
	if cmd == nil {
		return nil, -1, exec.ErrNotFound
	}
	out, err := cmd.CombinedOutput()
	if cmd.ProcessState != nil {
		return out, cmd.ProcessState.ExitCode(), err
	}
	return out, -1, err
}

func Output(cmd *exec.Cmd) ([]byte, int, error) {
	if cmd == nil {
		return nil, -1, exec.ErrNotFound
	}
	out, err := cmd.Output()
	if cmd.ProcessState != nil {
		return out, cmd.ProcessState.ExitCode(), err
	}
	return out, -1, err
}
