// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package docker

import (
	"fmt"
	"path"
	"strings"

	"github.com/docker/docker/api/types/mount"
)

func (d *DockerClient) startDevcontainerProject(opts *CreateProjectOptions) (RemoteUser, error) {
	go func() {
		err := d.runDevcontainerUserCommands(opts)
		if err != nil {
			opts.LogWriter.Write([]byte(fmt.Sprintf("Error running devcontainer user commands: %s\n", err)))
		}
	}()

	_, remoteUser, err := d.CreateFromDevcontainer(d.toCreateDevcontainerOptions(opts, false))
	return remoteUser, err
}

func (d *DockerClient) runDevcontainerUserCommands(opts *CreateProjectOptions) error {
	socketForwardId, err := d.ensureDockerSockForward(opts.LogWriter)
	if err != nil {
		return err
	}

	opts.LogWriter.Write([]byte("Running devcontainer user commands...\n"))

	paths := d.getDevcontainerPaths(opts.ProjectDir, opts.Project.BuildConfig.Devcontainer.FilePath)

	devcontainerCmd := []string{
		"devcontainer",
		"run-user-commands",
		"--workspace-folder=" + paths.ProjectTarget,
		"--config=" + paths.TargetConfigFilePath,
		"--override-config=" + path.Join(paths.OverridesTarget, "devcontainer.json"),
		"--id-label=daytona.workspace.id=" + opts.Project.WorkspaceId,
		"--id-label=daytona.project.name=" + opts.Project.Name,
	}

	cmd := strings.Join(devcontainerCmd, " ")

	createDevcontainerOptions := d.toCreateDevcontainerOptions(opts, true)

	_, err = d.execDevcontainerCommand(cmd, &createDevcontainerOptions, paths, paths.ProjectTarget, socketForwardId, true, []mount.Mount{
		{
			Type:   mount.TypeBind,
			Source: paths.OverridesDir,
			Target: paths.OverridesTarget,
		},
	})

	return err
}
