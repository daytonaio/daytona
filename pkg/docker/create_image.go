// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package docker

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	log "github.com/sirupsen/logrus"
)

// pulledImages map keeps track of pulled images for workspace creation in order to avoid pulling the same image multiple times
// This is only an optimisation for images with tag 'latest'
func (d *DockerClient) createWorkspaceFromImage(opts *CreateWorkspaceOptions, pulledImages map[string]bool) error {
	if pulledImages[opts.Workspace.Image] {
		return d.initWorkspaceContainer(opts)
	}

	err := d.PullImage(opts.Workspace.Image, opts.Cr, opts.LogWriter)
	if err != nil {
		return err
	}
	pulledImages[opts.Workspace.Image] = true

	return d.initWorkspaceContainer(opts)
}

func (d *DockerClient) initWorkspaceContainer(opts *CreateWorkspaceOptions) error {
	ctx := context.Background()

	c, err := d.apiClient.ContainerCreate(ctx, GetContainerCreateConfig(opts.Workspace), &container.HostConfig{
		Privileged: true,
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: opts.WorkspaceDir,
				Target: fmt.Sprintf("/home/%s/%s", opts.Workspace.User, opts.Workspace.WorkspaceFolderName()),
			},
		},
		ExtraHosts: []string{
			"host.docker.internal:host-gateway",
		},
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

	if runtime.GOOS != "windows" {
		_, err = d.updateContainerUserUidGid(c.ID, opts)
	}

	err = d.apiClient.ContainerStop(ctx, c.ID, container.StopOptions{})
	if err != nil {
		return err
	}

	return nil
}

func GetContainerCreateConfig(workspace *models.Workspace) *container.Config {
	envVars := []string{}

	for key, value := range workspace.EnvVars {
		envVars = append(envVars, fmt.Sprintf("%s=%s", key, value))
	}

	return &container.Config{
		Hostname: workspace.Id,
		Image:    workspace.Image,
		Labels: map[string]string{
			"daytona.target.id":                workspace.TargetId,
			"daytona.workspace.id":             workspace.Id,
			"daytona.workspace.repository.url": workspace.Repository.Url,
		},
		User:         workspace.User,
		Env:          envVars,
		Entrypoint:   []string{"sleep", "infinity"},
		AttachStdout: true,
		AttachStderr: true,
	}
}
