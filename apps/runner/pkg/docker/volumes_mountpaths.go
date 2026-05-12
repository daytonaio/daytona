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
	// Phase 1: fan out FUSE mounts for unique volumes in parallel. Each
	// ensureVolumeFuseMounted runs mount-s3 and then waits up to 5s for the
	// mount to become ready; doing them sequentially made create-time scale
	// linearly with the number of mounted volumes.
	uniqueMounts := make(map[string]string, len(volumes)) // volumeIdPrefixed -> baseMountPath
	for _, vol := range volumes {
		volumeIdPrefixed := fmt.Sprintf("%s%s", volumeMountPrefix, vol.VolumeId)
		if _, ok := uniqueMounts[volumeIdPrefixed]; !ok {
			uniqueMounts[volumeIdPrefixed] = filepath.Join(getVolumeMountBasePath(), volumeIdPrefixed)
		}
	}

	mountCtx, cancelMounts := context.WithCancel(ctx)
	defer cancelMounts()

	var (
		wg       sync.WaitGroup
		errMu    sync.Mutex
		firstErr error
	)
	for volumeIdPrefixed, baseMountPath := range uniqueMounts {
		wg.Add(1)
		go func(volumeId, mountPath string) {
			defer wg.Done()
			if err := d.ensureVolumeFuseMounted(mountCtx, volumeId, mountPath); err != nil {
				errMu.Lock()
				if firstErr == nil {
					firstErr = err
					cancelMounts()
				}
				errMu.Unlock()
			}
		}(volumeIdPrefixed, baseMountPath)
	}
	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
	}

	// Phase 2: build bind strings in input order. Subpath mkdir is cheap and
	// kept sequential so the returned slice order matches volumes.
	volumeMountPathBinds := make([]string, 0, len(volumes))
	for _, vol := range volumes {
		volumeIdPrefixed := fmt.Sprintf("%s%s", volumeMountPrefix, vol.VolumeId)
		baseMountPath := uniqueMounts[volumeIdPrefixed]

		subpathStr := ""
		if vol.Subpath != nil {
			subpathStr = *vol.Subpath
		}

		bindSource := baseMountPath
		if vol.Subpath != nil && *vol.Subpath != "" {
			bindSource = filepath.Join(baseMountPath, *vol.Subpath)
			// Ensure the resolved path stays within baseMountPath to prevent path traversal
			if !strings.HasPrefix(filepath.Clean(bindSource), filepath.Clean(baseMountPath)) {
				return nil, fmt.Errorf("invalid subpath %q: resolves outside volume mount", *vol.Subpath)
			}
			err := os.MkdirAll(bindSource, 0755)
			if err != nil {
				return nil, fmt.Errorf("failed to create subpath directory %s: %s", bindSource, err)
			}
		}

		d.logger.DebugContext(ctx, "binding volume subpath", "volumeId", volumeIdPrefixed, "subpath", subpathStr, "mountPath", vol.MountPath)
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
		d.logger.DebugContext(ctx, "volume already mounted", "volumeId", volumeId, "mountPath", mountPath)
		return nil
	}

	// Track if directory existed before we create it
	_, statErr := os.Stat(mountPath)
	dirExisted := statErr == nil

	err := os.MkdirAll(mountPath, 0755)
	if err != nil {
		return fmt.Errorf("failed to create mount directory %s: %s", mountPath, err)
	}

	d.logger.InfoContext(ctx, "mounting S3 volume", "volumeId", volumeId, "mountPath", mountPath)

	cmd := d.getMountCmd(ctx, volumeId, mountPath)
	err = cmd.Run()
	if err != nil {
		if !dirExisted {
			removeErr := os.Remove(mountPath)
			if removeErr != nil {
				d.logger.WarnContext(ctx, "failed to remove mount directory", "path", mountPath, "error", removeErr)
			}
		}
		return fmt.Errorf("failed to mount S3 volume %s to %s: %s", volumeId, mountPath, err)
	}

	err = d.waitForMountReady(ctx, mountPath)
	if err != nil {
		if !dirExisted {
			umountErr := exec.Command("umount", mountPath).Run()
			if umountErr != nil {
				d.logger.WarnContext(ctx, "failed to unmount during cleanup", "path", mountPath, "error", umountErr)
			}
			removeErr := os.Remove(mountPath)
			if removeErr != nil {
				d.logger.WarnContext(ctx, "failed to remove mount directory during cleanup", "path", mountPath, "error", removeErr)
			}
		}
		return fmt.Errorf("mount %s not ready after mounting: %s", mountPath, err)
	}

	d.logger.InfoContext(ctx, "mounted S3 volume", "volumeId", volumeId, "mountPath", mountPath)
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

func (d *DockerClient) getMountCmd(ctx context.Context, volume string, path string) *exec.Cmd {
	args := []string{"--allow-other", "--allow-delete", "--allow-overwrite", "--file-mode", "0666", "--dir-mode", "0777"}
	args = append(args, volume, path)

	var envVars []string
	if d.awsEndpointUrl != "" {
		envVars = append(envVars, "AWS_ENDPOINT_URL="+d.awsEndpointUrl)
	}
	if d.awsAccessKeyId != "" {
		envVars = append(envVars, "AWS_ACCESS_KEY_ID="+d.awsAccessKeyId)
	}
	if d.awsSecretAccessKey != "" {
		envVars = append(envVars, "AWS_SECRET_ACCESS_KEY="+d.awsSecretAccessKey)
	}
	if d.awsRegion != "" {
		envVars = append(envVars, "AWS_REGION="+d.awsRegion)
	}

	// No systemd (containerized) — daemon orphan survives runner restarts naturally.
	// CommandContext is used so ctx cancellation can stop a slow mount-s3 startup;
	// once mount-s3 daemonizes (no --foreground), cmd.Run returns and ctx no longer has a leash.
	cmd := exec.CommandContext(ctx, "mount-s3", args...)
	cmd.Env = envVars

	_, err := os.Stat("/run/systemd/system")
	if err == nil {
		// Isolate mount-s3 in its own cgroup so the FUSE daemon survives runner restarts.
		sdArgs := []string{"--scope"}
		for _, env := range envVars {
			sdArgs = append(sdArgs, "--setenv="+env)
		}
		sdArgs = append(sdArgs, "--", "mount-s3")
		sdArgs = append(sdArgs, args...)
		cmd = exec.CommandContext(ctx, "systemd-run", sdArgs...)
	}

	cmd.Stderr = io.Writer(&log.ErrorLogWriter{})
	cmd.Stdout = io.Writer(&log.InfoLogWriter{})

	return cmd
}
