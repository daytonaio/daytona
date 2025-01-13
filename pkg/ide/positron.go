// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ide

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/build/devcontainer"
	"github.com/daytonaio/daytona/pkg/views"
)

func OpenPositron(activeProfile config.Profile, workspaceId string, projectName string, projectProviderMetadata string, gpgkey string) error {
	path, err := GetPositronBinaryPath()
	if err != nil {
		return err
	}

	projectHostname := config.GetProjectHostname(activeProfile.Id, workspaceId, projectName)

	projectDir, err := util.GetProjectDir(activeProfile, workspaceId, projectName, gpgkey)
	if err != nil {
		return err
	}

	commandArgument := fmt.Sprintf("vscode-remote://ssh-remote+%s/%s", projectHostname, projectDir)
	if runtime.GOARCH == "arm64" {
		printPositronDisclaimer()
	}
	positronCommand := exec.Command(path, "--disable-extension", "ms-vscode-remote.remote-containers", "--folder-uri", commandArgument)

	err = positronCommand.Run()
	if err != nil {
		return err
	}

	if projectProviderMetadata == "" {
		return nil
	}

	return setupVSCodeCustomizations(projectHostname, projectProviderMetadata, devcontainer.Vscode, "*/.positron-server/*/bin/positron-server", "$HOME/.positron-server/data/Machine/settings.json", ".daytona-customizations-lock-positron")
}

func GetPositronBinaryPath() (string, error) {
	path, err := exec.LookPath("positron")
	if err == nil {
		return path, err
	}

	redBold := "\033[1;31m" // ANSI escape code for red and bold
	reset := "\033[0m"      // ANSI escape code to reset text formatting

	errorMessage := "Please install Positron from https://positron.posit.co/download.html and ensure it's in your PATH.\n"

	return "", errors.New(redBold + errorMessage + reset)
}

func printPositronDisclaimer() {
	views.RenderTip(`
Note: Positron does not currently support Linux ARM64 builds, including remote environments and SSH sessions.
Refer to https://github.com/posit-dev/positron/issues/5911 for updates.`)
}
