//go:build windows

// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package coderun

import (
	"net/http"
	"os/exec"

	common_errors "github.com/daytonaio/common-go/pkg/errors"

	"github.com/daytonaio/daemon/pkg/common"
)

// codeRunPlatformError returns 501 on Windows: the run commands emitted by
// the language toolboxes (`printf '%s' '<b64>' | base64 -d | python3 -u -`)
// are POSIX pipelines that cmd.exe/PowerShell cannot execute, and piping them
// into a bare shell's stdin would capture the shell banner and prompts as
// output with exit code 0. Report honestly until the run commands are ported.
func codeRunPlatformError() error {
	return common_errors.NewCustomError(http.StatusNotImplemented, "code execution is not yet implemented on Windows", "NOT_IMPLEMENTED")
}

// setNewProcessGroup is a no-op on Windows; killProcessGroupHard tears down
// the whole process tree via taskkill /T instead of POSIX process groups.
func setNewProcessGroup(cmd *exec.Cmd) {}

// killProcessGroupHard kills the process and all of its descendants, the
// closest Windows equivalent of the process-group SIGKILL in process_linux.go.
func killProcessGroupHard(pid int) error {
	return common.KillProcessTree(pid)
}
