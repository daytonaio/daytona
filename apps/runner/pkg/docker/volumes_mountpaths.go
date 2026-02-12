// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/daytonaio/runner/cmd/runner/config"
	"github.com/daytonaio/runner/internal/util"
	"github.com/daytonaio/runner/pkg/api/dto"
	log "github.com/sirupsen/logrus"
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

	// Tracks volumes with FUSE mounts already ensured in this call,
	// preventing duplicate mount attempts and mutex deadlocks when
	// multiple subpaths reference the same volume.
	fuseMountedVolumes := make(map[string]bool)

	for _, vol := range volumes {
		volumeIdPrefixed := fmt.Sprintf("%s%s", volumeMountPrefix, vol.VolumeId)
		baseMountPath := filepath.Join(getVolumeMountBasePath(), volumeIdPrefixed)

		subpathStr := ""
		if vol.Subpath != nil {
			subpathStr = *vol.Subpath
		}

		if !fuseMountedVolumes[volumeIdPrefixed] {
			err := d.ensureVolumeFuseMounted(ctx, volumeIdPrefixed, baseMountPath)
			if err != nil {
				return nil, err
			}
			fuseMountedVolumes[volumeIdPrefixed] = true
		}

		bindSource := baseMountPath
		if vol.Subpath != nil && *vol.Subpath != "" {
			bindSource = filepath.Join(baseMountPath, *vol.Subpath)
			err := os.MkdirAll(bindSource, 0755)
			if err != nil {
				return nil, fmt.Errorf("failed to create subpath directory %s: %s", bindSource, err)
			}
		}

		log.Debugf("binding volume %s (subpath: %s) to %s", volumeIdPrefixed, subpathStr, vol.MountPath)
		volumeMountPathBinds = append(volumeMountPathBinds, fmt.Sprintf("%s/:%s/", bindSource, vol.MountPath))
	}

	return volumeMountPathBinds, nil
}

func (d *DockerClient) ensureVolumeFuseMounted(ctx context.Context, volumeId string, mountPath string) error {
	d.volumeMutexesMutex.Lock()
	volumeMutex, exists := d.volumeMutexes[volumeId]
	if !exists {
		volumeMutex = &sync.Mutex{}
		d.volumeMutexes[volumeId] = volumeMutex
	}
	d.volumeMutexesMutex.Unlock()

	volumeMutex.Lock()
	defer volumeMutex.Unlock()

	if d.isDirectoryMounted(mountPath) {
		log.Debugf("volume %s already mounted at %s", volumeId, mountPath)
		return nil
	}

	// Track if directory existed before we create it
	_, statErr := os.Stat(mountPath)
	dirExisted := statErr == nil

	err := os.MkdirAll(mountPath, 0755)
	if err != nil {
		return fmt.Errorf("failed to create mount directory %s: %s", mountPath, err)
	}

	log.Infof("mounting S3 volume %s to %s", volumeId, mountPath)

	cmd := d.getMountCmd(ctx, volumeId, mountPath)
	err = cmd.Run()
	if err != nil {
		removeErr := os.Remove(mountPath)
		if removeErr != nil {
			log.Warnf("failed to remove mount directory %s: %v", mountPath, removeErr)
		}
		return fmt.Errorf("failed to mount S3 volume %s to %s: %s", volumeId, mountPath, err)
	}

	err = d.waitForMountReady(ctx, mountPath)
	if err != nil {
		if !dirExisted {
			umountErr := exec.Command("umount", mountPath).Run()
			if umountErr != nil {
				log.Warnf("failed to unmount %s during cleanup: %v", mountPath, umountErr)
			}
			removeErr := os.Remove(mountPath)
			if removeErr != nil {
				log.Warnf("failed to remove mount directory %s during cleanup: %v", mountPath, removeErr)
			}
		}
		return fmt.Errorf("mount %s not ready after mounting: %s", mountPath, err)
	}

	log.Infof("mounted S3 volume %s to %s", volumeId, mountPath)
	return nil
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
				log.Infof("mount %s is ready after %d attempts", path, i+1)
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

func (d *DockerClient) getMountCmd(ctx context.Context, volume string, path string) *exec.Cmd {
	args := []string{"--allow-other", "--allow-delete", "--allow-overwrite", "--file-mode", "0666", "--dir-mode", "0777"}

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

	cmd.Stderr = io.Writer(&util.ErrorLogWriter{})
	cmd.Stdout = io.Writer(&util.InfoLogWriter{})

	return cmd
}
