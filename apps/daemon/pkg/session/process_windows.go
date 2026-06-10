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

// SupportedOnPlatform is false on Windows: Execute drives the session shell
// with a POSIX wrapper script (see cmdWrapperFormat in execute.go) that
// neither cmd.exe nor PowerShell can run — the first write would kill the
// shell and leave clients polling for an exit code that never appears.
// Create refuses cleanly instead (see create.go), so the unconditionally
// registered /process/session routes degrade to honest errors rather than
// nil-service panics or wedged requests.
const SupportedOnPlatform = false
