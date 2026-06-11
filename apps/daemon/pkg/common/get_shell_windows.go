//go:build windows

// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"syscall"
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

// ShellCommand builds an exec.Cmd that runs commandLine through shell.
//
// For cmd.exe the command line must bypass Go's default per-argument
// quoting: os/exec escapes embedded quotes as \" (MSVCRT rules), which
// cmd.exe does not understand, so quoted arguments would reach the child
// program with literal quote characters. Instead the raw line is passed
// via SysProcAttr.CmdLine using `cmd /S /C "<commandLine>"` — with /S,
// cmd strips only the outer quote pair and runs the command verbatim.
// PowerShell parses \" natively, so the default quoting is correct there.
func ShellCommand(shell, commandLine string) *exec.Cmd {
	if IsPowerShell(shell) {
		return exec.Command(shell, "-NoProfile", "-NonInteractive", "-Command", commandLine)
	}
	cmd := exec.Command(shell)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CmdLine: fmt.Sprintf(`"%s" /S /C "%s"`, shell, commandLine),
	}
	return cmd
}

// IsPowerShell checks if the shell path refers to PowerShell
func IsPowerShell(shell string) bool {
	lower := strings.ToLower(shell)
	return strings.HasSuffix(lower, "pwsh.exe") || strings.HasSuffix(lower, "powershell.exe")
}
