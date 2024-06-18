// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package docker

import (
	"context"
	"io"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
)

func (d *DockerClient) GetContainerLogs(containerName string, logWriter io.Writer) error {
	if logWriter == nil {
		return nil
	}

	inspect, err := d.apiClient.ContainerInspect(context.Background(), containerName)
	if err != nil {
		return err
	}

	logs, err := d.apiClient.ContainerLogs(context.Background(), containerName, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	})
	if err != nil {
		return err
	}
	defer logs.Close()

	if inspect.Config.Tty {
		_, err = io.Copy(logWriter, logs)
		return err
	}

	_, err = stdcopy.StdCopy(logWriter, logWriter, logs)

	return err
}
