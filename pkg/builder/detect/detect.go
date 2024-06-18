// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package detect

import (
	"os"
	"path"
	"path/filepath"

	"github.com/daytonaio/daytona/pkg/ssh"
	"github.com/daytonaio/daytona/pkg/workspace"
)

type BuilderType string

var (
	BuilderTypeDevcontainer BuilderType = "devcontainer"
	BuilderTypeImage        BuilderType = "image"
)

func DetectProjectBuilderType(project *workspace.Project, projectDir string, sshClient *ssh.Client) (BuilderType, error) {
	if project.Build != nil && project.Build.Devcontainer != nil {
		return BuilderTypeDevcontainer, nil
	}

	if sshClient != nil {
		if _, err := sshClient.ReadFile(path.Join(projectDir, ".devcontainer/devcontainer.json")); err == nil {
			project.Build.Devcontainer = &workspace.ProjectBuildDevcontainer{
				DevContainerFilePath: ".devcontainer/devcontainer.json",
			}
			return BuilderTypeDevcontainer, nil
		}
		if _, err := sshClient.ReadFile(path.Join(projectDir, ".devcontainer.json")); err == nil {
			project.Build.Devcontainer = &workspace.ProjectBuildDevcontainer{
				DevContainerFilePath: ".devcontainer.json",
			}
			return BuilderTypeDevcontainer, nil
		}
	} else {
		if devcontainerFilePath, pathError := findDevcontainerConfigFilePath(projectDir); pathError == nil {
			project.Build.Devcontainer = &workspace.ProjectBuildDevcontainer{
				DevContainerFilePath: devcontainerFilePath,
			}

			return BuilderTypeDevcontainer, nil
		}
	}

	return BuilderTypeImage, nil
}

func findDevcontainerConfigFilePath(projectDir string) (string, error) {
	devcontainerPath := ".devcontainer/devcontainer.json"
	isDevcontainer, err := fileExists(filepath.Join(projectDir, devcontainerPath))
	if err != nil {
		devcontainerPath = ".devcontainer.json"
		isDevcontainer, err = fileExists(filepath.Join(projectDir, devcontainerPath))
		if err != nil {
			return devcontainerPath, nil
		}
	}

	if isDevcontainer {
		return devcontainerPath, nil
	}

	return "", os.ErrNotExist
}

func fileExists(filePath string) (bool, error) {
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		// There was an error checking for the file
		return false, err
	}
	return true, nil
}
