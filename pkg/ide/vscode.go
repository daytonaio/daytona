// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ide

import (
	"fmt"
	"os/exec"
	"path"

	"github.com/daytonaio/daytona/cmd/daytona/config"

	log "github.com/sirupsen/logrus"
)

func OpenVSCode(activeProfile config.Profile, workspaceId string, projectName string) error {
	err := config.EnsureSshConfigEntryAdded(activeProfile.Id, workspaceId, projectName)
	if err != nil {
		return err
	}

	checkAndAlertVSCodeInstalled()

	projectHostname := config.GetProjectHostname(activeProfile.Id, workspaceId, projectName)

	commandArgument := fmt.Sprintf("vscode-remote://ssh-remote+%s/%s", projectHostname, path.Join("/workspaces", projectName))

	var vscCommand *exec.Cmd = exec.Command("code", "--folder-uri", commandArgument)

	return vscCommand.Run()
}

func checkAndAlertVSCodeInstalled() {
	if err := isVSCodeInstalled(); err != nil {
		redBold := "\033[1;31m" // ANSI escape code for red and bold
		reset := "\033[0m"      // ANSI escape code to reset text formatting

		errorMessage := "Please install Visual Studio Code and ensure it's in your PATH. "
		infoMessage := "More information on: 'https://code.visualstudio.com/docs/editor/command-line#_launching-from-command-line'"

		log.Error(redBold + errorMessage + reset + infoMessage)

		return
	}
}

func isVSCodeInstalled() error {
	_, err := exec.LookPath("code")
	return err
}
