// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"os/exec"
	"path"
	"strings"

	"github.com/daytonaio/daytona/cmd/daytona/config"
)

func GetHomeDir(activeProfile config.Profile, targetId string, workspaceName string, gpgKey string) (string, error) {
	err := config.EnsureSshConfigEntryAdded(activeProfile.Id, targetId, workspaceName, gpgKey)
	if err != nil {
		return "", err
	}

	workspaceHostname := config.GetWorkspaceHostname(activeProfile.Id, targetId, workspaceName)

	homeDir, err := exec.Command("ssh", workspaceHostname, "echo", "$HOME").Output()
	if err != nil {
		return "", err
	}

	return strings.TrimRight(string(homeDir), "\n"), nil
}

func GetWorkspaceDir(activeProfile config.Profile, targetId string, workspaceName string, gpgKey string) (string, error) {
	err := config.EnsureSshConfigEntryAdded(activeProfile.Id, targetId, workspaceName, gpgKey)
	if err != nil {
		return "", err
	}

	workspaceHostname := config.GetWorkspaceHostname(activeProfile.Id, targetId, workspaceName)

	daytonaWorkspaceDir, err := exec.Command("ssh", workspaceHostname, "echo", "$DAYTONA_WORKSPACE_DIR").Output()
	if err != nil {
		return "", err
	}

	if strings.TrimRight(string(daytonaWorkspaceDir), "\n") != "" {
		return strings.TrimRight(string(daytonaWorkspaceDir), "\n"), nil
	}

	homeDir, err := GetHomeDir(activeProfile, targetId, workspaceName, gpgKey)
	if err != nil {
		return "", err
	}

	return path.Join(homeDir, workspaceName), nil
}
