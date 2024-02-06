// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"dagent/internal/util"
	"errors"

	"path"

	"github.com/docker/docker/api/types"
)

type DevContainerResponse struct {
	Outcome               string `json:"outcome"`
	ContainerId           string `json:"containerId"`
	ComposeProjectName    string `json:"composeProjectName"`
	RemoteUser            string `json:"remoteUser"`
	RemoteWorkspaceFolder string `json:"remoteWorkspaceFolder"`
}

func (project Project) isDevcontainer() bool {
	devcontainerPath := path.Join(project.GetPath(), ".devcontainer")
	devcontainerConfigPath := path.Join(devcontainerPath, "devcontainer.json")

	return util.FileExists(devcontainerConfigPath) // Update the function call to use the fully qualified name
}

func (project Project) initDevcontainerProject() error {
	workspacePath := "/workspace"

	execConfig := types.ExecConfig{
		Tty:          true,
		AttachStdout: true,
		AttachStderr: true,
		Cmd: []string{
			"devcontainer",
			"up",
			"--mount",
			"type=bind,source=/setup,target=/setup",
			"--workspace-folder=" + workspacePath,
		},
	}
	execResult, err := util.DockerExec(project.GetContainerName(), execConfig, nil)
	if err != nil {
		return err
	}

	if execResult.ExitCode != 0 {
		return errors.New("failed to initialize devcontainer: " + execResult.StdErr)
	}

	return nil
}
