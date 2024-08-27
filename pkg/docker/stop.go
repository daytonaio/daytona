// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package docker

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/daytonaio/daytona/pkg/workspace/project"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
)

func (d *DockerClient) StopProject(p *project.Project, logWriter io.Writer) error {
	return d.stopProjectContainer(p, logWriter)
}

func (d *DockerClient) stopProjectContainer(p *project.Project, logWriter io.Writer) error {
	containerName := d.GetProjectContainerName(p)
	ctx := context.Background()

	err := d.apiClient.ContainerStop(ctx, containerName, container.StopOptions{})
	if err != nil {
		return err
	}

	var c types.ContainerJSON

	//	TODO: timeout
	for {
		c, err = d.apiClient.ContainerInspect(ctx, containerName)
		if err != nil {
			return err
		}

		if !c.State.Running {
			break
		}

		time.Sleep(1 * time.Second)
	}

	// TODO: Add logging
	_, composeContainers, err := d.getComposeContainers(c)
	if err != nil {
		return err
	}

	if composeContainers == nil {
		return nil
	}

	if logWriter != nil {
		logWriter.Write([]byte("Stopping compose containers\n"))
	}

	for _, c := range composeContainers {
		err = d.apiClient.ContainerStop(ctx, c.ID, container.StopOptions{})
		if err != nil {
			return err
		}
		if logWriter != nil {
			logWriter.Write([]byte(fmt.Sprintf("Stopped %s\n", strings.TrimPrefix(c.Names[0], "/"))))
		}
	}

	return nil
}
