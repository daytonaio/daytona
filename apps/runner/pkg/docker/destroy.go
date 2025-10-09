// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"
	"time"

	"github.com/daytonaio/runner/internal/constants"
	"github.com/daytonaio/runner/pkg/common"
	"github.com/daytonaio/runner/pkg/models/enums"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/errdefs"

	log "github.com/sirupsen/logrus"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
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

	// Ignore err because we want to destroy the container even if it exited
	state, _ := d.DeduceSandboxState(ctx, containerId)
	if state == enums.SandboxStateDestroyed || state == enums.SandboxStateDestroying {
		return nil
	}

	d.statesCache.SetSandboxState(ctx, containerId, enums.SandboxStateDestroying)

	ct, err := d.ContainerInspect(ctx, containerId)
	if err != nil {
		if errdefs.IsNotFound(err) {
			d.statesCache.SetSandboxState(ctx, containerId, enums.SandboxStateDestroyed)
		}
		return err
	}

	// Use exponential backoff helper for container removal
	err = d.retryWithExponentialBackoff(
		"remove",
		containerId,
		constants.DEFAULT_MAX_RETRIES,
		constants.DEFAULT_BASE_DELAY,
		constants.DEFAULT_MAX_DELAY,
		func() error {
			return d.apiClient.ContainerRemove(ctx, containerId, container.RemoveOptions{
				Force: true,
			})
		},
	)
	if err != nil {
		// Handle NotFound error case
		if errdefs.IsNotFound(err) {
			d.statesCache.SetSandboxState(ctx, containerId, enums.SandboxStateDestroyed)
			return nil
		}
		return err
	}

	go func() {
		containerShortId := ct.ID[:12]
		err = d.netRulesManager.DeleteNetworkRules(containerShortId)
		if err != nil {
			log.Errorf("Failed to delete sandbox network settings: %v", err)
		}
	}()

	d.statesCache.SetSandboxState(ctx, containerId, enums.SandboxStateDestroyed)

	return nil
}

func (d *DockerClient) RemoveDestroyed(ctx context.Context, containerId string) error {

	// Check if container exists and is in destroyed state
	state, err := d.DeduceSandboxState(ctx, containerId)
	if err != nil {
		return err
	}

	if state != enums.SandboxStateDestroyed {
		return common_errors.NewBadRequestError(fmt.Errorf("container %s is not in destroyed state", containerId))
	}

	// Use exponential backoff helper for container removal
	err = d.retryWithExponentialBackoff(
		"remove",
		containerId,
		constants.DEFAULT_MAX_RETRIES,
		constants.DEFAULT_BASE_DELAY,
		constants.DEFAULT_MAX_DELAY,
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

	log.Infof("Destroyed container %s removed successfully", containerId)

	return nil
}
