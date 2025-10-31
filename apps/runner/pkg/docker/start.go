// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"time"

	"github.com/daytonaio/common-go/pkg/timer"
	"github.com/daytonaio/runner/pkg/models/enums"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
)

func (d *DockerClient) Start(ctx context.Context, containerId string, metadata map[string]string) error {
	defer timer.Timer()()
	d.statesCache.SetSandboxState(ctx, containerId, enums.SandboxStateStarting)

	// Cancel a backup if it's already in progress
	backup_context, ok := backup_context_map.Get(containerId)
	if ok {
		backup_context.cancel()
	}

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

		d.statesCache.SetSandboxState(ctx, containerId, enums.SandboxStateStarted)
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
		if err := d.startDaytonaDaemon(processesCtx, containerId, c.Config.WorkingDir); err != nil {
			slog.ErrorContext(ctx, "Failed to start Daytona daemon", "error", err)
		}
	}()

	err = d.waitForDaemonRunning(ctx, containerIP, 10*time.Second)
	if err != nil {
		return err
	}

	d.statesCache.SetSandboxState(ctx, containerId, enums.SandboxStateStarted)

	if metadata["limitNetworkEgress"] == "true" {
		go func() {
			containerShortId := c.ID[:12]
			err = d.netRulesManager.SetNetworkLimiter(containerShortId, containerIP)
			if err != nil {
				slog.ErrorContext(ctx, "Failed to set network limiter", "error", err)
			}
		}()
	}

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
		return common_errors.NewBadRequestError(fmt.Errorf("failed to parse target URL: %w", err))
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
