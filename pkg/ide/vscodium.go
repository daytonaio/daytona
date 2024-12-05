// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ide

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/build/devcontainer"
)

const requiredExtension = "jeanp413.open-remote-ssh"

func OpenVScodium(activeProfile config.Profile, workspaceId string, workspaceProviderMetadata string, gpgkey *string) error {
	path, err := GetCodiumBinaryPath()
	if err != nil {
		return err
	}
	// Install the extension if not installed
	err = installExtension(path, requiredExtension)
	if err != nil {
		return err
	}

	workspaceHostname := config.GetHostname(activeProfile.Id, workspaceId)

	workspaceDir, err := util.GetWorkspaceDir(activeProfile, workspaceId, gpgkey)
	if err != nil {
		return err
	}

	commandArgument := fmt.Sprintf("vscode-remote://ssh-remote+%s/%s", workspaceHostname, workspaceDir)

	codiumCommand := exec.Command(path, "--disable-extension", "ms-vscode-remote.remote-containers", "--folder-uri", commandArgument)

	err = codiumCommand.Run()
	if err != nil {
		return err
	}

	if workspaceProviderMetadata == "" {
		return nil
	}

	return setupVSCodeCustomizations(workspaceHostname, workspaceProviderMetadata, devcontainer.Vscode, "*/.vscodium-server/*/bin/codium-server", "$HOME/.vscodium-server/data/Machine/settings.json", ".daytona-customizations-lock-codium")
}

func GetCodiumBinaryPath() (string, error) {
	path, err := exec.LookPath("codium")
	if err == nil {
		return path, err
	}

	redBold := "\033[1;31m" // ANSI escape code for red and bold
	reset := "\033[0m"      // ANSI escape code to reset text formatting

	errorMessage := "Please install VScodium from https://vscodium.com/ and ensure it's in your PATH.\n"

	return "", errors.New(redBold + errorMessage + reset)
}

func installExtension(binaryPath string, extensionName string) error {
	// Check if the required extension is installed
	output, err := exec.Command(binaryPath, "--list-extensions").Output()
	if err != nil {
		return err
	}

	if !strings.Contains(string(output), extensionName) {
		fmt.Printf("Installing %s extension...\n", extensionName)
		err = exec.Command(binaryPath, "--install-extension", extensionName).Run()
		if err != nil {
			return err
		}
		fmt.Printf("%s extension successfully installed\n", extensionName)
	}
	return nil
}
