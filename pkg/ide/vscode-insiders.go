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

func OpenVSCodeInsiders(activeProfile config.Profile, workspaceId string, workspaceProviderMetadata string, gpgKey string) error {
	path, err := GetVSCodeInsidersBinaryPath()
	if err != nil {
		return err
	}
	err = installRemoteSSHExtension(path)
	if err != nil {
		return err
	}

	workspaceHostname := config.GetWorkspaceHostname(activeProfile.Id, workspaceId)

	workspaceDir, err := util.GetWorkspaceDir(activeProfile, workspaceId, gpgKey)
	if err != nil {
		return err
	}

	commandArgument := fmt.Sprintf("vscode-remote://ssh-remote+%s/%s", workspaceHostname, workspaceDir)

	vscCommand := exec.Command(path, "--disable-extension", "ms-vscode-remote.remote-containers", "--folder-uri", commandArgument)

	err = vscCommand.Run()
	if err != nil {
		return err
	}

	if workspaceProviderMetadata == "" {
		return nil
	}

	return setupVSCodeCustomizations(workspaceHostname, workspaceProviderMetadata, devcontainer.Vscode, "*/.vscode-server-insiders/*/bin/code-server-insiders", "$HOME/.vscode-server-insiders/data/Machine/settings.json", ".daytona-customizations-lock-vscode-insiders")
}

func GetVSCodeInsidersBinaryPath() (string, error) {
	path, err := exec.LookPath("code-insiders")
	if err == nil {
		return path, err
	}

	redBold := "\033[1;31m" // ANSI escape code for red and bold
	reset := "\033[0m"      // ANSI escape code to reset text formatting

	errorMessage := "Please install Visual Studio Code Insiders from https://code.visualstudio.com/insiders/\n\n"
	infoMessage := `
If you have already installed Visual Studio Code Insiders, please ensure it is added to your PATH environment variable.
You can verify this by running the following command in your terminal:

	code-insiders
`

	return "", errors.New(redBold + errorMessage + reset + infoMessage)
}
