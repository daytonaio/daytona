// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"os"
	"os/exec"
	"strings"
)

func GetShell() string {
	out, err := exec.Command("sh", "-c", "grep '^[^#]' /etc/shells").Output()
	if err != nil {
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

// GetShellArgs returns the shell path and no-init flags to prevent the shell
// from reading init files that may consume stdin bytes or run exit.
func GetShellArgs() []string {
	shell := GetShell()
	switch {
	case shell == "zsh", strings.HasSuffix(shell, "/zsh"):
		return []string{shell, "-f"}
	case shell == "bash", strings.HasSuffix(shell, "/bash"):
		return []string{shell, "--norc", "--noprofile"}
	default:
		return []string{shell}
	}
}
