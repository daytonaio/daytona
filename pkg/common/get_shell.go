// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

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
