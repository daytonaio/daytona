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

func (d *DockerClient) Start(ctx context.Context, containerId string) error {
	d.cache.SetSandboxState(ctx, containerId, enums.SandboxStateStarting)

	c, err := d.apiClient.ContainerInspect(ctx, containerId)
	if err != nil {
		return err
	}

	if c.State.Running {
		d.cache.SetSandboxState(ctx, containerId, enums.SandboxStateStarted)
		return nil
	}

	err = d.apiClient.ContainerStart(ctx, containerId, container.StartOptions{})
	if err != nil {
		return err
	}

	// make sure container is running
	err = d.waitForContainerRunning(ctx, containerId, 10*time.Second)
	if err != nil {
		return err
	}

	d.cache.SetSandboxState(ctx, containerId, enums.SandboxStateStarted)

	processesCtx := context.Background()

	// Start Daytona daemon and terminal process
	if d.daytonaBinaryURL != "" {
		go func() {
			if err := d.startDaytonaDaemon(processesCtx, containerId); err != nil {
				log.Errorf("Failed to start Daytona daemon: %s\n", err.Error())
			}
		}()
	}

	if d.terminalBinaryURL != "" {
		go func() {
			if err := d.startTerminalProcess(processesCtx, containerId, 22222); err != nil {
				log.Errorf("Failed to start terminal process: %s\n", err.Error())
			}
		}()
	}

	return nil
}

func (d *DockerClient) waitForContainerRunning(ctx context.Context, containerId string, timeout time.Duration) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			return fmt.Errorf("timeout waiting for container %s to start", containerId)
		case <-ticker.C:
			c, err := d.apiClient.ContainerInspect(ctx, containerId)
			if err != nil {
				return err
			}

			if c.State.Running {
				return nil
			}
		}
	}
}
