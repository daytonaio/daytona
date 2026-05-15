// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"os"
	"os/exec"
	"strings"

	"github.com/daytonaio/daemon/pkg/childreap"
)

func GetShell() string {
	cmd := exec.Command("sh", "-c", "grep '^[^#]' /etc/shells")
	// childreap.Output (not cmd.Output) so the PID-1 reaper winning the
	// race against cmd.Wait doesn't drop us into the err != nil branch
	// and silently fall back to "sh" on sandboxes that actually have
	// zsh/bash available.
	out, exitCode, err := childreap.Output(cmd)
	if err != nil || exitCode != 0 {
		return "sh"
	}

	if strings.Contains(string(out), "/usr/bin/zsh") {
		return "/usr/bin/zsh"
	}

	if strings.Contains(string(out), "/bin/zsh") {
		return "/bin/zsh"
	}

	if strings.Contains(string(out), "/usr/bin/bash") {
		return "/usr/bin/bash"
	}

	if strings.Contains(string(out), "/bin/bash") {
		return "/bin/bash"
	}

	shellEnv, shellSet := os.LookupEnv("SHELL")

	if shellSet {
		return shellEnv
	}

	return "sh"
}
