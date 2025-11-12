// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/containerd/errdefs"
	"github.com/daytonaio/runner/pkg/common"
	"github.com/daytonaio/runner/pkg/models/enums"
	"github.com/docker/docker/api/types/container"

	log "github.com/sirupsen/logrus"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/daytonaio/common-go/pkg/utils"
)

var DESTROYED_STATES []enums.SandboxState = []enums.SandboxState{
	enums.SandboxStateDestroyed,
	enums.SandboxStateDestroying,
}

func (d *DockerClient) Destroy(ctx context.Context, sandboxId string) error {
	startTime := time.Now()
	defer func() {
		obs, err := common.ContainerOperationDuration.GetMetricWithLabelValues("destroy")
		if err == nil {
			obs.Observe(time.Since(startTime).Seconds())
		}
	}()

	// Cancel a backup if it's already in progress
	backup_context, ok := backup_context_map.Get(sandboxId)
	if ok {
		backup_context.cancel()
	}

	ct, err := d.ContainerInspect(ctx, sandboxId)
	if err != nil {
		if common_errors.IsNotFoundError(err) {
			return nil
		}
		return err
	}

	// Ignore err because we want to destroy the container even if it exited
	state, _ := d.deduceSandboxState(ct)
	if slices.Contains(DESTROYED_STATES, state) {
		log.Debugf("Sandbox %s is already destroyed or destroying", sandboxId)
		return nil
	}

	if state == enums.SandboxStateStopped {
		err = d.apiClient.ContainerRemove(ctx, sandboxId, container.RemoveOptions{
			Force:         false,
			RemoveVolumes: true,
		})
		if err == nil {
			go func() {
				containerShortId := ct.ID[:12]
				err = d.netRulesManager.DeleteNetworkRules(containerShortId)
				if err != nil {
					log.Errorf("Failed to delete sandbox network settings: %v", err)
				}
			}()

			return nil
		}

		if err != nil && errdefs.IsNotFound(err) {
			// not returning not found error here because not found indicates it is already destroyed
			return nil
		}

		log.Warnf("Failed to remove stopped sandbox without force: %v", err)
		log.Warnf("Trying to remove stopped sandbox with force")
	}

	// Use exponential backoff helper for container removal
	err = utils.RetryWithExponentialBackoff(
		ctx,
		fmt.Sprintf("remove sandbox %s", sandboxId),
		utils.DEFAULT_MAX_RETRIES,
		utils.DEFAULT_BASE_DELAY,
		utils.DEFAULT_MAX_DELAY,
		func() error {
			return d.apiClient.ContainerRemove(ctx, sandboxId, container.RemoveOptions{
				Force: true,
			})
		},
	)
	if err != nil {
		// Handle NotFound error case
		if errdefs.IsNotFound(err) {
			// not returning not found error here because not found indicates it is already destroyed
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

	return nil
}
