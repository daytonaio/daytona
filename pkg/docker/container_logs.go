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

	logs, err := d.apiClient.ContainerLogs(context.Background(), containerName, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	})
	if err != nil {
		return err
	}
	defer logs.Close()

	_, err = stdcopy.StdCopy(logWriter, logWriter, logs)

	return err
}
