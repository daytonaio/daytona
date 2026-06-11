//go:build windows

// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"os"
	"os/exec"
	"path/filepath"
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
	// Resolve taskkill from System32 instead of PATH: service sessions can
	// run with a minimal or tampered PATH, silently degrading the tree kill
	// to the parent-only fallback.
	taskkill := "taskkill"
	if systemRoot := os.Getenv("SystemRoot"); systemRoot != "" {
		taskkill = filepath.Join(systemRoot, "System32", "taskkill.exe")
	}
	if err := exec.Command(taskkill, "/T", "/F", "/PID", strconv.Itoa(pid)).Run(); err == nil {
		return nil
	}
	// Fall back to killing the immediate process (e.g. taskkill unavailable).
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return p.Kill()
}
