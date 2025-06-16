// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/daytonaio/runner/pkg/common"
	"github.com/daytonaio/runner/pkg/models/enums"
	"github.com/docker/docker/api/types/container"

	log "github.com/sirupsen/logrus"
)

func (d *DockerClient) Start(ctx context.Context, containerId string) error {
	d.cache.SetSandboxState(ctx, containerId, enums.SandboxStateStarting)

	c, err := d.ContainerInspect(ctx, containerId)
	if err != nil {
		return err
	}

	var containerIP string
	for _, network := range c.NetworkSettings.Networks {
		containerIP = network.IPAddress
		break
	}

	if c.State.Running {
		err = d.waitForDaemonRunning(ctx, containerIP, 10*time.Second)
		if err != nil {
			return err
		}

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

	processesCtx := context.Background()

	go func() {
		if err := d.startDaytonaDaemon(processesCtx, containerId); err != nil {
			log.Errorf("Failed to start Daytona daemon: %s\n", err.Error())
		}
	}()

	err = d.waitForDaemonRunning(ctx, containerIP, 10*time.Second)
	if err != nil {
		return err
	}

	d.cache.SetSandboxState(ctx, containerId, enums.SandboxStateStarted)

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
			c, err := d.ContainerInspect(ctx, containerId)
			if err != nil {
				return err
			}

			if c.State.Running {
				return nil
			}
		}
	}
}

func (d *DockerClient) waitForDaemonRunning(ctx context.Context, containerIP string, timeout time.Duration) error {
	// Build the target URL
	targetURL := fmt.Sprintf("http://%s:2280", containerIP)
	target, err := url.Parse(targetURL)
	if err != nil {
		return common.NewBadRequestError(fmt.Errorf("failed to parse target URL: %w", err))
	}

	for i := 0; i < 10; i++ {
		conn, err := net.DialTimeout("tcp", target.Host, 1*time.Second)
		if err != nil {
			time.Sleep(50 * time.Millisecond)
			continue
		}
		conn.Close()
		break
	}

	return nil
}
