// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"os"
	"os/exec"
)

// GetShell returns the path to the preferred shell on Windows.
// By default uses cmd.exe for faster execution. Set DAYTONA_SHELL env var
// to override (e.g., "powershell.exe" or "pwsh.exe").
func GetShell() string {
	// Allow override via environment variable
	if shell := os.Getenv("DAYTONA_SHELL"); shell != "" {
		if path, err := exec.LookPath(shell); err == nil {
			return path
		}
	}

	// Default to cmd.exe for faster execution (~100ms vs ~2s for PowerShell)
	if cmd, err := exec.LookPath("cmd.exe"); err == nil {
		return cmd
	}

	// Fallback to PowerShell if cmd.exe not available
	if powershell, err := exec.LookPath("powershell.exe"); err == nil {
		return powershell
	}

	// Check for PowerShell Core
	if pwsh, err := exec.LookPath("pwsh.exe"); err == nil {
		return pwsh
	}

	// Last resort: check common paths
	commonPaths := []string{
		`C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe`,
		`C:\Windows\System32\cmd.exe`,
	}

	for _, p := range commonPaths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	// Return powershell.exe and let the caller handle errors
	return "powershell.exe"
}

// GetShellArgs returns the arguments needed to execute a command in the shell.
// For PowerShell: -NoProfile -NonInteractive -Command
// For cmd.exe: /C
func GetShellArgs(shell string) []string {
	// Check if it's PowerShell
	if isPowerShell(shell) {
		return []string{"-NoProfile", "-NonInteractive", "-Command"}
	}
	// Assume cmd.exe
	return []string{"/C"}
}

// isPowerShell checks if the shell path refers to PowerShell (internal use)
func isPowerShell(shell string) bool {
	return IsPowerShell(shell)
}

// IsPowerShell checks if the shell path refers to PowerShell
func IsPowerShell(shell string) bool {
	return shell == "pwsh.exe" ||
		shell == "powershell.exe" ||
		len(shell) > 8 && (shell[len(shell)-8:] == "pwsh.exe" || shell[len(shell)-14:] == "powershell.exe")
}
