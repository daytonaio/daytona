//go:build windows

// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

// Windows command invocation for cmd.exe sessions.
//
// The user command is CALLed by the session shell itself, so cd and set
// changes persist across commands within a session, matching the POSIX
// wrapper sourcing cmd.sh into the live shell.
//
// Known, accepted limitations:
//   - no live log streaming: the labeled log is materialized only once the
//     command completes;
//   - stdout/stderr interleaving across the two streams is not preserved
//     (ordering within each stream is);
//   - commands run with cmd.exe batch semantics: `for` variables need %%v,
//     unpaired %% signs are consumed by expansion, and single lines are
//     limited to 8191 characters;
//   - interactive stdin (SendInput) is unsupported because there is no input
//     FIFO: run.bat attaches the command's stdin to NUL.

package session

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// buildCommandInvocation persists the per-command artifacts under logDir and
// returns the line written to the session shell's stdin to run them.
//
// It writes three files:
//   - cmd.bat:   the raw user command bytes (cmd.exe batch semantics are the
//     Windows analogue of sourcing cmd.sh),
//   - run.bat:   CALLs cmd.bat in the session shell's own context with stdin
//     attached to NUL and stdout/stderr redirected to out.raw/err.raw, then
//     runs merge.ps1,
//   - merge.ps1: labels every out.raw/err.raw line with the stdout or stderr
//     prefix, appends them to the log, and writes the exit_code file once the
//     log is complete.
//
// out.raw and err.raw are created by run.bat's redirection, not by Go.
//
// The session cmd.exe executes the returned `call "run.bat"` line
// synchronously, which serializes commands within a session exactly like the
// POSIX wrapper blocking the shell. The async parameter is accepted and
// ignored: there is no stdin holder on Windows, so async and sync commands
// are invoked identically.
func buildCommandInvocation(logDir, logFilePath, exitCodeFilePath, inputFilePath, cmd string, async bool) (string, error) {
	cmdFilePath := filepath.Join(logDir, "cmd.bat")
	runBatFilePath := filepath.Join(logDir, "run.bat")
	mergePsFilePath := filepath.Join(logDir, "merge.ps1")
	outRawFilePath := filepath.Join(logDir, "out.raw")
	errRawFilePath := filepath.Join(logDir, "err.raw")

	// Percent signs are expanded by cmd.exe even inside double quotes and
	// would corrupt the embedded paths (client-chosen sessionIds land in
	// logDir).
	for _, p := range []string{logDir, logFilePath, exitCodeFilePath, cmdFilePath, runBatFilePath, mergePsFilePath, outRawFilePath, errRawFilePath} {
		if strings.ContainsAny(p, "\"%\r\n") {
			return "", fmt.Errorf("path %q contains a double quote, percent sign, carriage return, or newline and cannot be embedded in the batch/PowerShell wrappers", p)
		}
	}

	if err := os.WriteFile(cmdFilePath, []byte(cmd), 0600); err != nil {
		return "", fmt.Errorf("failed to write command file: %w", err)
	}

	// run.bat is CALLed by the session shell, so `call cmd.bat` runs in the
	// session's own context and cd/set persist across commands. Batch files
	// expand %errorlevel%/%DAYTONA_SESSION_EC% per line at execution time, so
	// the values are correct and parse-safe. The powershell.exe path goes
	// through %SystemRoot% instead of PATH (PATH lookup in the service
	// session is what GetShell deliberately avoids). The `if not exist` line
	// is the fallback exit-code writer so a PowerShell startup failure can
	// never leave the sync poll loop spinning forever. The final set clears
	// the temp variable from the session environment.
	runBat := "@echo off\r\n" +
		"call \"" + cmdFilePath + "\" < NUL >> \"" + outRawFilePath + "\" 2>> \"" + errRawFilePath + "\"\r\n" +
		"set DAYTONA_SESSION_EC=%errorlevel%\r\n" +
		"\"%SystemRoot%\\System32\\WindowsPowerShell\\v1.0\\powershell.exe\" -NoProfile -NonInteractive -ExecutionPolicy Bypass -File \"" + mergePsFilePath + "\" %DAYTONA_SESSION_EC%\r\n" +
		"if not exist \"" + exitCodeFilePath + "\" (>\"" + exitCodeFilePath + "\" echo %DAYTONA_SESSION_EC%)\r\n" +
		"set \"DAYTONA_SESSION_EC=\"\r\n"
	if err := os.WriteFile(runBatFilePath, []byte(runBat), 0600); err != nil {
		return "", fmt.Errorf("failed to write run.bat: %w", err)
	}

	mergePs := fmt.Sprintf(mergePsFormat, psQuote(logFilePath), psQuote(outRawFilePath), psQuote(errRawFilePath), psQuote(exitCodeFilePath))
	if err := os.WriteFile(mergePsFilePath, []byte(strings.ReplaceAll(mergePs, "\n", "\r\n")), 0600); err != nil {
		return "", fmt.Errorf("failed to write merge.ps1: %w", err)
	}

	return "call \"" + runBatFilePath + "\"\r\n", nil
}

// psQuote single-quotes s for PowerShell, escaping embedded ' by doubling.
func psQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

const mergePsFormat = `$ErrorActionPreference = 'Stop'
$code = 1
if ($args.Count -ge 1 -and $args[0] -match '^-?\d+$') { $code = [int]$args[0] }
$outPrefix = New-Object string ([char]1), 3
$errPrefix = New-Object string ([char]2), 3
try {
    # Raw redirected cmd.exe output is in the console OEM codepage; decode it
    # so the shared log is valid UTF-8. UTF8Encoding($false) avoids a BOM;
    # FileShare ReadWrite lets the daemon read or recreate the log
    # concurrently.
    $oem = [System.Text.Encoding]::GetEncoding([System.Globalization.CultureInfo]::CurrentCulture.TextInfo.OEMCodePage)
    $enc = New-Object System.Text.UTF8Encoding $false
    $fs = New-Object System.IO.FileStream %[1]s, ([System.IO.FileMode]::Append), ([System.IO.FileAccess]::Write), ([System.IO.FileShare]::ReadWrite)
    $writer = New-Object System.IO.StreamWriter $fs, $enc
    try {
        # out.raw/err.raw are guaranteed to exist: run.bat's >> redirections
        # create them before cmd.bat runs. Labeling them here preserves the
        # stdout/stderr demux contract (stream interleaving across the two
        # files is not preserved).
        foreach ($line in [System.IO.File]::ReadLines(%[2]s, $oem)) { $writer.WriteLine($outPrefix + $line) }
        foreach ($line in [System.IO.File]::ReadLines(%[3]s, $oem)) { $writer.WriteLine($errPrefix + $line) }
    } finally {
        $writer.Dispose()
    }
} finally {
    # Written last so pollers never observe an exit code before the log is
    # complete; written even when the merge fails so callers cannot hang.
    [System.IO.File]::WriteAllText(%[4]s, [string]$code)
}
`
