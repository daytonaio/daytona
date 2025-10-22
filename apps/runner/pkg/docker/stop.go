// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"

	"github.com/daytonaio/common-go/pkg/utils"
	"github.com/daytonaio/runner/pkg/models/enums"
	"github.com/docker/docker/api/types/container"
)

func (d *DockerClient) Stop(ctx context.Context, containerId string) error {
	// Deduce sandbox state first
	state, err := d.DeduceSandboxState(ctx, containerId)
	if err == nil && state == enums.SandboxStateStopped {
		d.logger.DebugContext(ctx, "Sandbox is already stopped", "containerId", containerId)
		d.statesCache.SetSandboxState(ctx, containerId, enums.SandboxStateStopped)
		return nil
	}

	d.statesCache.SetSandboxState(ctx, containerId, enums.SandboxStateStopping)

	if err != nil {
		d.logger.WarnContext(ctx, "Failed to deduce sandbox state", "containerId", containerId, "error", err)
		d.logger.WarnContext(ctx, "Continuing with stop operation")
	}

	// Cancel a backup if it's already in progress
	backup_context, ok := backup_context_map.Get(containerId)
	if ok {
		backup_context.cancel()
	}

	err = d.stopContainerWithRetry(ctx, containerId, 2)
	if err != nil {
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

	d.logger.DebugContext(ctx, "Sandbox stopped successfully", "containerId", containerId)
	d.statesCache.SetSandboxState(ctx, containerId, enums.SandboxStateStopped)

	return nil
}

// stopContainerWithRetry attempts to stop the specified container by sending a stop signal,
// retrying the operation with exponential backoff up to a maximum number of attempts.
// If stopping fails after all retries, it falls back to forcefully killing the container.
//
// Parameters:
//   - ctx: context for cancellation and timeout
//   - containerId: ID of the container to stop
//   - timeout: number of seconds to wait for graceful stop before forcing a kill
//
// Returns an error if the container could not be stopped or killed.
func (d *DockerClient) stopContainerWithRetry(ctx context.Context, containerId string, timeout int) error {
	// Use exponential backoff helper for container stopping
	err := utils.RetryWithExponentialBackoff(
		ctx,
		fmt.Sprintf("stop sandbox %s", containerId),
		utils.DEFAULT_MAX_RETRIES,
		utils.DEFAULT_BASE_DELAY,
		utils.DEFAULT_MAX_DELAY,
		func() error {
			return d.apiClient.ContainerStop(ctx, containerId, container.StopOptions{
				Signal:  "SIGKILL",
				Timeout: &timeout,
			})
		},
	)
	if err != nil {
		d.logger.WarnContext(ctx, "Failed to stop sandbox for multiple attempts", "containerId", containerId, "attempts", utils.DEFAULT_MAX_RETRIES, "error", err)
		d.logger.WarnContext(ctx, "Trying to kill sandbox", "containerId", containerId)
		err = d.apiClient.ContainerKill(ctx, containerId, "KILL")
		if err != nil {
			d.logger.WarnContext(ctx, "Failed to kill sandbox", "containerId", containerId, "error", err)
		}
		return err
	}
	return nil
}
