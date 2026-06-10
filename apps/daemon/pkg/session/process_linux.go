//go:build linux

// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package session

import (
	"os/exec"
	"syscall"
)

func setNewProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func killProcessGroup(pid int, sig syscall.Signal) {
	_ = syscall.Kill(-pid, sig)
}

// SupportedOnPlatform reports whether sessions work on this platform.
// Linux session shells speak the POSIX wrapper used by Execute.
const SupportedOnPlatform = true
