// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/daytonaio/daytona/cmd/daytona/config"
)

func GetHomeDir(activeProfile config.Profile, workspaceId string, projectName string) (string, error) {
	err := config.EnsureSshConfigEntryAdded(activeProfile.Id, workspaceId, projectName)
	if err != nil {
		return "", err
	}

	projectHostname := config.GetProjectHostname(activeProfile.Id, workspaceId, projectName)

	homeDir, err := exec.Command("ssh", projectHostname, "echo", "$HOME").Output()
	if err != nil {
		return "", err
	}

	return strings.TrimRight(string(homeDir), "\n"), nil
}

func GetProjectDir(activeProfile config.Profile, workspaceId string, projectName string) (string, error) {
	homeDir, err := GetHomeDir(activeProfile, workspaceId, projectName)
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, projectName), nil
}
