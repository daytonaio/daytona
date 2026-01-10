// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"os"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
)

// GetShell returns the path to the preferred shell on Windows.
// By default uses cmd.exe for faster execution. Set DAYTONA_SHELL env var
// to override (e.g., "powershell.exe" or "pwsh.exe").
func GetShell() string {
	// Allow override via environment variable
	if shell := os.Getenv("DAYTONA_SHELL"); shell != "" {
		if path, err := exec.LookPath(shell); err == nil {
			log.Debugf("GetShell: using DAYTONA_SHELL override: %s", path)
			return path
		}
		log.Warnf("GetShell: DAYTONA_SHELL=%s not found, falling back", shell)
	}

	// Try direct path to cmd.exe first (most reliable, ~100ms vs ~2s for PowerShell)
	cmdExePath := `C:\Windows\System32\cmd.exe`
	if _, err := os.Stat(cmdExePath); err == nil {
		log.Debugf("GetShell: using cmd.exe at %s", cmdExePath)
		return cmdExePath
	} else {
		log.Warnf("GetShell: cmd.exe not found at %s: %v", cmdExePath, err)
	}

	// Fallback to LookPath for cmd.exe
	if cmd, err := exec.LookPath("cmd.exe"); err == nil {
		log.Debugf("GetShell: using cmd.exe from PATH: %s", cmd)
		return cmd
	}

	// Fallback to PowerShell if cmd.exe not available
	if powershell, err := exec.LookPath("powershell.exe"); err == nil {
		log.Warnf("GetShell: falling back to PowerShell: %s (this will be slow!)", powershell)
		return powershell
	}

	// Check for PowerShell Core
	if pwsh, err := exec.LookPath("pwsh.exe"); err == nil {
		log.Warnf("GetShell: falling back to pwsh: %s (this will be slow!)", pwsh)
		return pwsh
	}

	// Last resort: check common paths
	commonPaths := []string{
		`C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe`,
	}

	for _, p := range commonPaths {
		if _, err := os.Stat(p); err == nil {
			log.Warnf("GetShell: using fallback path: %s (this will be slow!)", p)
			return p
		}
	}

	// Return powershell.exe and let the caller handle errors
	log.Errorf("GetShell: no shell found, defaulting to powershell.exe")
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
