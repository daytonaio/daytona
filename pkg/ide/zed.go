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
	"github.com/daytonaio/daytona/pkg/views"
)

func OpenZed(activeProfile config.Profile, workspaceId, repoName string, gpgKey *string) error {
	path, err := GetZedBinaryPath()
	if err != nil {
		return err
	}

	workspaceHostname := config.GetHostname(activeProfile.Id, workspaceId)
	workspaceDir, err := util.GetWorkspaceDir(activeProfile, workspaceId, repoName, gpgKey)
	if err != nil {
		return err
	}
	printDisclaimer()
	zedCmd := exec.Command(path, fmt.Sprintf("ssh://%s%s", workspaceHostname, workspaceDir))

	err = zedCmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func GetZedBinaryPath() (string, error) {
	path, err := exec.LookPath("zed")
	if err == nil {
		return path, err
	}

	redBold := "\033[1;31m" // ANSI escape code for red and bold
	reset := "\033[0m"      // ANSI escape code to reset text formatting

	errorMessage := "Please install Zed and ensure it's in your PATH.\n\n"
	moreInfo := []string{
		"More information: \n",
		"1) Install Zed by following: https://zed.dev/docs/getting-started or download from https://zed.dev/download\n",
		"2) To install the zed command line tool, select Zed > Install CLI from the application menu\n\n",
	}

	return "", errors.New(redBold + errorMessage + reset + strings.Join(moreInfo, ""))
}

func printDisclaimer() {
	views.RenderTip("Note: Zed remote development is not yet stable. Issues like opening terminals and using extensions are expected.")
}
