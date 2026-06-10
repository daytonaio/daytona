//go:build windows

// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package coderun

import (
	"os"
	"os/exec"
	"strconv"
)

// setNewProcessGroup is a no-op on Windows; killProcessGroupHard tears down
// the whole process tree via taskkill /T instead of POSIX process groups.
func setNewProcessGroup(cmd *exec.Cmd) {}

// killProcessGroupHard kills the process and all of its descendants.
// taskkill /T walks the parent-PID tree, the closest Windows equivalent of
// the process-group SIGKILL in process_linux.go.
func killProcessGroupHard(pid int) error {
	if err := exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(pid)).Run(); err == nil {
		return nil
	}
	// Fall back to killing the immediate process (e.g. taskkill unavailable).
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return p.Kill()
}
