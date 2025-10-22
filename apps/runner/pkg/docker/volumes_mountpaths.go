// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/daytonaio/common-go/pkg/log"
	"github.com/daytonaio/runner/cmd/runner/config"
	"github.com/daytonaio/runner/pkg/api/dto"
)

const volumeMountPrefix = "daytona-volume-"

func getVolumeMountBasePath() string {
	if config.GetEnvironment() == "development" {
		return "/tmp"
	}
	return "/mnt"
}

func (d *DockerClient) getVolumesMountPathBinds(ctx context.Context, volumes []dto.VolumeDTO) ([]string, error) {
	volumeMountPathBinds := make([]string, 0)

	for _, vol := range volumes {
		volumeIdPrefixed := fmt.Sprintf("%s%s", volumeMountPrefix, vol.VolumeId)
		runnerVolumeMountPath := d.getRunnerVolumeMountPath(volumeIdPrefixed, vol.Subpath)

		// Create unique key for this volume+subpath combination for mutex
		volumeKey := d.getVolumeKey(volumeIdPrefixed, vol.Subpath)

		// Get or create mutex for this volume+subpath
		d.volumeMutexesMutex.Lock()
		volumeMutex, exists := d.volumeMutexes[volumeKey]
		if !exists {
			volumeMutex = &sync.Mutex{}
			d.volumeMutexes[volumeKey] = volumeMutex
		}
		d.volumeMutexesMutex.Unlock()

		// Lock this specific volume's mutex
		volumeMutex.Lock()
		defer volumeMutex.Unlock()

		subpathStr := ""
		if vol.Subpath != nil {
			subpathStr = *vol.Subpath
		}

		if d.isDirectoryMounted(runnerVolumeMountPath) {
			d.logger.InfoContext(ctx, "volume is already mounted", "volumeId", volumeIdPrefixed, "subpath", subpathStr, "runnerVolumeMountPath", runnerVolumeMountPath)
			volumeMountPathBinds = append(volumeMountPathBinds, fmt.Sprintf("%s/:%s/", runnerVolumeMountPath, vol.MountPath))
			continue
		}

		// Track if directory existed before we create it
		_, statErr := os.Stat(runnerVolumeMountPath)
		dirExisted := statErr == nil

		err := os.MkdirAll(runnerVolumeMountPath, 0755)
		if err != nil {
			return nil, fmt.Errorf("failed to create mount directory %s: %s", runnerVolumeMountPath, err)
		}

		d.logger.InfoContext(ctx, "mounting S3 volume", "volumeId", volumeIdPrefixed, "subpath", subpathStr, "runnerVolumeMountPath", runnerVolumeMountPath)

		cmd := d.getMountCmd(ctx, volumeIdPrefixed, vol.Subpath, runnerVolumeMountPath)
		err = cmd.Run()
		if err != nil {
			if !dirExisted {
				os.Remove(runnerVolumeMountPath)
			}
			return nil, fmt.Errorf("failed to mount S3 volume %s (subpath: %s) to %s: %s", volumeIdPrefixed, subpathStr, runnerVolumeMountPath, err)
		}

		// Wait for FUSE mount to be fully ready before proceeding
		err = d.waitForMountReady(ctx, runnerVolumeMountPath)
		if err != nil {
			if !dirExisted {
				umountErr := exec.Command("umount", runnerVolumeMountPath).Run()
				if umountErr != nil {
					d.logger.WarnContext(ctx, "Failed to unmount directory during cleanup", "path", runnerVolumeMountPath, "error", umountErr)
				}
				removeErr := os.Remove(runnerVolumeMountPath)
				if removeErr != nil {
					d.logger.WarnContext(ctx, "Failed to remove mount directory during cleanup", "path", runnerVolumeMountPath, "error", removeErr)
				}
			}
			return nil, fmt.Errorf("mount %s not ready after mounting: %s", runnerVolumeMountPath, err)
		}

		d.logger.InfoContext(ctx, "mounted S3 volume", "volumeId", volumeIdPrefixed, "subpath", subpathStr, "runnerVolumeMountPath", runnerVolumeMountPath)

		volumeMountPathBinds = append(volumeMountPathBinds, fmt.Sprintf("%s/:%s/", runnerVolumeMountPath, vol.MountPath))
	}

	return volumeMountPathBinds, nil
}

