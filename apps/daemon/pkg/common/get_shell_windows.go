//go:build windows

// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
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

// NewShellCommand returns an exec.Cmd that runs command through shell.
// An empty command yields an interactive shell invocation.
//
// PowerShell parses the CommandLineToArgvW-style quoting Go produces for
// argv elements (embedded quotes escaped as \"), so the command is passed
// as a regular argument. cmd.exe does NOT understand that escaping
// (golang/go#17149): any command containing a double quote would arrive
// mangled. For cmd.exe the raw command line is therefore set verbatim via
// SysProcAttr.CmdLine, as the os/exec documentation prescribes for
// cmd.exe-style parsers.
func NewShellCommand(shell, command string) *exec.Cmd {
	var cmd *exec.Cmd
	switch {
	case command == "":
		cmd = exec.Command(shell)
	case IsPowerShell(shell):
		cmd = exec.Command(shell, "-NoProfile", "-NonInteractive", "-Command", command)
	default:
		// cmd.Args is bypassed when SysProcAttr.CmdLine is set; keep it
		// populated anyway so logs and debuggers show the intended invocation.
		cmd = exec.Command(shell, "/C", command)
		cmd.SysProcAttr = &syscall.SysProcAttr{
			CmdLine: `"` + shell + `" /C ` + command,
		}
	}
	// I/O-drain backstop: cmd.Wait blocks until the internal pipe-copy
	// goroutines hit EOF when Stdout/Stderr are not *os.File, so an orphaned
	// descendant that inherited the handles (e.g. `start /B server`) would
	// wedge the caller forever (golang/go#23019). Bound that drain phase the
	// way childreap.Wait's hangTimeout does on Linux. childreap.Wait
	// recovers the real exit code when the backstop fires (exec.ErrWaitDelay).
	cmd.WaitDelay = 30 * time.Second
	return cmd
}

// IsPowerShell checks if the shell path refers to PowerShell
func IsPowerShell(shell string) bool {
	lower := strings.ToLower(shell)
	return strings.HasSuffix(lower, "pwsh.exe") || strings.HasSuffix(lower, "powershell.exe")
}
