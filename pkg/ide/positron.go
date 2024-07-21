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

func OpenPositronIDE(activeProfile config.Profile, workspaceId, projectName, projectProviderMetadata string) error {
	projectDir, err := util.GetProjectDir(activeProfile, workspaceId, projectName)
	if err != nil {
		return err
	}

	positronPath, err := findPositronPath()
	if err != nil {
		return err
	}

	command := fmt.Sprintf("%s %s", positronPath, filepath.ToSlash(projectDir))

	if err := runCommand(command); err != nil {
		return err
	}

	views.RenderInfoMessage(fmt.Sprintf("Opening project '%s' in Positron IDE", projectName))
	return nil
}

func findPositronPath() (string, error) {
	// Attempt to find Positron IDE in common installation directories
	paths := []string{
		"/usr/local/bin/positron",
		"/opt/positron/bin/positron",
		"/usr/bin/positron",
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	// Try to find Positron using a command-line search
	return findPositronInPath()
}

func findPositronInPath() (string, error) {
	cmd := exec.Command("which", "positron")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("could not locate Positron IDE: %v", err)
	}

	positronPath := strings.TrimSpace(string(output))
	if positronPath == "" {
		return "", fmt.Errorf("Positron IDE not found in system PATH")
	}

	return positronPath, nil
}

func runCommand(command string) error {
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
