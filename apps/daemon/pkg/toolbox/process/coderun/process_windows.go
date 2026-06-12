//go:build windows

// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package coderun

import (
	"os/exec"

	"github.com/daytonaio/daemon/pkg/common"
)

// setNewProcessGroup is a no-op on Windows; killProcessGroupHard tears down
// the whole process tree via taskkill /T instead of POSIX process groups.
func setNewProcessGroup(cmd *exec.Cmd) {}

// killProcessGroupHard kills the process and all of its descendants, the
// closest Windows equivalent of the process-group SIGKILL in process_linux.go.
func killProcessGroupHard(pid int) error {
	return common.KillProcessTree(pid)
}
