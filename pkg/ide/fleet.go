// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ide

import (
	"fmt"
	"os/exec"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/pkg/browser"
	log "github.com/sirupsen/logrus"
)

func OpenFleet(activeProfile config.Profile, workspaceId string, projectName string, gpgForward bool) error {
	if err := CheckFleetInstallation(); err != nil {
		return err
	}

	projectHostname := config.GetProjectHostname(activeProfile.Id, workspaceId, projectName)
	projectDir, err := util.GetProjectDir(activeProfile, workspaceId, projectName, gpgForward)
	if err != nil {
		return err
	}

	ideURL := fmt.Sprintf("fleet://fleet.ssh/%s?pwd=%s", projectHostname, projectDir)

	err = browser.OpenURL(ideURL)
	if err != nil {
		return err
	}

	return nil
}

func CheckFleetInstallation() error {
	_, err := exec.LookPath("fleet")
	if err != nil {
		redBold := "\033[1;31m" // ANSI escape code for red and bold
		reset := "\033[0m"      // ANSI escape code to reset text formatting

		errorMessage := "Please install Fleet using JetBrains Toolbox and ensure it's in your PATH. "
		infoMessage := "\nMore information: \n1) Install JetBrains Toolbox: 'https://www.jetbrains.com/toolbox-app/'\n2) Install Fleet: Using JetBrains Toolbox, search for 'Fleet' and install it. \n3) Ensure Fleet is in your PATH and keep your 'Shell script name' as 'fleet': 'https://www.jetbrains.com/help/fleet/launch-from-cli.html'"

		log.Error(redBold + errorMessage + reset + infoMessage)

		return err
	}

	return nil
}
