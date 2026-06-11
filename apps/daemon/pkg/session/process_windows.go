//go:build windows

// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package session

import (
	"os/exec"
	"syscall"
)

// setNewProcessGroup is a no-op on Windows. Process group semantics are a
// POSIX concept; Windows uses Job Objects (not wired up here).
func setNewProcessGroup(cmd *exec.Cmd) {}

// killProcessGroup is a no-op on Windows. Callers also invoke
// signalProcessTree, which is portable and handles termination via gopsutil.
func killProcessGroup(pid int, sig syscall.Signal) {}

// SupportedOnPlatform is true on Windows: Execute drives the session shell
// through the Windows-native wrapper (wrapper_windows.go — cmd.bat CALLed by
// run.bat in the session shell's own context, merge.ps1 stream labeling), so
// cd/set persist across commands and exit codes demux exactly like the POSIX
// wrapper on Linux.
const SupportedOnPlatform = true
