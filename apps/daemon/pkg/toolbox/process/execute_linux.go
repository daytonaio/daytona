//go:build linux

// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package process

import (
	"os/exec"
	"strings"
	"syscall"

	"github.com/daytonaio/daemon/pkg/common"
)

// buildExecCmd pipes the command to the shell via stdin to avoid OS ARG_MAX
// limits on large commands, and starts it in its own process group so a
// timeout can kill the whole tree.
func buildExecCmd(command string) *exec.Cmd {
	cmd := exec.Command(common.GetShell())
	cmd.Stdin = strings.NewReader(command)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return cmd
}

// killExecProcessGroup kills the entire process group so child processes
// are also terminated.
func killExecProcessGroup(cmd *exec.Cmd) error {
	return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
}
