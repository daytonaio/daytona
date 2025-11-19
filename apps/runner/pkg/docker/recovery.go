// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/daytonaio/common-go/pkg/timer"
	"github.com/daytonaio/runner/internal/constants"
	"github.com/daytonaio/runner/pkg/common"
	"github.com/docker/docker/api/types/container"
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
			var sizeBytes int64
			fmt.Sscanf(sizeStr, "%d", &sizeBytes)
			currentStorage = float64(sizeBytes) / (1024 * 1024 * 1024)
		}
	}

	maxExpansion := originalStorageQuota * 0.1 // 10% of original
	currentExpansion := currentStorage - originalStorageQuota
	increment := 0.1 // 100MB
	newExpansion := currentExpansion + increment
	newStorageQuota := originalStorageQuota + newExpansion

	log.Infof("Storage recovery for sandbox %s: original=%.2fGB, current=%.2fGB, currentExpansion=%.2fGB, increment=%.2fGB, newExpansion=%.2fGB, newTotal=%.2fGB, max=%.2fGB",
		sandboxId, originalStorageQuota, currentStorage, currentExpansion, increment, newExpansion, newStorageQuota, maxExpansion)

	// Validate expansion limit
	if newExpansion > maxExpansion {
		common.ContainerOperationCount.WithLabelValues("recovery", string(common.PrometheusOperationStatusFailure)).Inc()
		return fmt.Errorf("storage cannot be expanded further. Maximum expansion of %.2fGB (10%% of original %.2fGB) has been reached. Please contact support", maxExpansion, originalStorageQuota)
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

	if filesystem == "xfs" {
		newStorageBytes := int64(newStorageQuota * 1024 * 1024 * 1024)
		if newHostConfig.StorageOpt == nil {
			newHostConfig.StorageOpt = make(map[string]string)
		}
		newHostConfig.StorageOpt["size"] = fmt.Sprintf("%d", newStorageBytes)
		log.Infof("Setting storage to %d bytes (%.2fGB) on %s filesystem",
			newStorageBytes, float64(newStorageBytes)/(1024*1024*1024), filesystem)
	} else {
		log.Warnf("Filesystem %s does not support storage-opt", filesystem)
	}

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

	// Start container temporarily to copy data, then stop it
	// The API will trigger the normal start flow through SandboxManager
	if overlayDiffPath != "" {
		log.Info("Starting container temporarily for data copy")
		err = d.retryWithExponentialBackoff(
			ctx,
			"start",
			sandboxId,
			constants.DEFAULT_MAX_RETRIES,
			constants.DEFAULT_BASE_DELAY,
			constants.DEFAULT_MAX_DELAY,
			func() error {
				return d.apiClient.ContainerStart(ctx, sandboxId, container.StartOptions{})
			},
		)
		if err != nil {
			log.Errorf("Failed to start container for data copy: %v", err)
			_ = d.apiClient.ContainerRemove(ctx, sandboxId, container.RemoveOptions{Force: true})
			_ = d.apiClient.ContainerRename(ctx, oldName, sandboxId)
			common.ContainerOperationCount.WithLabelValues("recovery", string(common.PrometheusOperationStatusFailure)).Inc()
			return fmt.Errorf("failed to start container for data copy: %w", err)
		}

		log.Info("Copying data from old to new container")
		err = d.copyOverlayData(ctx, oldName, sandboxId, overlayDiffPath)
		if err != nil {
			log.Errorf("Failed to copy data: %v", err)
			log.Warnf("Old container preserved as %s for manual data recovery", oldName)
			_ = d.apiClient.ContainerStop(ctx, sandboxId, container.StopOptions{})
			common.ContainerOperationCount.WithLabelValues("recovery", string(common.PrometheusOperationStatusFailure)).Inc()
			return fmt.Errorf("failed to copy data: %w", err)
		}
		log.Debugf("Data copy completed")

		log.Info("Stopping container after data copy")
		timeout := 2
		err = d.apiClient.ContainerStop(ctx, sandboxId, container.StopOptions{
			Signal:  "SIGKILL",
			Timeout: &timeout,
		})
		if err != nil {
			log.Warnf("Failed to stop container after data copy: %v", err)
		}
	} else {
		log.Warn("Could not determine overlay2 path, skipping data copy")
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

// copyOverlayData copies the upper layer from the old container to the new container using tar
func (d *DockerClient) copyOverlayData(ctx context.Context, oldName, newName, overlayPath string) error {
	log.Debugf("Copying overlay data from %s", overlayPath)

	copyCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	// tar to create archive from source, pipe to docker exec to extract in container
	tarCmd := exec.CommandContext(copyCtx, "tar", "-C", overlayPath, "-cf", "-", ".")
	dockerCmd := exec.CommandContext(copyCtx, "docker", "exec", "-i", newName, "tar", "-C", "/", "-xf", "-")

	pipe, err := tarCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create pipe: %w", err)
	}
	dockerCmd.Stdin = pipe

	var dockerErr strings.Builder
	dockerCmd.Stderr = &dockerErr

	if err := dockerCmd.Start(); err != nil {
		return fmt.Errorf("failed to start docker exec: %w", err)
	}

	if err := tarCmd.Start(); err != nil {
		return fmt.Errorf("failed to start tar: %w", err)
	}

	if err := tarCmd.Wait(); err != nil {
		return fmt.Errorf("tar failed: %w", err)
	}

	if err := dockerCmd.Wait(); err != nil {
		if errMsg := dockerErr.String(); errMsg != "" {
			log.Errorf("docker exec stderr: %s", errMsg)
		}
		return fmt.Errorf("docker exec failed: %w", err)
	}

	log.Info("Successfully copied overlay data")
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
