// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"
	"time"

	"github.com/daytonaio/common-go/pkg/utils"
	"github.com/daytonaio/runner/pkg/api/dto"
	"github.com/daytonaio/runner/pkg/common"
	"github.com/daytonaio/runner/pkg/models/enums"

	"github.com/docker/docker/api/types/container"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"

	log "github.com/sirupsen/logrus"
)

func (d *DockerClient) Resize(ctx context.Context, sandboxId string, sandboxDto dto.ResizeSandboxDTO) error {
	// Handle disk resize (requires container recreation)
	// Value of 0 means "don't change" (minimum valid value is 1)
	if sandboxDto.Disk > 0 {
		// Validate container is stopped (disk resize is cold-only)
		containerInfo, err := d.ContainerInspect(ctx, sandboxId)
		if err != nil {
			return fmt.Errorf("failed to inspect container: %w", err)
		}
		if containerInfo.State.Running {
			return fmt.Errorf("disk resize requires stopped container")
		}

		err = d.ContainerDiskResize(ctx, sandboxId, float64(sandboxDto.Disk), sandboxDto.Cpu, sandboxDto.Memory, "resize")
		if err != nil {
			return err
		}
		// CPU/memory already applied during container recreation
		return nil
	}

	// Check if there's anything to resize (CPU/memory only, no disk change)
	if sandboxDto.Cpu == 0 && sandboxDto.Memory == 0 {
		return nil // Nothing to resize
	}

	// Get the current state to restore after resize
	originalState, err := d.DeduceSandboxState(ctx, sandboxId)
	if err != nil {
		// Default to started if we can't deduce state
		originalState = enums.SandboxStateStarted
	}

	d.statesCache.SetSandboxState(ctx, sandboxId, enums.SandboxStateResizing)

	// Build resources with only the fields that need to change (0 = don't change)
	resources := container.Resources{}
	if sandboxDto.Cpu > 0 {
		resources.CPUQuota = sandboxDto.Cpu * 100000 // 1 core = 100000
		resources.CPUPeriod = 100000
	}
	if sandboxDto.Memory > 0 {
		resources.Memory = common.GBToBytes(float64(sandboxDto.Memory))
		resources.MemorySwap = resources.Memory // Disable swap
	}

	_, err = d.apiClient.ContainerUpdate(ctx, sandboxId, container.UpdateConfig{
		Resources: resources,
	})
	if err != nil {
		d.statesCache.SetSandboxState(ctx, sandboxId, originalState)
		return err
	}

	d.statesCache.SetSandboxState(ctx, sandboxId, originalState)

	return nil
}

// ContainerDiskResize recreates a container with new storage size, preserving data via rsync.
// Optionally updates CPU/memory at the same time (0 = don't change).
// Used by both storage recovery and disk resize.
// Container must be stopped before calling this function.
func (d *DockerClient) ContainerDiskResize(ctx context.Context, sandboxId string, newStorageGB float64, cpu int64, memory int64, operationName string) error {
	log.Infof("Starting %s for sandbox %s with new storage %.2fGB", operationName, sandboxId, newStorageGB)

	originalContainer, err := d.ContainerInspect(ctx, sandboxId)
	if err != nil {
		return fmt.Errorf("failed to inspect container: %w", err)
	}

	// Get overlay2 path for data copy
	var overlayDiffPath string
	if originalContainer.GraphDriver.Name == "overlay2" {
		if upperDir, ok := originalContainer.GraphDriver.Data["UpperDir"]; ok {
			overlayDiffPath = upperDir
			log.Debugf("Overlay2 UpperDir: %s", overlayDiffPath)
		}
	}

	// Get filesystem type to determine if we can use storage-opt
	info, err := d.apiClient.Info(ctx)
	if err != nil {
		return fmt.Errorf("failed to get docker info: %w", err)
	}

	filesystem := d.getFilesystem(info)
	if filesystem != "xfs" {
		return fmt.Errorf("%s requires XFS filesystem, current filesystem: %s", operationName, filesystem)
	}

	// Rename container after validation checks to reduce error handling complexity
	timestamp := time.Now().Unix()
	oldName := fmt.Sprintf("%s-%s-%d", sandboxId, operationName, timestamp)
	log.Debugf("Renaming container to %s", oldName)

	err = d.apiClient.ContainerRename(ctx, sandboxId, oldName)
	if err != nil {
		return fmt.Errorf("failed to rename container: %w", err)
	}

	// Ensure the image is available for container recreation.
	// If the image tag was pruned (e.g., declarative-build or backup snapshot),
	// fall back to the image ID — Docker retains layers while the container exists.
	imageRef := originalContainer.Config.Image
	imageExists, _ := d.ImageExists(ctx, imageRef, true)
	if !imageExists {
		log.Warnf("Image %s not found by tag, falling back to image ID %s", imageRef, originalContainer.Image)
		originalContainer.Config.Image = originalContainer.Image
	}

	// Create new container with new storage
	newHostConfig := originalContainer.HostConfig
	newStorageBytes := common.GBToBytes(newStorageGB)
	if newHostConfig.StorageOpt == nil {
		newHostConfig.StorageOpt = make(map[string]string)
	}
	newHostConfig.StorageOpt["size"] = fmt.Sprintf("%d", newStorageBytes)
	log.Debugf("Setting storage to %d bytes (%.2fGB) on %s filesystem",
		newStorageBytes, float64(newStorageBytes)/(1024*1024*1024), filesystem)

	// Apply CPU/memory changes if specified (0 = don't change)
	if cpu > 0 {
		newHostConfig.CPUQuota = cpu * 100000
		newHostConfig.CPUPeriod = 100000
		log.Debugf("Setting CPU quota to %d cores", cpu)
	}
	if memory > 0 {
		newHostConfig.Memory = common.GBToBytes(float64(memory))
		newHostConfig.MemorySwap = newHostConfig.Memory
		log.Debugf("Setting memory to %dGB", memory)
	}

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

	// Copy data directly between overlay2 layers using rsync
	if overlayDiffPath != "" {
		log.Debug("Copying data directly between overlay2 layers using rsync")
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

	log.Infof("%s completed - container ready to be started", operationName)
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
