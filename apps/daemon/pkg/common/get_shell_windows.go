//go:build windows

// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"log/slog"
	"os"
	"os/exec"
	"strings"
)

// GetShell returns the path to the preferred shell on Windows.
// By default uses cmd.exe for faster execution. Set DAYTONA_SHELL env var
// to override (e.g., "powershell.exe" or "pwsh.exe").
func GetShell() string {
	if shell := os.Getenv("DAYTONA_SHELL"); shell != "" {
		if path, err := exec.LookPath(shell); err == nil {
			slog.Debug("GetShell: using DAYTONA_SHELL override", "path", path)
			return path
		}
		slog.Warn("GetShell: DAYTONA_SHELL not found, falling back", "shell", shell)
	}

	// Try direct path to cmd.exe first (most reliable, ~100ms vs ~2s for PowerShell)
	cmdExePath := `C:\Windows\System32\cmd.exe`
	if _, err := os.Stat(cmdExePath); err == nil {
		slog.Debug("GetShell: using cmd.exe", "path", cmdExePath)
		return cmdExePath
	} else {
		slog.Warn("GetShell: cmd.exe not found at default path", "path", cmdExePath, "error", err)
	}

	if cmd, err := exec.LookPath("cmd.exe"); err == nil {
		slog.Debug("GetShell: using cmd.exe from PATH", "path", cmd)
		return cmd
	}

	if powershell, err := exec.LookPath("powershell.exe"); err == nil {
		slog.Warn("GetShell: falling back to PowerShell (this will be slow!)", "path", powershell)
		return powershell
	}

	if pwsh, err := exec.LookPath("pwsh.exe"); err == nil {
		slog.Warn("GetShell: falling back to pwsh (this will be slow!)", "path", pwsh)
		return pwsh
	}

	commonPaths := []string{
		`C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe`,
	}

	for _, p := range commonPaths {
		if _, err := os.Stat(p); err == nil {
			slog.Warn("GetShell: using fallback path (this will be slow!)", "path", p)
			return p
		}
	}

	slog.Error("GetShell: no shell found, defaulting to powershell.exe")
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
	lower := strings.ToLower(shell)
	return strings.HasSuffix(lower, "pwsh.exe") || strings.HasSuffix(lower, "powershell.exe")
}
