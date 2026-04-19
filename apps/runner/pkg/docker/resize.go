// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"
	"time"

	"github.com/containerd/errdefs"
	"github.com/daytonaio/common-go/pkg/utils"
	"github.com/daytonaio/runner/pkg/api/dto"
	"github.com/daytonaio/runner/pkg/common"

	"github.com/docker/docker/api/types/container"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
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

		err = d.ContainerDiskResize(ctx, sandboxId, float64(sandboxDto.Disk), sandboxDto.Cpu, sandboxDto.Memory, "resize", sandboxDto.Registry)
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

	_, err := d.apiClient.ContainerUpdate(ctx, sandboxId, container.UpdateConfig{
		Resources: resources,
	})
	if err != nil {
		return err
	}

	return nil
}

// ContainerDiskResize recreates a container with new storage size, preserving data via rsync.
// Optionally updates CPU/memory at the same time (0 = don't change).
// Used by both storage recovery and disk resize.
// Container must be stopped before calling this function.
func (d *DockerClient) ContainerDiskResize(ctx context.Context, sandboxId string, newStorageGB float64, cpu int64, memory int64, operationName string, registry *dto.RegistryDTO) error {
	if d.filesystem != "xfs" {
		return fmt.Errorf("%s requires XFS filesystem, current filesystem: %s", operationName, d.filesystem)
	}

	d.logger.InfoContext(ctx, "Starting operation for sandbox with new storage", "operation", operationName, "sandboxId", sandboxId, "newStorageGB", newStorageGB)

	originalContainer, err := d.ContainerInspect(ctx, sandboxId)
	if err != nil {
		return fmt.Errorf("failed to inspect container: %w", err)
	}

	// Get overlay2 path for data copy
	var overlayDiffPath string
	if originalContainer.GraphDriver.Name == "overlay2" {
		if upperDir, ok := originalContainer.GraphDriver.Data["UpperDir"]; ok {
			overlayDiffPath = upperDir
			d.logger.DebugContext(ctx, "Overlay2 UpperDir", "path", overlayDiffPath)
		}
	}

	resolvedImage, err := d.resolveContainerImage(ctx, sandboxId, originalContainer, registry)
	if err != nil {
		return err
	}
	originalContainer.Config.Image = resolvedImage

	// Rename container after validation checks to reduce error handling complexity
	timestamp := time.Now().Unix()
	oldName := fmt.Sprintf("%s-%s-%d", sandboxId, operationName, timestamp)
	d.logger.DebugContext(ctx, "Renaming sandbox", "oldName", oldName)

	err = d.apiClient.ContainerRename(ctx, sandboxId, oldName)
	if err != nil {
		return fmt.Errorf("failed to rename container: %w", err)
	}

	// Create new container with new storage
	newHostConfig := originalContainer.HostConfig
	newStorageBytes := common.GBToBytes(newStorageGB)
	if newHostConfig.StorageOpt == nil {
		newHostConfig.StorageOpt = make(map[string]string)
	}
	newHostConfig.StorageOpt["size"] = fmt.Sprintf("%d", newStorageBytes)
	d.logger.DebugContext(ctx, "Setting storage", "bytes", newStorageBytes, "gigabytes", float64(newStorageBytes)/(1024*1024*1024), "filesystem", d.filesystem)

	// Apply CPU/memory changes if specified (0 = don't change)
	if cpu > 0 {
		newHostConfig.CPUQuota = cpu * 100000
		newHostConfig.CPUPeriod = 100000
		d.logger.DebugContext(ctx, "Setting CPU quota", "cores", cpu)
	}
	if memory > 0 {
		newHostConfig.Memory = common.GBToBytes(float64(memory))
		newHostConfig.MemorySwap = newHostConfig.Memory
		d.logger.DebugContext(ctx, "Setting memory", "gigabytes", memory)
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

	// Copy data directly between overlay2 layers using rsync
	if overlayDiffPath != "" {
		d.logger.DebugContext(ctx, "Copying data directly between overlay2 layers using rsync")
		err = d.copyContainerOverlayData(ctx, overlayDiffPath, sandboxId)
		if err != nil {
			d.logger.ErrorContext(ctx, "Failed to copy overlay data", "error", err)
			d.logger.WarnContext(ctx, "Old sandbox preserved as for manual data recovery", "oldName", oldName)
			_ = d.apiClient.ContainerRemove(ctx, sandboxId, container.RemoveOptions{Force: true})
			_ = d.apiClient.ContainerRename(ctx, oldName, sandboxId)
			return fmt.Errorf("failed to copy data: %w", err)
		}
		d.logger.DebugContext(ctx, "Data copy completed")
	} else {
		d.logger.WarnContext(ctx, "Could not determine old container overlay2 path, skipping data copy")
	}

	// Remove old container after successful data copy
	d.logger.DebugContext(ctx, "Removing old container", "oldName", oldName)
	err = d.apiClient.ContainerRemove(ctx, oldName, container.RemoveOptions{Force: true})
	if err != nil {
		d.logger.WarnContext(ctx, "Failed to remove old container", "oldName", oldName, "error", err)
	}

	d.logger.InfoContext(ctx, "Operation completed - container ready to be started", "operation", operationName, "sandboxId", sandboxId)
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
			d.logger.DebugContext(ctx, "New container overlay2 UpperDir", "newUpperDir", newUpperDir)
		}
	}

	if newUpperDir == "" {
		d.logger.WarnContext(ctx, "Could not determine new container overlay2 path, skipping data copy")
		return nil
	}

	d.logger.DebugContext(ctx, "Copying overlay data", "from", oldContainerOverlayPath, "to", newUpperDir)

	// Use rsync with timeout to copy data
	copyCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	return common.RsyncCopy(copyCtx, d.logger, oldContainerOverlayPath, newUpperDir)
}

func (d *DockerClient) resolveContainerImage(ctx context.Context, sandboxId string, originalContainer *container.InspectResponse, registry *dto.RegistryDTO) (string, error) {
	imageRef := originalContainer.Config.Image

	_, err := d.apiClient.ImageInspect(ctx, imageRef)
	if err == nil {
		return imageRef, nil
	}
	if !errdefs.IsNotFound(err) {
		return "", fmt.Errorf("failed to inspect image %s: %w", imageRef, err)
	}

	_, idErr := d.apiClient.ImageInspect(ctx, originalContainer.Image)
	if idErr == nil {
		d.logger.WarnContext(ctx, "Image not found by tag, falling back to image ID",
			"imageRef", imageRef, "imageID", originalContainer.Image)
		return originalContainer.Image, nil
	}
	if !errdefs.IsNotFound(idErr) {
		return "", fmt.Errorf("failed to inspect image by ID %s: %w", originalContainer.Image, idErr)
	}

	if registry == nil {
		return "", fmt.Errorf("image %s is not available locally", imageRef)
	}
	d.logger.WarnContext(ctx, "Image not found locally, pulling from registry before resize",
		"containerName", originalContainer.Name, "imageRef", imageRef)
	if _, pullErr := d.PullImage(ctx, imageRef, registry, &sandboxId); pullErr != nil {
		return "", fmt.Errorf("image %s not available locally and pull from registry failed: %w", imageRef, pullErr)
	}
	return imageRef, nil
}
