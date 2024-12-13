// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package docker

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/ports"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/go-connections/nat"
	log "github.com/sirupsen/logrus"
)

// pulledImages map keeps track of pulled images for workspace creation in order to avoid pulling the same image multiple times
// This is only an optimisation for images with tag 'latest'
func (d *DockerClient) createWorkspaceFromImage(opts *CreateWorkspaceOptions, pulledImages map[string]bool, mountWorkspaceDir bool) error {
	if pulledImages[opts.Workspace.Image] {
		return d.initWorkspaceContainer(opts, mountWorkspaceDir)
	}

	cr := opts.ContainerRegistries.FindContainerRegistryByImageName(opts.Workspace.Image)
	err := d.PullImage(opts.Workspace.Image, cr, opts.LogWriter)
	if err != nil {
		return err
	}
	pulledImages[opts.Workspace.Image] = true

	return d.initWorkspaceContainer(opts, mountWorkspaceDir)
}

func (d *DockerClient) initWorkspaceContainer(opts *CreateWorkspaceOptions, mountWorkspaceDir bool) error {
	ctx := context.Background()

	mounts := []mount.Mount{}
	if mountWorkspaceDir {
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: opts.WorkspaceDir,
			Target: fmt.Sprintf("/home/%s/%s", opts.Workspace.User, opts.Workspace.WorkspaceFolderName()),
		})
	}

	var availablePort *uint16
	var portBindings map[nat.Port][]nat.PortBinding

	if opts.Workspace.TargetId == "local" {
		p, err := ports.GetAvailableEphemeralPort()
		if err != nil {
			log.Error(err)
		} else {
			availablePort = &p
			portBindings = make(map[nat.Port][]nat.PortBinding)
			portBindings["2280/tcp"] = []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: fmt.Sprintf("%d", *availablePort),
				},
			}
		}
	}

	c, err := d.apiClient.ContainerCreate(ctx, GetContainerCreateConfig(opts.Workspace, availablePort), &container.HostConfig{
		Privileged: true,
		Mounts:     mounts,
		ExtraHosts: []string{
			"host.docker.internal:host-gateway",
		},
		PortBindings: portBindings,
	}, nil, nil, d.GetWorkspaceContainerName(opts.Workspace))
	if err != nil {
		return err
	}

	err = d.apiClient.ContainerStart(ctx, c.ID, container.StartOptions{})
	if err != nil {
		return err
	}

	go func() {
		for {
			err = d.GetContainerLogs(c.ID, opts.LogWriter)
			if err == nil {
				break
			}
			log.Error(err)
			time.Sleep(100 * time.Millisecond)
		}
	}()

	if runtime.GOOS != "windows" && mountWorkspaceDir {
		_, err = d.updateContainerUserUidGid(c.ID, opts)
	}

	err = d.apiClient.ContainerStop(ctx, c.ID, container.StopOptions{
		Signal: "SIGKILL",
	})
	if err != nil {
		return err
	}

	return nil
}

func GetContainerCreateConfig(workspace *models.Workspace, toolboxApiHostPort *uint16) *container.Config {
	envVars := []string{}

	for key, value := range workspace.EnvVars {
		envVars = append(envVars, fmt.Sprintf("%s=%s", key, value))
	}

	labels := map[string]string{
		"daytona.target.id":                workspace.TargetId,
		"daytona.workspace.id":             workspace.Id,
		"daytona.workspace.repository.url": workspace.Repository.Url,
	}

	if toolboxApiHostPort != nil {
		labels["daytona.toolbox.api.hostPort"] = fmt.Sprintf("%d", *toolboxApiHostPort)
	}

	exposedPorts := nat.PortSet{}
	if toolboxApiHostPort != nil {
		exposedPorts["2280/tcp"] = struct{}{}
	}

	return &container.Config{
		Hostname:     workspace.Id,
		Image:        workspace.Image,
		Labels:       labels,
		User:         workspace.User,
		Env:          envVars,
		Entrypoint:   []string{"sleep", "infinity"},
		AttachStdout: true,
		AttachStderr: true,
		ExposedPorts: exposedPorts,
	}
}
