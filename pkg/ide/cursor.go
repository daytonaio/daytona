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

func OpenCursor(activeProfile config.Profile, targetId string, workspaceName string, workspaceProviderMetadata string, gpgkey string) error {
	path, err := GetCursorBinaryPath()
	if err != nil {
		return err
	}

	workspaceHostname := config.GetWorkspaceHostname(activeProfile.Id, targetId, workspaceName)

	workspaceDir, err := util.GetWorkspaceDir(activeProfile, targetId, workspaceName, gpgkey)
	if err != nil {
		return err
	}

	commandArgument := fmt.Sprintf("vscode-remote://ssh-remote+%s/%s", workspaceHostname, workspaceDir)

	cursorCommand := exec.Command(path, "--disable-extension", "ms-vscode-remote.remote-containers", "--folder-uri", commandArgument)

	err = cursorCommand.Run()
	if err != nil {
		return err
	}

	if workspaceProviderMetadata == "" {
		return nil
	}

	return setupVSCodeCustomizations(workspaceHostname, workspaceProviderMetadata, devcontainer.Vscode, "*/.cursor-server/*/bin/cursor-server", "$HOME/.cursor-server/data/Machine/settings.json", ".daytona-customizations-lock-cursor")
}

func GetCursorBinaryPath() (string, error) {
	path, err := exec.LookPath("cursor")
	if err == nil {
		return path, err
	}

	// Cursor asks the user if they want to override the 'code' binary
	path, err = exec.LookPath("code")
	if err == nil {
		// Check that the code binary is actually Cursor
		output, err := exec.Command(path, "--help").Output()
		if err == nil && strings.HasPrefix(string(output), "Cursor") {
			return path, nil
		}
	}

	redBold := "\033[1;31m" // ANSI escape code for red and bold
	reset := "\033[0m"      // ANSI escape code to reset text formatting

	errorMessage := "Please install Cursor from https://www.cursor.com/ and ensure it's in your PATH.\n"
	infoMessage := "After installing the IDE, run the `Install 'cursor' command` from the command palette."

	return "", errors.New(redBold + errorMessage + reset + infoMessage)
}
