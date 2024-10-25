// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"os/exec"
	"path"
	"strings"

	"github.com/daytonaio/daytona/cmd/daytona/config"
)

func GetHomeDir(activeProfile config.Profile, targetId string, projectName string, gpgKey string) (string, error) {
	err := config.EnsureSshConfigEntryAdded(activeProfile.Id, targetId, projectName, gpgKey)
	if err != nil {
		return "", err
	}

	projectHostname := config.GetProjectHostname(activeProfile.Id, targetId, projectName)

	homeDir, err := exec.Command("ssh", projectHostname, "echo", "$HOME").Output()
	if err != nil {
		return "", err
	}

	return strings.TrimRight(string(homeDir), "\n"), nil
}

func GetProjectDir(activeProfile config.Profile, targetId string, projectName string, gpgKey string) (string, error) {
	err := config.EnsureSshConfigEntryAdded(activeProfile.Id, targetId, projectName, gpgKey)
	if err != nil {
		return "", err
	}

	projectHostname := config.GetProjectHostname(activeProfile.Id, targetId, projectName)

	daytonaProjectDir, err := exec.Command("ssh", projectHostname, "echo", "$DAYTONA_PROJECT_DIR").Output()
	if err != nil {
		return "", err
	}

	if strings.TrimRight(string(daytonaProjectDir), "\n") != "" {
		return strings.TrimRight(string(daytonaProjectDir), "\n"), nil
	}

	homeDir, err := GetHomeDir(activeProfile, targetId, projectName, gpgKey)
	if err != nil {
		return "", err
	}

	return path.Join(homeDir, projectName), nil
}
