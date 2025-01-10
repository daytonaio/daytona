// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ide

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/views"
)

func OpenVim(activeProfile config.Profile, workspaceId, projectName, gpgKey string) error {
	path, err := GetVimBinaryPath()
	if err != nil {
		return err
	}

	printVimDisclaimer()
	projectHostname := config.GetProjectHostname(activeProfile.Id, workspaceId, projectName)
	projectDir, err := util.GetProjectDir(activeProfile, workspaceId, projectName, gpgKey)
	if err != nil {
		return err
	}
	// This suppresses the "Permanently added ... to the list of known hosts" message and reduce file listing errors
	vimConfig := fmt.Sprintf(`:let g:netrw_list_cmd = "ssh -o LogLevel=QUIET %s ls -Fa"`, projectHostname)

	// Both scp and sftp are supported by Vim's built-in netrw plugin
	vimCmd := exec.Command(path, "-c", vimConfig, fmt.Sprintf("scp://%s/%s/", projectHostname, projectDir))
	vimCmd.Stdin = os.Stdin
	vimCmd.Stdout = os.Stdout
	vimCmd.Stderr = os.Stderr

	return vimCmd.Run()
}

func GetVimBinaryPath() (string, error) {
	path, err := exec.LookPath("vim")
	if err == nil {
		return path, err
	}

	redBold := "\033[1;31m" // ANSI escape code for red and bold
	reset := "\033[0m"      // ANSI escape code to reset text formatting

	errorMessage := "Please install Vim and ensure it's in your PATH.\n"

	return "", errors.New(redBold + errorMessage + reset)
}

func printVimDisclaimer() {
	views.RenderTip(`Note: Vim only allows you to edit remote files using Netrw.
For a better experience, consider using a dedicated IDE.
If you need to run commands, please use the 'daytona ssh' command instead.`)
}
