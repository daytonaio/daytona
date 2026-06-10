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
	if command == "" {
		return exec.Command(shell)
	}
	if IsPowerShell(shell) {
		return exec.Command(shell, append(GetShellArgs(shell), command)...)
	}
	// cmd.Args is bypassed when SysProcAttr.CmdLine is set; keep it
	// populated anyway so logs and debuggers show the intended invocation.
	cmd := exec.Command(shell, "/C", command)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CmdLine: `"` + shell + `" /C ` + command,
	}
	return cmd
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
