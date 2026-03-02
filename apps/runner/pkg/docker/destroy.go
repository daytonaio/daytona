// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"
	"time"

	"github.com/containerd/errdefs"
	"github.com/daytonaio/runner/pkg/common"
	"github.com/daytonaio/runner/pkg/models/enums"
	"github.com/docker/docker/api/types/container"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/daytonaio/common-go/pkg/utils"
)

func (d *DockerClient) Destroy(ctx context.Context, containerId string) error {
	startTime := time.Now()
	defer func() {
		obs, err := common.ContainerOperationDuration.GetMetricWithLabelValues("destroy")
		if err == nil {
			obs.Observe(time.Since(startTime).Seconds())
		}
	}()

	// Cancel a backup if it's already in progress
	backup_context, ok := backup_context_map.Get(containerId)
	if ok {
		backup_context.cancel()
	}

	ct, err := d.ContainerInspect(ctx, containerId)
	if err != nil {
		if common_errors.IsNotFoundError(err) {
			return nil
		}
		return err
	}

	// Ignore err because we want to destroy the container even if it exited
	state, _ := d.GetSandboxState(ctx, containerId)
	if state == enums.SandboxStateDestroyed || state == enums.SandboxStateDestroying {
		d.logger.DebugContext(ctx, "Sandbox is already destroyed or destroying", "containerId", containerId)
		return nil
	}

	if state == enums.SandboxStateStopped {
		err = d.apiClient.ContainerRemove(ctx, containerId, container.RemoveOptions{
			Force:         false,
			RemoveVolumes: true,
		})
		if err == nil {
			go func() {
				containerShortId := ct.ID[:12]
				err := d.netRulesManager.DeleteNetworkRules(containerShortId)
				if err != nil {
					d.logger.ErrorContext(ctx, "Failed to delete sandbox network settings", "error", err)
				}
			}()

			return nil
		}

		// Handle not found case
		if errdefs.IsNotFound(err) {
			return nil
		}

		d.logger.WarnContext(ctx, "Failed to remove stopped sandbox without force", "error", err)
		d.logger.WarnContext(ctx, "Trying to remove stopped sandbox with force")
	}

	// Use exponential backoff helper for container removal
	err = utils.RetryWithExponentialBackoff(
		ctx,
		fmt.Sprintf("remove sandbox %s", containerId),
		utils.DEFAULT_MAX_RETRIES,
		utils.DEFAULT_BASE_DELAY,
		utils.DEFAULT_MAX_DELAY,
		func() error {
			return d.apiClient.ContainerRemove(ctx, containerId, container.RemoveOptions{
				Force: true,
			})
		},
	)
	if err != nil {
		// Handle NotFound error case
		if errdefs.IsNotFound(err) {
			return nil
		}
		return err
	}

	go func() {
		containerShortId := ct.ID[:12]
		err := d.netRulesManager.DeleteNetworkRules(containerShortId)
		if err != nil {
			d.logger.ErrorContext(ctx, "Failed to delete sandbox network settings", "error", err)
		}
	}()

	return nil
}
