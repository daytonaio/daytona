//go:build windows

// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package session

import (
	"os/exec"
	"syscall"
)

// setNewProcessGroup is a no-op on Windows. Process group semantics are a
// POSIX concept; Windows uses Job Objects (not wired up here). Per-process
// termination via the signalProcessTree fallback is sufficient.
func setNewProcessGroup(cmd *exec.Cmd) {}

// killProcessGroup is a no-op on Windows. Callers also invoke
// signalProcessTree, which is portable and handles termination via gopsutil.
func killProcessGroup(pid int, sig syscall.Signal) {}
