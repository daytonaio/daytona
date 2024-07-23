// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package detect

import (
	"os"
	"path"
	"path/filepath"

	"github.com/daytonaio/daytona/pkg/ssh"
	"github.com/daytonaio/daytona/pkg/workspace/project"
	"github.com/daytonaio/daytona/pkg/workspace/project/build"
)

type BuilderType string

var (
	BuilderTypeDevcontainer BuilderType = "devcontainer"
	BuilderTypeImage        BuilderType = "image"
)

func DetectProjectBuilderType(project *project.Project, projectDir string, sshClient *ssh.Client) (BuilderType, error) {
	if project.Build != nil && project.Build.Devcontainer != nil {
		return BuilderTypeDevcontainer, nil
	}

	if sshClient != nil {
		if _, err := sshClient.ReadFile(path.Join(projectDir, ".devcontainer/devcontainer.json")); err == nil {
			project.Build.Devcontainer = &build.DevcontainerConfig{
				FilePath: ".devcontainer/devcontainer.json",
			}
			return BuilderTypeDevcontainer, nil
		}
		if _, err := sshClient.ReadFile(path.Join(projectDir, ".devcontainer.json")); err == nil {
			project.Build.Devcontainer = &build.DevcontainerConfig{
				FilePath: ".devcontainer.json",
			}
			return BuilderTypeDevcontainer, nil
		}
	} else {
		if devcontainerFilePath, pathError := findDevcontainerConfigFilePath(projectDir); pathError == nil {
			project.Build.Devcontainer = &build.DevcontainerConfig{
				FilePath: devcontainerFilePath,
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
