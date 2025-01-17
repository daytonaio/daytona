// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package docker

import (
	"fmt"
	"path"
	"strings"

	"github.com/docker/docker/api/types/mount"
)

func (d *DockerClient) startDevcontainerWorkspace(opts *CreateWorkspaceOptions) (RemoteUser, error) {
	go func() {
		err := d.runDevcontainerUserCommands(opts)
		if err != nil {
			opts.LogWriter.Write([]byte(fmt.Sprintf("Error running devcontainer user commands: %s\n", err)))
		}
	}()

	_, remoteUser, err := d.CreateFromDevcontainer(d.toCreateDevcontainerOptions(opts, false))
	return remoteUser, err
}

func (d *DockerClient) runDevcontainerUserCommands(opts *CreateWorkspaceOptions) error {
	cr := opts.ContainerRegistries.FindContainerRegistryByImageName(opts.BuilderImage)
	socketForwardId, err := d.ensureDockerSockForward(opts.BuilderImage, cr, opts.LogWriter)
	if err != nil {
		return err
	}

	opts.LogWriter.Write([]byte("Running devcontainer user commands...\n"))

	paths := d.getDevcontainerPaths(opts.WorkspaceDir, opts.Workspace.BuildConfig.Devcontainer.FilePath)

	devcontainerCmd := []string{
		"devcontainer",
		"run-user-commands",
		"--workspace-folder=" + paths.WorkspaceDirTarget,
		"--config=" + paths.TargetConfigFilePath,
		"--override-config=" + path.Join(paths.OverridesTarget, "devcontainer.json"),
		"--id-label=daytona.target.id=" + opts.Workspace.TargetId,
		"--id-label=daytona.workspace.id=" + opts.Workspace.Id,
	}

	cmd := strings.Join(devcontainerCmd, " ")

	createDevcontainerOptions := d.toCreateDevcontainerOptions(opts, true)

	_, err = d.execDevcontainerCommand(cmd, &createDevcontainerOptions, paths, paths.WorkspaceDirTarget, socketForwardId, true, []mount.Mount{
		{
			Type:   mount.TypeBind,
			Source: paths.OverridesDir,
			Target: paths.OverridesTarget,
		},
	})

	return err
}
