// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package docker

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/daytonaio/daytona/pkg/ports"
	"github.com/daytonaio/daytona/pkg/target/project"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/go-connections/nat"
	log "github.com/sirupsen/logrus"
)

// pulledImages map keeps track of pulled images for project creation in order to avoid pulling the same image multiple times
// This is only an optimisation for images with tag 'latest'
func (d *DockerClient) createProjectFromImage(opts *CreateProjectOptions, pulledImages map[string]bool, mountProjectDir bool) error {
	if pulledImages[opts.Project.Image] {
		return d.initProjectContainer(opts, mountProjectDir)
	}

	err := d.PullImage(opts.Project.Image, opts.ContainerRegistry, opts.LogWriter)
	if err != nil {
		return err
	}
	pulledImages[opts.Project.Image] = true

	return d.initProjectContainer(opts, mountProjectDir)
}

func (d *DockerClient) initProjectContainer(opts *CreateProjectOptions, mountProjectDir bool) error {
	ctx := context.Background()

	mounts := []mount.Mount{}
	if mountProjectDir {
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: opts.ProjectDir,
			Target: fmt.Sprintf("/home/%s/%s", opts.Project.User, opts.Project.Name),
		})
	}

	var availablePort *uint16
	var portBindings map[nat.Port][]nat.PortBinding

	if opts.Project.TargetConfig == "local" {
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

	c, err := d.apiClient.ContainerCreate(ctx, GetContainerCreateConfig(opts.Project, availablePort), &container.HostConfig{
		Privileged: true,
		Mounts:     mounts,
		ExtraHosts: []string{
			"host.docker.internal:host-gateway",
		},
		PortBindings: portBindings,
	}, nil, nil, d.GetProjectContainerName(opts.Project))
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

	if runtime.GOOS != "windows" && mountProjectDir {
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

func GetContainerCreateConfig(project *project.Project, toolboxApiHostPort *uint16) *container.Config {
	envVars := []string{}

	for key, value := range project.EnvVars {
		envVars = append(envVars, fmt.Sprintf("%s=%s", key, value))
	}

	labels := map[string]string{
		"daytona.target.id":              project.TargetId,
		"daytona.project.name":           project.Name,
		"daytona.project.repository.url": project.Repository.Url,
	}

	if toolboxApiHostPort != nil {
		labels["daytona.toolbox.api.hostPort"] = fmt.Sprintf("%d", *toolboxApiHostPort)
	}

	exposedPorts := nat.PortSet{}
	if toolboxApiHostPort != nil {
		exposedPorts["2280/tcp"] = struct{}{}
	}

	return &container.Config{
		Hostname:     project.Name,
		Image:        project.Image,
		Labels:       labels,
		User:         project.User,
		Env:          envVars,
		Entrypoint:   []string{"sleep", "infinity"},
		AttachStdout: true,
		AttachStderr: true,
		ExposedPorts: exposedPorts,
	}
}
