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

func (d *DockerClient) Stop(ctx context.Context, containerId string) error {
	d.cache.SetSandboxState(ctx, containerId, enums.SandboxStateStopping)

	// Cancel a backup if it's already in progress
	backup_context, ok := backup_context_map.Get(containerId)
	if ok {
		backup_context.cancel()
	}

	// Exponential backoff retry configuration
	maxRetries := 5
	baseDelay := 100 * time.Millisecond
	maxDelay := 5 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Infof("Stopping container %s (attempt %d/%d)...", containerId, attempt, maxRetries)

		err := d.apiClient.ContainerStop(ctx, containerId, container.StopOptions{
			Signal: "SIGKILL",
		})
		if err == nil {
			break
		}

		if attempt < maxRetries {
			// Calculate exponential backoff delay
			delay := baseDelay * time.Duration(1<<(attempt-1))
			if delay > maxDelay {
				delay = maxDelay
			}

			log.Warnf("Failed to stop container %s (attempt %d/%d): %v. Retrying in %v...", containerId, attempt, maxRetries, err, delay)
			time.Sleep(delay)
			continue
		}

		return fmt.Errorf("failed to stop container after %d attempts: %w", maxRetries, err)
	}

	err := d.waitForContainerStopped(ctx, containerId, 10*time.Second)
	if err != nil {
		return err
	}

	d.cache.SetSandboxState(ctx, containerId, enums.SandboxStateStopped)

	return nil
}

func (d *DockerClient) waitForContainerStopped(ctx context.Context, containerId string, timeout time.Duration) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			return fmt.Errorf("timeout waiting for container %s to stop", containerId)
		case <-ticker.C:
			c, err := d.ContainerInspect(ctx, containerId)
			if err != nil {
				return err
			}

			if !c.State.Running {
				return nil
			}
		}
	}
}
