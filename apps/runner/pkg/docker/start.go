// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/daytonaio/common-go/pkg/timer"
	"github.com/daytonaio/runner/pkg/common"
	"github.com/daytonaio/runner/pkg/models/enums"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"

	log "github.com/sirupsen/logrus"
)

func (d *DockerClient) Start(ctx context.Context, containerId string) error {
	defer timer.Timer()()
	d.cache.SetSandboxState(ctx, containerId, enums.SandboxStateStarting)

	c, err := d.ContainerInspect(ctx, containerId)
	if err != nil {
		return err
	}

	if c.State.Running {
		containerIP, err := getContainerIP(&c)
		if err != nil {
			return err
		}

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

	c, err = d.ContainerInspect(ctx, containerId)
	if err != nil {
		return err
	}

	containerIP, err := getContainerIP(&c)
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
	defer timer.Timer()()

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
	defer timer.Timer()()

	// Build the target URL
	targetURL := fmt.Sprintf("http://%s:2280/version", containerIP)
	target, err := url.Parse(targetURL)
	if err != nil {
		return common.NewBadRequestError(fmt.Errorf("failed to parse target URL: %w", err))
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		select {
		case <-timeoutCtx.Done():
			return fmt.Errorf("timeout waiting for daemon to start")
		default:
			conn, err := net.DialTimeout("tcp", target.Host, 1*time.Second)
			if err != nil {
				time.Sleep(5 * time.Millisecond)
				continue
			}
			conn.Close()
			return nil
		}
	}
}

func getContainerIP(container *types.ContainerJSON) (string, error) {
	for _, network := range container.NetworkSettings.Networks {
		return network.IPAddress, nil
	}
	return "", fmt.Errorf("no IP address found. Is the Sandbox started?")
}
