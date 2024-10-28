// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package detect

import (
	"os"
	"path"
	"path/filepath"

	"github.com/daytonaio/daytona/pkg/ssh"
	"github.com/daytonaio/daytona/pkg/target/workspace/buildconfig"
)

type BuilderType string

var (
	BuilderTypeDevcontainer BuilderType = "devcontainer"
	BuilderTypeImage        BuilderType = "image"
)

func DetectWorkspaceBuilderType(buildConfig *buildconfig.BuildConfig, workspaceDir string, sshClient *ssh.Client) (BuilderType, error) {
	if buildConfig == nil {
		return BuilderTypeImage, nil
	}

	if buildConfig.Devcontainer != nil {
		return BuilderTypeDevcontainer, nil
	}

	if sshClient != nil {
		if _, err := sshClient.ReadFile(path.Join(workspaceDir, ".devcontainer/devcontainer.json")); err == nil {
			buildConfig.Devcontainer = &buildconfig.DevcontainerConfig{
				FilePath: ".devcontainer/devcontainer.json",
			}
			return BuilderTypeDevcontainer, nil
		}
		if _, err := sshClient.ReadFile(path.Join(workspaceDir, ".devcontainer.json")); err == nil {
			buildConfig.Devcontainer = &buildconfig.DevcontainerConfig{
				FilePath: ".devcontainer.json",
			}
			return BuilderTypeDevcontainer, nil
		}
	} else {
		if devcontainerFilePath, pathError := findDevcontainerConfigFilePath(workspaceDir); pathError == nil {
			buildConfig.Devcontainer = &buildconfig.DevcontainerConfig{
				FilePath: devcontainerFilePath,
			}

			return BuilderTypeDevcontainer, nil
		}
	}

	return BuilderTypeImage, nil
}

func findDevcontainerConfigFilePath(workspaceDir string) (string, error) {
	devcontainerPath := ".devcontainer/devcontainer.json"
	isDevcontainer, err := fileExists(filepath.Join(workspaceDir, devcontainerPath))
	if !isDevcontainer || err != nil {
		devcontainerPath = ".devcontainer.json"
		isDevcontainer, err = fileExists(filepath.Join(workspaceDir, devcontainerPath))
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
