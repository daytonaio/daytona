// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/daytonaio/common-go/pkg/timer"
	"github.com/daytonaio/runner/internal/constants"
	"github.com/daytonaio/runner/pkg/api/dto"
	"github.com/daytonaio/runner/pkg/common"
	"github.com/daytonaio/runner/pkg/models/enums"
	"github.com/daytonaio/runner/pkg/sdisk"
	"github.com/docker/docker/errdefs"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"

	log "github.com/sirupsen/logrus"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
)

func (d *DockerClient) Create(ctx context.Context, sandboxDto dto.CreateSandboxDTO) (string, error) {
	defer timer.Timer()()

	startTime := time.Now()
	defer func() {
		obs, err := common.ContainerOperationDuration.GetMetricWithLabelValues("create")
		if err == nil {
			obs.Observe(time.Since(startTime).Seconds())
		}
	}()

	state, err := d.DeduceSandboxState(ctx, sandboxDto.Id)
	if err != nil && state == enums.SandboxStateError {
		return "", err
	}

	if state == enums.SandboxStateStarted || state == enums.SandboxStatePullingSnapshot || state == enums.SandboxStateStarting {
		return sandboxDto.Id, nil
	}

	if state == enums.SandboxStateStopped || state == enums.SandboxStateCreating {
		err = d.Start(ctx, sandboxDto.Id, sandboxDto.Metadata)
		if err != nil {
			return "", err
		}

		return sandboxDto.Id, nil
	}

	d.statesCache.SetSandboxState(ctx, sandboxDto.Id, enums.SandboxStateCreating)

	ctx = context.WithValue(ctx, constants.ID_KEY, sandboxDto.Id)
	err = d.PullImage(ctx, sandboxDto.Snapshot, sandboxDto.Registry)
	if err != nil {
		return "", err
	}

	d.statesCache.SetSandboxState(ctx, sandboxDto.Id, enums.SandboxStateCreating)

	err = d.validateImageArchitecture(ctx, sandboxDto.Snapshot)
	if err != nil {
		log.Errorf("ERROR: %s.\n", err.Error())
		return "", err
	}

	volumeMountPathBinds := make([]string, 0)

	// Shared volumes
	if sandboxDto.Volumes != nil {
		volumeMountPathBinds, err = d.getVolumesMountPathBinds(ctx, sandboxDto.Volumes)
		if err != nil {
			return "", err
		}
	}

	// Sandbox Disk
	if sandboxDto.DiskId != "" {
		disks, err := d.sdisk.List(ctx)
		if err != nil {
			return "", err
		}
		diskExists := false
		for _, disk := range disks {
			if disk.Name == sandboxDto.DiskId {
				diskExists = true
				break
			}
		}
		var disk sdisk.Disk
		if !diskExists {
			disk, err = d.sdisk.Create(ctx, sandboxDto.DiskId, 10)
			if err != nil {
				return "", err
			}
		} else {
			disk, err = d.sdisk.Open(ctx, sandboxDto.DiskId)
			if err != nil {
				return "", err
			}
		}
		// CRITICAL: ALWAYS force a fresh mount to avoid any stale state issues
		// DETAILED LOGGING
		createLog, _ := os.OpenFile("/tmp/create-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if createLog != nil {
			fmt.Fprintf(createLog, "\n[CREATE] Starting container creation for disk %s\n", sandboxDto.DiskId)
			defer createLog.Close()
		}

		// Unmount if currently mounted (clears any stale state)
		log.Debugf("Force unmounting disk %s to ensure fresh mount", sandboxDto.DiskId)
		if createLog != nil {
			fmt.Fprintf(createLog, "[CREATE] Checking if disk is mounted: %v\n", disk.IsMounted())
		}
		if disk.IsMounted() {
			if createLog != nil {
				fmt.Fprintf(createLog, "[CREATE] Unmounting disk\n")
			}
			if err := disk.Unmount(ctx); err != nil {
				log.Warnf("Failed to unmount disk %s: %v", sandboxDto.DiskId, err)
				if createLog != nil {
					fmt.Fprintf(createLog, "[CREATE] ERROR: Unmount failed: %v\n", err)
				}
			}
		}

		// Force remove from pool and clear database state
		d.sdisk.ForceRemoveFromPool(sandboxDto.DiskId)
		log.Debugf("Cleared pool and database state for disk %s", sandboxDto.DiskId)
		if createLog != nil {
			fmt.Fprintf(createLog, "[CREATE] Cleared pool and database state\n")
		}

		// Now do a fresh mount
		if createLog != nil {
			fmt.Fprintf(createLog, "[CREATE] Calling disk.Mount\n")
		}
		mountPath, err := disk.Mount(ctx)
		if err != nil {
			if createLog != nil {
				fmt.Fprintf(createLog, "[CREATE] ERROR: Mount failed: %v\n", err)
			}
			return "", fmt.Errorf("failed to mount disk %s: %w", sandboxDto.DiskId, err)
		}
		log.Debugf("Successfully mounted disk %s at %s", sandboxDto.DiskId, mountPath)
		if createLog != nil {
			fmt.Fprintf(createLog, "[CREATE] Mount returned success, mountPath=%s\n", mountPath)
		}

		// Verify mount actually happened
		if createLog != nil {
			fmt.Fprintf(createLog, "[CREATE] Verifying mount in /proc/mounts\n")
		}
		verifyCmd := exec.CommandContext(ctx, "mount")
		if output, err := verifyCmd.CombinedOutput(); err == nil {
			if !strings.Contains(string(output), mountPath) {
				if createLog != nil {
					fmt.Fprintf(createLog, "[CREATE] ERROR: Mount path %s NOT in /proc/mounts!\n", mountPath)
				}
				return "", fmt.Errorf("disk mount verification failed: mount path %s not in /proc/mounts", mountPath)
			}
			log.Debugf("Verified disk %s is mounted at %s", sandboxDto.DiskId, mountPath)
			if createLog != nil {
				fmt.Fprintf(createLog, "[CREATE] SUCCESS: Verified mount in /proc/mounts\n")
			}
		}

		// Set up bind mount from the QCOW2 mount point directly to /workspace
		// No subdirectory needed - mount root contains the workspace files
		bindMount := fmt.Sprintf("%s:%s", mountPath, "/workspace")
		log.Debugf("Setting up disk bind mount: %s", bindMount)
		volumeMountPathBinds = append(volumeMountPathBinds, bindMount)
	}

	containerConfig, hostConfig, networkingConfig, err := d.getContainerConfigs(ctx, sandboxDto, volumeMountPathBinds)
	if err != nil {
		return "", err
	}

	c, err := d.apiClient.ContainerCreate(ctx, containerConfig, hostConfig, networkingConfig, &v1.Platform{
		Architecture: "amd64",
		OS:           "linux",
	}, sandboxDto.Id)
	if err != nil {
		return "", err
	}

	err = d.Start(ctx, sandboxDto.Id, sandboxDto.Metadata)
	if err != nil {
		return "", err
	}

	containerShortId := c.ID[:12]
	info, err := d.apiClient.ContainerInspect(context.Background(), sandboxDto.Id)
	if err != nil {
		log.Errorf("Failed to inspect container: %v", err)
	}
	ip := info.NetworkSettings.IPAddress

	if sandboxDto.NetworkBlockAll != nil && *sandboxDto.NetworkBlockAll {
		go func() {
			err = d.netRulesManager.SetNetworkRules(containerShortId, ip, "")
			if err != nil {
				log.Errorf("Failed to update sandbox network settings: %v", err)
			}
		}()
	} else if sandboxDto.NetworkAllowList != nil && *sandboxDto.NetworkAllowList != "" {
		go func() {
			err = d.netRulesManager.SetNetworkRules(containerShortId, ip, *sandboxDto.NetworkAllowList)
			if err != nil {
				log.Errorf("Failed to update sandbox network settings: %v", err)
			}
		}()
	}

	if sandboxDto.Metadata != nil && sandboxDto.Metadata["limitNetworkEgress"] == "true" {
		go func() {
			err = d.netRulesManager.SetNetworkLimiter(containerShortId, ip)
			if err != nil {
				log.Errorf("Failed to update sandbox network settings: %v", err)
			}
		}()
	}

	return c.ID, nil
}

func (p *DockerClient) validateImageArchitecture(ctx context.Context, image string) error {
	defer timer.Timer()()

	inspect, _, err := p.apiClient.ImageInspectWithRaw(ctx, image)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return err
		}
		return fmt.Errorf("failed to inspect image: %w", err)
	}

	arch := strings.ToLower(inspect.Architecture)
	validArchs := []string{"amd64", "x86_64"}

	for _, validArch := range validArchs {
		if arch == validArch {
			return nil
		}
	}

	return common_errors.NewConflictError(fmt.Errorf("image %s architecture (%s) is not x64 compatible", image, inspect.Architecture))
}
