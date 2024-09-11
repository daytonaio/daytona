// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ide

import (
	"fmt"
	"os/exec"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	log "github.com/sirupsen/logrus"
)

const (
	redBold = "\033[1;31m" // ANSI escape code for red and bold
	reset   = "\033[0m"    // ANSI escape code to reset text formatting
)

func OpenZed(activeProfile config.Profile, workspaceId string, projectName string) error {

	if err := IsZedInstalled(); err != nil {
		return err
	}

	projectHostname := config.GetProjectHostname(activeProfile.Id, workspaceId, projectName)
	projectDir, err := util.GetProjectDir(activeProfile, workspaceId, projectName)
	if err != nil {
		return err
	}

	zedCommand := fmt.Sprintf("zed ssh://%s:443%s", projectHostname, projectDir)

	runZedCommand := exec.Command(zedCommand)
	err = runZedCommand.Run()
	if err != nil {
		return err
	}

	return nil
}

func IsZedInstalled() error {
	_, err := exec.LookPath("zed")
	if err != nil {

		errorMessage := "Please install Zed and ensure it's in your PATH. "
		infoMessage := "\nMore information: Install Zed from : https://zed.dev/docs/getting-started"

		log.Error(redBold + errorMessage + reset + infoMessage)

		return err
	}

	return nil
}
