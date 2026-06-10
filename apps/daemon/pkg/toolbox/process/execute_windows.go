//go:build windows

// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package process

import (
	"os"
	"os/exec"
	"strconv"
	"time"

	common_errors "github.com/daytonaio/common-go/pkg/errors"

	"github.com/daytonaio/daemon/pkg/common"
)

// buildExecCmd runs the command through the configured shell (cmd.exe or
// PowerShell). Legacy `sh -c "... | base64 -d | sh"` wrappers sent by older
// SDKs are translated into a native shell command first, with extracted env
// vars applied to the process environment — same convention as the SSH
// non-PTY handler (ssh/server_windows.go).
//
// Unlike the Linux build, which pipes the command to the shell via stdin to
// avoid ARG_MAX, the command here is passed on the command line
// (`<shell> /C <command>`), so Windows length limits apply: cmd.exe rejects
// lines longer than 8191 characters ("The input line is too long") and
// CreateProcess caps the full command line at 32767. Oversized commands fail
// loudly with a non-zero exit. The divergence is deliberate: the obvious
// workaround — writing the command to a temp .cmd and running that — is not
// semantically transparent (batch files echo commands and expand % args
// differently than `/C`), which would turn today's loud failure into silent
// mangling. A semantics-preserving large-command path is future work.
func buildExecCmd(command string) *exec.Cmd {
	parsedCommand, envVars := common.ParseShellWrapper(command)
	cmd := common.NewShellCommand(common.GetShell(), parsedCommand)
	common.ApplyEnvs(cmd, envVars)
	// I/O-drain backstop: if the shell dies but an orphaned descendant still
	// holds the inherited output pipe, stop waiting for EOF after this delay
	// instead of wedging the handler forever (golang/go#23019). The Linux
	// build gets the equivalent from childreap.Wait's hangTimeout.
	cmd.WaitDelay = 30 * time.Second
	return cmd
}

// killExecProcessGroup kills the shell and all of its descendants.
// taskkill /T walks the parent-PID tree — the closest Windows equivalent of
// the process-group SIGKILL in execute_linux.go. Mirrors
// coderun.killProcessGroupHard.
func killExecProcessGroup(cmd *exec.Cmd) error {
	if err := exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(cmd.Process.Pid)).Run(); err == nil {
		return nil
	}
	// Fall back to killing the immediate process (e.g. taskkill unavailable).
	p, err := os.FindProcess(cmd.Process.Pid)
	if err != nil {
		return err
	}
	return p.Kill()
}
