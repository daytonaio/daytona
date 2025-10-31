// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/daytonaio/runner/internal/constants"
	"github.com/daytonaio/runner/pkg/models/enums"
	"github.com/docker/docker/api/types/container"
)

func (d *DockerClient) Stop(ctx context.Context, containerId string) error {
	// Deduce sandbox state first
	state, err := d.DeduceSandboxState(ctx, containerId)
	if err == nil && state == enums.SandboxStateStopped {
		slog.DebugContext(ctx, "Sandbox is already stopped", "containerId", containerId)
		d.statesCache.SetSandboxState(ctx, containerId, enums.SandboxStateStopped)
		return nil
	}

	d.statesCache.SetSandboxState(ctx, containerId, enums.SandboxStateStopping)

	if err != nil {
		slog.WarnContext(ctx, "Failed to deduce sandbox state", "containerId", containerId, "error", err)
		slog.WarnContext(ctx, "Continuing with stop operation")
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
		slog.WarnContext(ctx, "Failed to stop sandbox for attempts", "containerId", containerId, "attempts", constants.DEFAULT_MAX_RETRIES, "error", err)
		slog.WarnContext(ctx, "Trying to kill sandbox", "containerId", containerId)
		err = d.apiClient.ContainerKill(ctx, containerId, "KILL")
		if err != nil {
			slog.WarnContext(ctx, "Failed to kill sandbox", "containerId", containerId, "error", err)
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

	slog.DebugContext(ctx, "Sandbox stopped successfully", "containerId", containerId)
	d.statesCache.SetSandboxState(ctx, containerId, enums.SandboxStateStopped)

	return nil
}
