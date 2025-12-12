/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

package executor

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types/container"

	apiclient "github.com/daytonaio/apiclient"
)

func (e *Executor) startSandbox(ctx context.Context, job *apiclient.Job) error {
	sandboxId := job.GetResourceId()
	e.log.Debug("starting sandbox", "job_id", job.GetId(), "sandbox_id", sandboxId)

	// Check if container is already running
	containerInfo, err := e.dockerClient.ContainerInspect(ctx, sandboxId)
	if err != nil {
		return fmt.Errorf("inspect container: %w", err)
	}

	if !containerInfo.State.Running {
		// Start the container
		if err := e.dockerClient.ContainerStart(ctx, sandboxId, container.StartOptions{}); err != nil {
			return fmt.Errorf("start container: %w", err)
		}
		e.log.Debug("container started", "sandbox_id", sandboxId)

		// Re-inspect to get updated network info and entrypoint
		containerInfo, err = e.dockerClient.ContainerInspect(ctx, sandboxId)
		if err != nil {
			return fmt.Errorf("inspect container after start: %w", err)
		}
	} else {
		e.log.Debug("container already running", "sandbox_id", sandboxId)
	}

	// Get container IP for daemon health check
	containerIP := ""
	for _, network := range containerInfo.NetworkSettings.Networks {
		containerIP = network.IPAddress
		break
	}

	if containerIP == "" {
		return fmt.Errorf("no IP address found for container")
	}

	// Wait for daemon to be ready
	e.log.Debug("waiting for daemon to be ready", "sandbox_id", sandboxId, "ip", containerIP)
	if err := e.waitForDaemonRunning(ctx, containerIP, 10*time.Second); err != nil {
		e.log.Error("daemon failed to start", "error", err)
		return fmt.Errorf("daemon not ready: %w", err)
	}

	e.log.Debug("daemon is ready", "sandbox_id", sandboxId)
	e.log.Info("sandbox started successfully", "sandbox_id", sandboxId)
	return nil
}
