// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
)

// getSandboxDiskId retrieves the disk ID from container labels
func (d *DockerClient) getSandboxDiskId(ctx context.Context, containerId string) (string, error) {
	info, err := d.ContainerInspect(ctx, containerId)
	if err != nil {
		return "", fmt.Errorf("failed to inspect container: %w", err)
	}

	diskId, exists := info.Config.Labels["daytona.disk_id"]
	if !exists || diskId == "" {
		// No disk attached to this sandbox
		return "", nil
	}

	return diskId, nil
}

// unmountSandboxDisk unmounts the disk attached to the sandbox
func (d *DockerClient) unmountSandboxDisk(ctx context.Context, containerId string) error {
	diskId, err := d.getSandboxDiskId(ctx, containerId)
	if err != nil {
		return err
	}

	if diskId == "" {
		// No disk attached, nothing to unmount
		return nil
	}

	disk, err := d.sdisk.Open(ctx, diskId)
	if err != nil {
		return fmt.Errorf("failed to open disk %s: %w", diskId, err)
	}
	defer disk.Close()

	if disk.IsMounted() {
		if err := disk.Unmount(ctx); err != nil {
			return fmt.Errorf("failed to unmount disk %s: %w", diskId, err)
		}
		log.Debugf("Unmounted disk %s for sandbox %s", diskId, containerId)
	}

	return nil
}

// mountSandboxDisk mounts the disk attached to the sandbox
func (d *DockerClient) mountSandboxDisk(ctx context.Context, containerId string) error {
	diskId, err := d.getSandboxDiskId(ctx, containerId)
	if err != nil {
		return err
	}

	if diskId == "" {
		// No disk attached, nothing to mount
		return nil
	}

	disk, err := d.sdisk.Open(ctx, diskId)
	if err != nil {
		return fmt.Errorf("failed to open disk %s: %w", diskId, err)
	}
	defer disk.Close()

	if !disk.IsMounted() {
		_, err := disk.Mount(ctx)
		if err != nil {
			return fmt.Errorf("failed to mount disk %s: %w", diskId, err)
		}
		log.Debugf("Mounted disk %s for sandbox %s", diskId, containerId)
	}

	return nil
}
