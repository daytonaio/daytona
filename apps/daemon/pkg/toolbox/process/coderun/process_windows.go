//go:build windows

// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package coderun

import (
	"os"
	"os/exec"
)

// setNewProcessGroup is a no-op on Windows; Job Objects are not wired up here.
func setNewProcessGroup(cmd *exec.Cmd) {}

// killProcessGroupHard kills only the immediate process on Windows.
// Process group / Job Object semantics differ from POSIX; child processes
// must be torn down individually or via a Job Object (not used here).
func killProcessGroupHard(pid int) error {
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return p.Kill()
}
