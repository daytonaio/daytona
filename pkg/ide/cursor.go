// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ide

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/views"
)

func OpenCursorIDE(activeProfile config.Profile, workspaceId string, projectName string, projectProviderMetadata string) error {
	checkAndAlertVSCodeInstalled()

	projectDir, err := util.GetProjectDir(activeProfile, workspaceId, projectName)
	if err != nil {
		return err
	}

	cursorPath, err := findPositronPath()
	if err != nil {
		return err
	}

	command := fmt.Sprintf("%s %s", cursorPath, filepath.ToSlash(projectDir))

	if err := runCommand(command); err != nil {
		return err
	}

	views.RenderInfoMessage(fmt.Sprintf("Opening project '%s' in Cursor IDE", projectName))
	return nil
}

func findCursorPath() (string, error) {
	paths := []string{
		"/usr/local/bin/cursor",
		"/opt/cursor/bin/cursor",
		"/usr/bin/cursor",
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return findCursorInPath()
}

func findCursorInPath() (string, error) {
	cmd := exec.Command("which", "cursor")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("could not locate Cursor IDE: %v", err)
	}

	cursorPath := strings.TrimSpace(string(output))
	if cursorPath == "" {
		return "", fmt.Errorf("Cursor IDE not found in system PATH")
	}

	return cursorPath, nil
}

func runCommand(command string) error {
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
