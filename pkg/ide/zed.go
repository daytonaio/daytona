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
)

func OpenZed(activeProfile config.Profile, workspaceId, projectName string) error {
	path, err := GetZedBinaryPath()
	if err != nil {
		return err
	}

	projectHostname := config.GetProjectHostname(activeProfile.Id, workspaceId, projectName)
	projectDir, err := util.GetProjectDir(activeProfile, workspaceId, projectName)
	if err != nil {
		return err
	}

	zedCmd := exec.Command(path, fmt.Sprintf("ssh://%s%s", projectHostname, projectDir))

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
		"Note: Zed SSH connections does not support most features yet! You cannot use project search, language servers, or basically do anything except edit files.",
	}

	return "", errors.New(redBold + errorMessage + reset + strings.Join(moreInfo, ""))
}
