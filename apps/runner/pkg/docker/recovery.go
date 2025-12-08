// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/daytonaio/common-go/pkg/timer"
	"github.com/daytonaio/runner/internal/constants"
	"github.com/daytonaio/runner/pkg/common"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-units"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"

	log "github.com/sirupsen/logrus"
)

// RecoverFromStorageLimit attempts to recover a sandbox from storage limit issues
// by expanding its storage quota by creating new ones with 100MB increments up to 10% of original.
func (d *DockerClient) RecoverFromStorageLimit(ctx context.Context, sandboxId string, originalStorageQuota float64) error {
	defer timer.Timer()()

	startTime := time.Now()
	defer func() {
		obs, err := common.ContainerOperationDuration.GetMetricWithLabelValues("recovery")
		if err == nil {
			obs.Observe(time.Since(startTime).Seconds())
		}
	}()

	originalContainer, err := d.ContainerInspect(ctx, sandboxId)
	if err != nil {
		common.ContainerOperationCount.WithLabelValues("recovery", string(common.PrometheusOperationStatusFailure)).Inc()
		return fmt.Errorf("failed to inspect container: %w", err)
	}

	// Get current storage size from StorageOpt
	currentStorage := float64(0)
	if originalContainer.HostConfig.StorageOpt != nil {
		if sizeStr, ok := originalContainer.HostConfig.StorageOpt["size"]; ok {
			// Docker storage-opt uses binary units (GiB) for both string ("3G") and numeric formats
			sizeBytes, err := units.RAMInBytes(sizeStr)
			if err != nil {
				common.ContainerOperationCount.WithLabelValues("recovery", string(common.PrometheusOperationStatusFailure)).Inc()
				return fmt.Errorf("failed to parse storage size '%s': %w", sizeStr, err)
			}
			currentStorage = float64(sizeBytes) / (1024 * 1024 * 1024)
		}
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
		common.ContainerOperationCount.WithLabelValues("recovery", string(common.PrometheusOperationStatusFailure)).Inc()
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

		timeout := 2
		err = d.retryWithExponentialBackoff(
			ctx,
			"stop",
			sandboxId,
			constants.DEFAULT_MAX_RETRIES,
			constants.DEFAULT_BASE_DELAY,
			constants.DEFAULT_MAX_DELAY,
			func() error {
				return d.apiClient.ContainerStop(ctx, sandboxId, container.StopOptions{
					Signal:  "SIGKILL",
					Timeout: &timeout,
				})
			},
		)
		if err != nil {
			log.Warnf("Failed to stop sandbox %s for %d attempts: %v", sandboxId, constants.DEFAULT_MAX_RETRIES, err)
			log.Warnf("Trying to kill sandbox %s", sandboxId)
			err = d.apiClient.ContainerKill(ctx, sandboxId, "KILL")
			if err != nil {
				common.ContainerOperationCount.WithLabelValues("recovery", string(common.PrometheusOperationStatusFailure)).Inc()
				return fmt.Errorf("failed to stop sandbox: %w", err)
			}
		}
	}

	timestamp := time.Now().Unix()
	oldName := fmt.Sprintf("%s-recovery-%d", sandboxId, timestamp)
	log.Debugf("Renaming container to %s", oldName)

	err = d.apiClient.ContainerRename(ctx, sandboxId, oldName)
	if err != nil {
		common.ContainerOperationCount.WithLabelValues("recovery", string(common.PrometheusOperationStatusFailure)).Inc()
		return fmt.Errorf("failed to rename container: %w", err)
	}

	log.Info("Creating new container with expanded storage")

	// Get filesystem type to determine if we can use storage-opt
	info, err := d.apiClient.Info(ctx)
	if err != nil {
		_ = d.apiClient.ContainerRename(ctx, oldName, sandboxId)
		common.ContainerOperationCount.WithLabelValues("recovery", string(common.PrometheusOperationStatusFailure)).Inc()
		return fmt.Errorf("failed to get docker info: %w", err)
	}

	newHostConfig := originalContainer.HostConfig
	filesystem := d.getFilesystem(info)

	if filesystem != "xfs" {
		_ = d.apiClient.ContainerRename(ctx, oldName, sandboxId)
		common.ContainerOperationCount.WithLabelValues("recovery", string(common.PrometheusOperationStatusFailure)).Inc()
		return fmt.Errorf("storage recovery requires XFS filesystem, current filesystem: %s", filesystem)
	}

	newStorageBytes := int64(newStorageQuota * 1024 * 1024 * 1024)
	if newHostConfig.StorageOpt == nil {
		newHostConfig.StorageOpt = make(map[string]string)
	}
	newHostConfig.StorageOpt["size"] = fmt.Sprintf("%d", newStorageBytes)
	log.Infof("Setting storage to %d bytes (%.2fGB) on %s filesystem",
		newStorageBytes, float64(newStorageBytes)/(1024*1024*1024), filesystem)

	err = d.retryWithExponentialBackoff(
		ctx,
		"create",
		sandboxId,
		constants.DEFAULT_MAX_RETRIES,
		constants.DEFAULT_BASE_DELAY,
		constants.DEFAULT_MAX_DELAY,
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
		common.ContainerOperationCount.WithLabelValues("recovery", string(common.PrometheusOperationStatusFailure)).Inc()
		return fmt.Errorf("failed to create new container: %w", err)
	}

	// Copy data directly between overlay2 layers (no need to start container)
	// The API will trigger the normal start flow through SandboxManager
	if overlayDiffPath != "" {
		// Get the new container's overlay2 UpperDir
		newContainer, err := d.ContainerInspect(ctx, sandboxId)
		if err != nil {
			log.Errorf("Failed to inspect new container: %v", err)
			_ = d.apiClient.ContainerRemove(ctx, sandboxId, container.RemoveOptions{Force: true})
			_ = d.apiClient.ContainerRename(ctx, oldName, sandboxId)
			common.ContainerOperationCount.WithLabelValues("recovery", string(common.PrometheusOperationStatusFailure)).Inc()
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
		} else {
			log.Info("Copying data directly between overlay2 layers using rsync")
			err = d.copyOverlayData(ctx, overlayDiffPath, newUpperDir)
			if err != nil {
				log.Errorf("Failed to copy data: %v", err)
				log.Warnf("Old container preserved as %s for manual data recovery", oldName)
				_ = d.apiClient.ContainerRemove(ctx, sandboxId, container.RemoveOptions{Force: true})
				_ = d.apiClient.ContainerRename(ctx, oldName, sandboxId)
				common.ContainerOperationCount.WithLabelValues("recovery", string(common.PrometheusOperationStatusFailure)).Inc()
				return fmt.Errorf("failed to copy data: %w", err)
			}
			log.Debugf("Data copy completed")
		}
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
	common.ContainerOperationCount.WithLabelValues("recovery", string(common.PrometheusOperationStatusSuccess)).Inc()

	return nil
}

// copyOverlayData copies the upper layer from old to new container overlay2 directories using rsync
// This preserves all file attributes, permissions, ownership, ACLs, and extended attributes
func (d *DockerClient) copyOverlayData(ctx context.Context, srcPath, destPath string) error {
	log.Debugf("Copying overlay data from %s to %s", srcPath, destPath)

	copyCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	// Use rsync with -aAX flags:
	// -a = archive mode (preserves permissions, ownership, timestamps, symlinks, devices)
	// -A = preserve ACLs
	// -X = preserve extended attributes (xattrs)
	// Trailing slashes ensure we copy contents, not the directory itself
	src := filepath.Clean(srcPath) + "/"
	dest := filepath.Clean(destPath) + "/"
	rsyncCmd := exec.CommandContext(copyCtx, "rsync", "-aAX", src, dest)

	var rsyncOut strings.Builder
	var rsyncErr strings.Builder
	rsyncCmd.Stdout = &rsyncOut
	rsyncCmd.Stderr = &rsyncErr

	log.Debug("Starting rsync...")
	if err := rsyncCmd.Run(); err != nil {
		if errMsg := rsyncErr.String(); errMsg != "" {
			log.Errorf("rsync stderr: %s", errMsg)
		}
		return fmt.Errorf("rsync failed: %w", err)
	}

	if outMsg := rsyncOut.String(); outMsg != "" {
		log.Debugf("rsync output: %s", outMsg)
	}

	log.Info("Successfully copied overlay data with rsync")
	return nil
}

func isStorageLimitError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "no space left on device") ||
		strings.Contains(errMsg, "storage limit") ||
		strings.Contains(errMsg, "disk quota exceeded")
}
