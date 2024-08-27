// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ide

import (
	"fmt"
	"os/exec"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/build/devcontainer"

	log "github.com/sirupsen/logrus"
)

func OpenCursor(activeProfile config.Profile, workspaceId string, projectName string, projectProviderMetadata string) error {
	CheckAndAlertCursorInstalled()

	projectHostname := config.GetProjectHostname(activeProfile.Id, workspaceId, projectName)

	projectDir, err := util.GetProjectDir(activeProfile, workspaceId, projectName)
	if err != nil {
		return err
	}

	commandArgument := fmt.Sprintf("vscode-remote://ssh-remote+%s/%s", projectHostname, projectDir)

	vscCommand := exec.Command("cursor", "--folder-uri", commandArgument, "--disable-extension", "ms-vscode-remote.remote-containers")

	err = vscCommand.Run()
	if err != nil {
		return err
	}

	return setupVSCodeCustomizations(projectHostname, projectProviderMetadata, devcontainer.Vscode, "*/.cursor-server/*/bin/cursor-server", "$HOME/.cursor-server/data/Machine/settings.json", ".daytona-customizations-lock-cursor")
}

func CheckAndAlertCursorInstalled() {
	if err := isCursorInstalled(); err != nil {
		redBold := "\033[1;31m" // ANSI escape code for red and bold
		reset := "\033[0m"      // ANSI escape code to reset text formatting

		errorMessage := "Please install Cursor and ensure it's in your PATH. "
		infoMessage := "After installing the IDE, run the `Install 'cursor' command` from the command palette."

		log.Error(redBold + errorMessage + reset + infoMessage)

		return
	}
}

func isCursorInstalled() error {
	_, err := exec.LookPath("cursor")
	return err
}
