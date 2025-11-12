// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"
	"time"

	"github.com/daytonaio/common-go/pkg/utils"
	"github.com/daytonaio/runner/pkg/common"
	"github.com/docker/docker/api/types/container"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"

	log "github.com/sirupsen/logrus"
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

	log.Infof("Storage recovery for sandbox %s: original=%.2fGB, current=%.2fGB, currentExpansion=%.2fGB, increment=%.2fGB, newExpansion=%.2fGB, newTotal=%.2fGB, max=%.2fGB",
		sandboxId, originalStorageQuota, currentStorage, currentExpansion, increment, newExpansion, newStorageQuota, maxExpansion)

	// Validate expansion limit
	if newExpansion > maxExpansion {
		return fmt.Errorf("storage cannot be further expanded")
	}

	var overlayDiffPath string
	if originalContainer.GraphDriver.Name == "overlay2" {
		if upperDir, ok := originalContainer.GraphDriver.Data["UpperDir"]; ok {
			overlayDiffPath = upperDir
			log.Debugf("Overlay2 UpperDir: %s", overlayDiffPath)
		}
	}

	if originalContainer.State.Running {
		log.Info("Stopping sandbox")
		err = d.stopContainerWithRetry(ctx, sandboxId, 2)
		if err != nil {
			return fmt.Errorf("failed to stop sandbox: %w", err)
		}
	}

	log.Info("Creating new container with expanded storage")

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
	log.Debugf("Renaming container to %s", oldName)

	err = d.apiClient.ContainerRename(ctx, sandboxId, oldName)
	if err != nil {
		return fmt.Errorf("failed to rename container: %w", err)
	}

	newStorageBytes := common.GBToBytes(newStorageQuota)
	if newHostConfig.StorageOpt == nil {
		newHostConfig.StorageOpt = make(map[string]string)
	}
	newHostConfig.StorageOpt["size"] = fmt.Sprintf("%d", newStorageBytes)
	log.Infof("Setting storage to %d bytes (%.2fGB) on %s filesystem",
		newStorageBytes, float64(newStorageBytes)/(1024*1024*1024), filesystem)

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

	// Copy data directly between overlay2 layers (no need to start container)
	// The API will trigger the normal start flow through SandboxManager
	if overlayDiffPath != "" {
		log.Info("Copying data directly between overlay2 layers using rsync")
		err = d.copyContainerOverlayData(ctx, overlayDiffPath, sandboxId)
		if err != nil {
			log.Errorf("Failed to copy overlay data: %v", err)
			log.Warnf("Old container preserved as %s for manual data recovery", oldName)
			_ = d.apiClient.ContainerRemove(ctx, sandboxId, container.RemoveOptions{Force: true})
			_ = d.apiClient.ContainerRename(ctx, oldName, sandboxId)
			return fmt.Errorf("failed to copy data: %w", err)
		}
		log.Debugf("Data copy completed")
	} else {
		log.Warn("Could not determine old container overlay2 path, skipping data copy")
	}

	// Remove old container after successful data copy
	log.Debugf("Removing old container %s", oldName)
	err = d.apiClient.ContainerRemove(ctx, oldName, container.RemoveOptions{Force: true})
	if err != nil {
		log.Warnf("Failed to remove old container %s: %v", oldName, err)
	}

	// Note: Container is now stopped. The API will emit a STARTED event
	// which will trigger the normal start flow through SandboxManager
	log.Info("Storage expansion completed - container ready to be started by SandboxManager")

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
			log.Debugf("New container overlay2 UpperDir: %s", newUpperDir)
		}
	}

	if newUpperDir == "" {
		log.Warn("Could not determine new container overlay2 path, skipping data copy")
		return nil
	}

	log.Debugf("Copying overlay data from %s to %s", oldContainerOverlayPath, newUpperDir)

	// Use rsync with timeout to copy data
	copyCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	return common.RsyncCopy(copyCtx, oldContainerOverlayPath, newUpperDir)
}
