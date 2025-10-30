// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"

	"github.com/daytonaio/runner/internal/constants"
	"github.com/daytonaio/runner/pkg/models/enums"
	"github.com/docker/docker/api/types/container"

	log "github.com/sirupsen/logrus"
)

func (d *DockerClient) Stop(ctx context.Context, containerId string) error {
	// Deduce sandbox state first
	state, err := d.DeduceSandboxState(ctx, containerId)
	if err == nil && state == enums.SandboxStateStopped {
		log.Debugf("Sandbox %s is already stopped", containerId)
		d.statesCache.SetSandboxState(ctx, containerId, enums.SandboxStateStopped)
		return nil
	}

	d.statesCache.SetSandboxState(ctx, containerId, enums.SandboxStateStopping)

	if err != nil {
		log.Warnf("Failed to deduce sandbox %s state: %v", containerId, err)
		log.Warnf("Continuing with stop operation")
	}

	// Cancel a backup if it's already in progress
	backup_context, ok := backup_context_map.Get(containerId)
	if ok {
		backup_context.cancel()
	}

	timeout := 2 // seconds
	// Use exponential backoff helper for container stopping
	err = d.retryWithExponentialBackoff(
		ctx,
		"stop",
		containerId,
		constants.DEFAULT_MAX_RETRIES,
		constants.DEFAULT_BASE_DELAY,
		constants.DEFAULT_MAX_DELAY,
		func() error {
			return d.apiClient.ContainerStop(ctx, containerId, container.StopOptions{
				Signal:  "SIGKILL",
				Timeout: &timeout,
			})
		},
	)
	if err != nil {
		log.Warnf("Failed to stop sandbox %s for %d attempts: %v", containerId, constants.DEFAULT_MAX_RETRIES, err)
		log.Warnf("Trying to kill sandbox %s", containerId)
		err = d.apiClient.ContainerKill(ctx, containerId, "KILL")
		if err != nil {
			log.Warnf("Failed to kill sandbox %s: %v", containerId, err)
		}
		return err
	}

	// Wait for container to actually stop
	statusCh, errCh := d.apiClient.ContainerWait(ctx, containerId, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("error waiting for sandbox %s to stop: %w", containerId, err)
		}
	case <-statusCh:
		// Container stopped successfully
		d.statesCache.SetSandboxState(ctx, containerId, enums.SandboxStateStopped)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}

	log.Debugf("Sandbox %s stopped successfully", containerId)
	d.statesCache.SetSandboxState(ctx, containerId, enums.SandboxStateStopped)

	return nil
}
