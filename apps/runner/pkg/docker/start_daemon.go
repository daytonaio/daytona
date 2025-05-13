// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/daytonaio/runner/pkg/common"
	"github.com/docker/docker/api/types/container"

	log "github.com/sirupsen/logrus"
)

func (d *DockerClient) StartDaytonaDaemon(ctx context.Context, containerName string) error {
	execOptions := container.ExecOptions{
		Cmd:          []string{"sh", "-c", "daytona"},
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
	}

	execStartOptions := container.ExecStartOptions{
		Detach: false,
	}

	result, err := d.execSync(ctx, containerName, execOptions, execStartOptions)
	if err != nil {
		log.Errorf("Error starting Daytona daemon: %s", err.Error())
		return nil
	}

	if result.ExitCode != 0 && result.StdErr != "" {
		log.Errorf("Error starting Daytona daemon: %s", string(result.StdErr))
		return nil
	}

	return nil
}

func (d *DockerClient) DaemonStartedCheck(ctx context.Context, targetURL string, numRetries int, timeout, retryInterval time.Duration) error {
	// wait for the daemon to start listening on port 2280
	target, err := url.Parse(targetURL)
	if err != nil {
		return common.NewBadRequestError(fmt.Errorf("failed to parse target URL: %w", err))
	}

	for i := 0; i < numRetries; i++ {
		conn, err := net.DialTimeout("tcp", target.Host, timeout)
		if err != nil {
			time.Sleep(retryInterval)
			continue
		}
		conn.Close()
		break
	}

	return nil
}

func (d *DockerClient) GetContainerTargetURL(ctx context.Context, containerName string) (string, error) {
	container, err := d.ContainerInspect(ctx, containerName)
	if err != nil {
		return "", common.NewNotFoundError(fmt.Errorf("sandbox container not found: %w", err))
	}

	for _, network := range container.NetworkSettings.Networks {
		return fmt.Sprintf("http://%s:2280", network.IPAddress), nil
	}

	return "", errors.New("container has no IP address, it might not be running")
}