func (d *DockerClient) getRunnerVolumeMountPath(volumeId string, subpath *string) string {
	// If subpath is provided, create a unique mount point for this volume+subpath combination
	mountDirName := volumeId
	if subpath != nil && *subpath != "" {
		// Create a short hash of the subpath to keep the path reasonable
		hash := md5.Sum([]byte(*subpath))
		hashStr := hex.EncodeToString(hash[:])[:8]
		mountDirName = fmt.Sprintf("%s-%s", volumeId, hashStr)
	}

	volumePath := filepath.Join(getVolumeMountBasePath(), mountDirName)

	return volumePath
}

func (d *DockerClient) getVolumeKey(volumeId string, subpath *string) string {
	if subpath == nil || *subpath == "" {
		return volumeId
	}
	return fmt.Sprintf("%s:%s", volumeId, *subpath)
}

func (d *DockerClient) isDirectoryMounted(path string) bool {
	cmd := exec.Command("mountpoint", path)
	_, err := cmd.Output()

	return err == nil
}

// waitForMountReady waits for a FUSE mount to be fully accessible
// FUSE mounts can be asynchronous - the mount command may return before the filesystem is ready
// This prevents a race condition where the container writes to the directory before the mount is ready
func (d *DockerClient) waitForMountReady(ctx context.Context, path string) error {
	maxAttempts := 50 // 5 seconds total (50 * 100ms)
	sleepDuration := 100 * time.Millisecond

	for i := 0; i < maxAttempts; i++ {
		// First verify the mountpoint is still registered
		if !d.isDirectoryMounted(path) {
			return fmt.Errorf("mount disappeared during readiness check")
		}

		// Try to stat the mount point to ensure filesystem is responsive
		// This will fail if FUSE is not ready yet
		_, err := os.Stat(path)
		if err == nil {
			// Try to read directory to ensure it's fully operational
			_, err = os.ReadDir(path)
			if err == nil {
				d.logger.InfoContext(ctx, "mount is ready", "path", path, "attempts", i+1)
				return nil
			}
		}

		// Wait a bit before retrying
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled while waiting for mount ready: %w", ctx.Err())
		case <-time.After(sleepDuration):
			// Continue to next iteration
		}
	}

	return fmt.Errorf("mount did not become ready within timeout")
}

func (d *DockerClient) getMountCmd(ctx context.Context, volume string, subpath *string, path string) *exec.Cmd {
	args := []string{"--allow-other", "--allow-delete", "--allow-overwrite", "--file-mode", "0666", "--dir-mode", "0777"}

	if subpath != nil && *subpath != "" {
		// Ensure subpath ends with /
		prefix := *subpath
		if !strings.HasSuffix(prefix, "/") {
			prefix = prefix + "/"
		}
		args = append(args, "--prefix", prefix)
	}

	args = append(args, volume, path)

	cmd := exec.CommandContext(ctx, "mount-s3", args...)

	if d.awsEndpointUrl != "" {
		cmd.Env = append(cmd.Env, "AWS_ENDPOINT_URL="+d.awsEndpointUrl)
	}

	if d.awsAccessKeyId != "" {
		cmd.Env = append(cmd.Env, "AWS_ACCESS_KEY_ID="+d.awsAccessKeyId)
	}

	if d.awsSecretAccessKey != "" {
		cmd.Env = append(cmd.Env, "AWS_SECRET_ACCESS_KEY="+d.awsSecretAccessKey)
	}

	if d.awsRegion != "" {
		cmd.Env = append(cmd.Env, "AWS_REGION="+d.awsRegion)
	}

	cmd.Stderr = io.Writer(&log.ErrorLogWriter{})
	cmd.Stdout = io.Writer(&log.InfoLogWriter{})

	return cmd
}
