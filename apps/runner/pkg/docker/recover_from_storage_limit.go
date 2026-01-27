// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"
	"time"

	"github.com/daytonaio/common-go/pkg/utils"
	"github.com/daytonaio/runner/pkg/common"
	"github.com/daytonaio/runner/pkg/models/enums"
	"github.com/docker/docker/api/types/container"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
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

	d.log.InfoContext(ctx, "Sandbox storage recovery",
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

	var overlayDiffPath string
	if originalContainer.GraphDriver.Name == "overlay2" {
		if upperDir, ok := originalContainer.GraphDriver.Data["UpperDir"]; ok {
			overlayDiffPath = upperDir
			d.log.DebugContext(ctx, "Overlay UpperDir Path", "overlay", overlayDiffPath)
		}
	}

	if originalContainer.State.Running {
		d.log.InfoContext(ctx, "Stopping sandbox", "sandboxId", sandboxId)
		err = d.stopContainerWithRetry(ctx, sandboxId, 2)
		if err != nil {
			return fmt.Errorf("failed to stop sandbox: %w", err)
		}
	}

	d.log.InfoContext(ctx, "Creating new container with expanded storage", "sandboxId", sandboxId)

	// Get filesystem type to determine if we can use storage-opt
	info, err := d.apiClient.Info(ctx)
	if err != nil {
		return fmt.Errorf("failed to get docker info: %w", err)
	}

	newHostConfig := originalContainer.HostConfig
	filesystem := d.getFilesystem(info)

	if filesystem != "xfs" {
		return fmt.Errorf("storage recovery requires XFS filesystem, current filesystem: %s", filesystem)
	}

	// Rename container after validation checks to reduce error handling complexity
	timestamp := time.Now().Unix()
	oldName := fmt.Sprintf("%s-recovery-%d", sandboxId, timestamp)
	d.log.DebugContext(ctx, "Renaming container", "oldName", oldName)

	err = d.apiClient.ContainerRename(ctx, sandboxId, oldName)
	if err != nil {
		return fmt.Errorf("failed to rename container: %w", err)
	}

	newStorageBytes := common.GBToBytes(newStorageQuota)
	if newHostConfig.StorageOpt == nil {
		newHostConfig.StorageOpt = make(map[string]string)
	}
	newHostConfig.StorageOpt["size"] = fmt.Sprintf("%d", newStorageBytes)

	d.log.InfoContext(ctx, "Setting storage size on filesystem",
		"sizeBytes", newStorageBytes,
		"sizeGB", float64(newStorageBytes)/(1024*1024*1024),
		"filesystem", filesystem,
	)

	err = utils.RetryWithExponentialBackoff(
		ctx,
		fmt.Sprintf("create sandbox %s", sandboxId),
		utils.DEFAULT_MAX_RETRIES,
		utils.DEFAULT_BASE_DELAY,
		utils.DEFAULT_MAX_DELAY,
		func() error {
			_, createErr := d.apiClient.ContainerCreate(
				ctx,
				originalContainer.Config,
				newHostConfig,
				nil,
				&v1.Platform{
					Architecture: "amd64",
					OS:           "linux",
				},
				sandboxId,
			)
			return createErr
		},
	)
	if err != nil {
		_ = d.apiClient.ContainerRename(ctx, oldName, sandboxId)
		return fmt.Errorf("failed to create new container: %w", err)
	}

	d.statesCache.SetSandboxState(ctx, sandboxId, enums.SandboxStateStopped)

	// Copy data directly between overlay2 layers (no need to start container)
	// The API will trigger the normal start flow through SandboxManager
	if overlayDiffPath != "" {
		d.log.InfoContext(ctx, "Copying data directly between overlay2 layers using rsync")
		err = d.copyContainerOverlayData(ctx, overlayDiffPath, sandboxId)
		if err != nil {
			d.log.ErrorContext(ctx, "Failed to copy overlay data", "error", err)
			d.log.WarnContext(ctx, "Old container preserved for manual data recovery", "oldName", oldName)
			_ = d.apiClient.ContainerRemove(ctx, sandboxId, container.RemoveOptions{Force: true})
			_ = d.apiClient.ContainerRename(ctx, oldName, sandboxId)
			return fmt.Errorf("failed to copy data: %w", err)
		}
		d.log.DebugContext(ctx, "Data copy completed")
	} else {
		d.log.WarnContext(ctx, "Could not determine old container overlay2 path, skipping data copy")
	}

	// Remove old container after successful data copy
	d.log.DebugContext(ctx, "Removing old container", "oldName", oldName)
	err = d.apiClient.ContainerRemove(ctx, oldName, container.RemoveOptions{Force: true})
	if err != nil {
		d.log.WarnContext(ctx, "Failed to remove old container", "oldName", oldName, "error", err)
	}

	// Note: Container is now stopped. The API will emit a STARTED event
	// which will trigger the normal start flow through SandboxManager
	d.log.InfoContext(ctx, "Storage expansion completed - container ready to be started by SandboxManager", "sandboxId", sandboxId)

	return nil
}

// copyContainerOverlayData copies overlay2 data from old container path to new container
// by inspecting the new container for its overlay path and using rsync to copy the data
func (d *DockerClient) copyContainerOverlayData(ctx context.Context, oldContainerOverlayPath, newContainerId string) error {
	// Get the new container's overlay2 UpperDir
	newContainer, err := d.ContainerInspect(ctx, newContainerId)
	if err != nil {
		return fmt.Errorf("failed to inspect new container: %w", err)
	}

	var newUpperDir string
	if newContainer.GraphDriver.Name == "overlay2" {
		if upperDir, ok := newContainer.GraphDriver.Data["UpperDir"]; ok {
			newUpperDir = upperDir
			d.log.DebugContext(ctx, "New container overlay2 UpperDir", "upperDir", newUpperDir)
		}
	}

	if newUpperDir == "" {
		d.log.WarnContext(ctx, "Could not determine new container overlay2 path, skipping data copy")
		return nil
	}

	d.log.DebugContext(ctx, "Copying overlay data", "source", oldContainerOverlayPath, "destination", newUpperDir)
	// Use rsync with timeout to copy data
	copyCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	return common.RsyncCopy(copyCtx, oldContainerOverlayPath, newUpperDir)
}
