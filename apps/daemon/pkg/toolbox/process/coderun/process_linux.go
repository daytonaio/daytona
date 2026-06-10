//go:build linux

// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package coderun

import (
	"os/exec"
	"syscall"
)

// codeRunPlatformError reports whether code execution is supported on this
// platform. Always nil on Linux.
func codeRunPlatformError() error {
	return nil
}

func setNewProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func killProcessGroupHard(pid int) error {
	return syscall.Kill(-pid, syscall.SIGKILL)
}
