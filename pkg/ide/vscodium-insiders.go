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

func OpenVScodiumInsiders(activeProfile config.Profile, workspaceId, repoName string, workspaceProviderMetadata string, gpgkey *string) error {
	path, err := GetCodiumInsidersBinaryPath()
	if err != nil {
		return err
	}
	// Install the extension if not installed
	err = installExtension(path, requiredExtension)
	if err != nil {
		return err
	}

	workspaceHostname := config.GetHostname(activeProfile.Id, workspaceId)

	workspaceDir, err := util.GetWorkspaceDir(activeProfile, workspaceId, repoName, gpgkey)
	if err != nil {
		return err
	}

	commandArgument := fmt.Sprintf("vscode-remote://ssh-remote+%s/%s", workspaceHostname, workspaceDir)

	codiumInsidersCommand := exec.Command(path, "--disable-extension", "ms-vscode-remote.remote-containers", "--folder-uri", commandArgument)

	err = codiumInsidersCommand.Run()
	if err != nil {
		return err
	}

	if workspaceProviderMetadata == "" {
		return nil
	}

	return setupVSCodeCustomizations(workspaceHostname, workspaceProviderMetadata, devcontainer.Vscode, "*/.vscodium-server-insiders/*/bin/codium-server-insiders", "$HOME/.vscodium-server-insiders/data/Machine/settings.json", ".daytona-customizations-lock-codium-insiders")
}

func GetCodiumInsidersBinaryPath() (string, error) {
	path, err := exec.LookPath("codium-insiders")
	if err == nil {
		return path, err
	}

	redBold := "\033[1;31m" // ANSI escape code for red and bold
	reset := "\033[0m"      // ANSI escape code to reset text formatting

	errorMessage := "Please install VScodium Insiders from https://github.com/VSCodium/vscodium-insiders\n\n"
	infoMessage := `
If you have already installed VScodium Insiders, please ensure it is added to your PATH environment variable.
You can verify this by running the following command in your terminal:

	codium-insiders
`

	return "", errors.New(redBold + errorMessage + reset + infoMessage)
}
