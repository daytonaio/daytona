//go:build windows

// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"os"
	"os/exec"
	"strconv"
)

// KillProcessTree kills the process and all of its descendants.
// taskkill /T walks the live parent-PID tree, the closest Windows equivalent
// of the process-group SIGKILL on Linux: grandchildren spawned by a cmd.exe
// wrapper inherit the output pipes and would otherwise keep readers blocked
// long after the parent was killed. Descendants whose intermediate parent
// already exited are not reachable this way; killing those would need Job
// Objects.
func KillProcessTree(pid int) error {
	if err := exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(pid)).Run(); err == nil {
		return nil
	}
	// Fall back to killing the immediate process (e.g. taskkill unavailable).
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return p.Kill()
}
