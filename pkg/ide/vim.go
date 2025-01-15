// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ide

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/views"
)

// OpenEditor opens the specified editor (vim/nvim) for a given project
func OpenEditor(editorBinary string, activeProfile config.Profile, workspaceId, projectName, gpgKey string) error {
	path, err := GetEditorBinaryPath(editorBinary)
	if err != nil {
		return err
	}

	printEditorDisclaimer(editorBinary)
	projectHostname := config.GetProjectHostname(activeProfile.Id, workspaceId, projectName)
	projectDir, err := util.GetProjectDir(activeProfile, workspaceId, projectName, gpgKey)
	if err != nil {
		return err
	}
	// Suppresses "Permanently added ... to the list of known hosts" message and reduces file listing errors
	editorConfig := fmt.Sprintf(`:let g:netrw_list_cmd = "ssh -o LogLevel=QUIET %s ls -Fa"`, projectHostname)

	// Both scp and sftp are supported by Vim/Nvim's built-in netrw plugin
	editorCmd := exec.Command(path, "-c", editorConfig, fmt.Sprintf("scp://%s/%s/", projectHostname, projectDir))
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr

	return editorCmd.Run()
}

// GetEditorBinaryPath returns the path to the specified editor binary
func GetEditorBinaryPath(editorBinary string) (string, error) {
	path, err := exec.LookPath(strings.ToLower(editorBinary))
	if err == nil {
		return path, err
	}

	redBold := "\033[1;31m" // ANSI escape code for red and bold
	reset := "\033[0m"      // ANSI escape code to reset text formatting

	errorMessage := fmt.Sprintf("Please install %s and ensure it's in your PATH.\n", editorBinary)

	return "", errors.New(redBold + errorMessage + reset)
}

// printEditorDisclaimer displays a disclaimer message for the specified editor
func printEditorDisclaimer(editorName string) {
	views.RenderTip(fmt.Sprintf(`Note: %s only allows you to edit remote files using Netrw.
For a better experience, consider using a dedicated IDE.
If you need to run commands, please use the 'daytona ssh' command instead.`, editorName))
}
