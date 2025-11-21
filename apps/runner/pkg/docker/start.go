// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/daytonaio/common-go/pkg/timer"
	"github.com/daytonaio/runner/pkg/models/enums"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"

	log "github.com/sirupsen/logrus"
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

	if !d.useDaemonEntrypoint {
		processesCtx := context.Background()
		go func() {
			if err := d.startDaytonaDaemon(processesCtx, containerId, c.Config.WorkingDir); err != nil {
				log.Errorf("Failed to start Daytona daemon: %s\n", err.Error())
			}
		}()
	}

	// If daemonEntrypoint is enabled, daemon is started as part of the container entrypoint; otherwise, it's started separately above. In either case, we wait for it here.
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
				log.Errorf("Failed to set network limiter: %v", err)
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
			return errors.New("timeout waiting for the sandbox to start - please ensure that your entrypoint is long-running")
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

func getContainerIP(container *types.ContainerJSON) (string, error) {
	for _, network := range container.NetworkSettings.Networks {
		return network.IPAddress, nil
	}
	return "", fmt.Errorf("no IP address found. Is the Sandbox started?")
}
