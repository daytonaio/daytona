// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"

	"github.com/daytonaio/runner/pkg/common"
)

// RecoverFromStorageLimit attempts to recover a sandbox from storage limit issues
// by expanding its storage quota by creating new ones with 100MB increments up to 10% of original.
func (d *DockerClient) RecoverFromStorageLimit(ctx context.Context, sandboxId string, originalStorageQuota float64) error {
	originalContainer, err := d.ContainerInspect(ctx, sandboxId)
	if err != nil {
		return fmt.Errorf("failed to inspect container: %w", err)
	}

	// Get current storage size from StorageOpt
	currentStorage := float64(0)
	if originalContainer.HostConfig.StorageOpt != nil {
		storageGB, err := common.ParseStorageOptSizeGB(originalContainer.HostConfig.StorageOpt)
		if err != nil {
			return err
		}
		currentStorage = storageGB
	}

	maxExpansion := originalStorageQuota * 0.1 // 10% of original
	currentExpansion := currentStorage - originalStorageQuota
	increment := 0.1 // ~107MB
	newExpansion := currentExpansion + increment
	newStorageQuota := originalStorageQuota + newExpansion

	d.logger.InfoContext(ctx, "Sandbox storage recovery",
		"sandboxId", sandboxId,
		"originalStorageQuota", originalStorageQuota,
		"currentStorage", currentStorage,
		"currentExpansion", currentExpansion,
		"increment", increment,
		"newExpansion", newExpansion,
		"newStorageQuota", newStorageQuota,
		"maxExpansion", maxExpansion,
	)

	// Validate expansion limit
	if newExpansion > maxExpansion {
		return fmt.Errorf("storage cannot be further expanded")
	}

	// Stop container if running
	if originalContainer.State.Running {
		d.logger.InfoContext(ctx, "Stopping sandbox", "sandboxId", sandboxId)
		err = d.stopContainerWithRetry(ctx, sandboxId, 2)
		if err != nil {
			return fmt.Errorf("failed to stop sandbox: %w", err)
		}
	}

	return d.ContainerDiskResize(ctx, sandboxId, newStorageQuota, 0, 0, "recovery")
}
