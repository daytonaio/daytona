// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ide

import (
	"errors"
	"fmt"
	"os/exec"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/build/devcontainer"
)

func OpenWindsurf(activeProfile config.Profile, workspaceId string, projectName string, projectProviderMetadata string, gpgkey string) error {
	path, err := GetWindsurfBinaryPath()
	if err != nil {
		return err
	}

	projectHostname := config.GetProjectHostname(activeProfile.Id, workspaceId, projectName)

	projectDir, err := util.GetProjectDir(activeProfile, workspaceId, projectName, gpgkey)
	if err != nil {
		return err
	}

	commandArgument := fmt.Sprintf("vscode-remote://ssh-remote+%s/%s", projectHostname, projectDir)

	windsurfCommand := exec.Command(path, "--disable-extension", "ms-vscode-remote.remote-containers", "--folder-uri", commandArgument)

	err = windsurfCommand.Run()
	if err != nil {
		return err
	}

	if projectProviderMetadata == "" {
		return nil
	}

	return setupVSCodeCustomizations(projectHostname, projectProviderMetadata, devcontainer.Vscode, "*/.windsurf-server/*/bin/windsurf-server", "$HOME/.windsurf-server/data/Machine/settings.json", ".daytona-customizations-lock-windsurf")
}

func GetWindsurfBinaryPath() (string, error) {
	path, err := exec.LookPath("windsurf")
	if err == nil {
		return path, err
	}

	redBold := "\033[1;31m" // ANSI escape code for red and bold
	reset := "\033[0m"      // ANSI escape code to reset text formatting

	errorMessage := "Please install Windsurf from https://codeium.com/windsurf/download and ensure it's in your PATH.\n"

	return "", errors.New(redBold + errorMessage + reset)
}
