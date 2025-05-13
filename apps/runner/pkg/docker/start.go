// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"
	"time"

	"github.com/daytonaio/runner/pkg/models/enums"
	"github.com/docker/docker/api/types/container"

	log "github.com/sirupsen/logrus"
)

func (d *DockerClient) Start(ctx context.Context, containerName string) error {
	d.cache.SetSandboxState(ctx, containerName, enums.SandboxStateStarting)

	c, err := d.ContainerInspect(ctx, containerName)
	if err != nil {
		return err
	}

	if c.State.Running {
		d.cache.SetSandboxState(ctx, containerName, enums.SandboxStateStarted)
		return nil
	}

	err = d.apiClient.ContainerStart(ctx, containerName, container.StartOptions{})
	if err != nil {
		return err
	}

	// make sure container is running
	err = d.waitForContainerRunning(ctx, containerName, 10*time.Second)
	if err != nil {
		return err
	}

	d.cache.SetSandboxState(ctx, containerName, enums.SandboxStateStarted)

	go func() {
		if err := d.StartDaytonaDaemon(context.Background(), containerName); err != nil {
			log.Errorf("Failed to start Daytona daemon: %s\n", err.Error())
		}
	}()

	targetURL, err := d.GetContainerTargetURL(ctx, containerName)
	if err != nil {
		return err
	}

	err = d.DaemonStartedCheck(ctx, targetURL, 10, 1*time.Second, 50*time.Millisecond)
	if err != nil {
		return err
	}

	d.cache.SetSandboxState(ctx, containerName, enums.SandboxStateStarted)

	return nil
}

func (d *DockerClient) waitForContainerRunning(ctx context.Context, containerName string, timeout time.Duration) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			return fmt.Errorf("timeout waiting for container %s to start", containerName)
		case <-ticker.C:
			c, err := d.ContainerInspect(ctx, containerName)
			if err != nil {
				return err
			}

			if c.State.Running {
				return nil
			}
		}
	}
}
