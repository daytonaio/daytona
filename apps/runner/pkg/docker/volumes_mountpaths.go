// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/daytonaio/runner/cmd/runner/config"
	"github.com/daytonaio/runner/pkg/api/dto"
	"github.com/daytonaio/runner/pkg/volume"
)

const volumeMountPrefix = "daytona-volume-"

func getVolumeMountBasePath() string {
	if config.GetEnvironment() == "development" {
		return "/tmp"
	}
	return "/mnt"
}

func (d *DockerClient) getVolumesMountPathBinds(ctx context.Context, volumes []dto.VolumeDTO, mounter volume.Mounter) ([]string, error) {
	// In-container mounters don't create any host mounts or binds — the mount
	// happens inside the sandbox via the daemon. Return early.
	if _, ok := mounter.(volume.InContainerMounter); ok {
		return nil, nil
	}

	volumeMountPathBinds := make([]string, 0)

	// Tracks volumes with mounts already ensured in this call,
	// preventing duplicate mount attempts and mutex deadlocks when
	// multiple subpaths reference the same volume.
	mountedVolumes := make(map[string]bool)

	for _, vol := range volumes {
		volumeIdPrefixed := fmt.Sprintf("%s%s", volumeMountPrefix, vol.VolumeId)
		baseMountPath := filepath.Join(getVolumeMountBasePath(), volumeIdPrefixed)

		subpathStr := ""
		if vol.Subpath != nil {
			subpathStr = *vol.Subpath
		}

		if !mountedVolumes[volumeIdPrefixed] {
			err := d.ensureVolumeMounted(ctx, volumeIdPrefixed, baseMountPath, mounter)
			if err != nil {
				return nil, err
			}
			mountedVolumes[volumeIdPrefixed] = true
		}

		bindSource := baseMountPath
		if vol.Subpath != nil && *vol.Subpath != "" {
			bindSource = filepath.Join(baseMountPath, *vol.Subpath)
			if !strings.HasPrefix(filepath.Clean(bindSource), filepath.Clean(baseMountPath)) {
				return nil, fmt.Errorf("invalid subpath %q: resolves outside volume mount", *vol.Subpath)
			}
			err := os.MkdirAll(bindSource, 0755)
			if err != nil {
				return nil, fmt.Errorf("failed to create subpath directory %s: %s", bindSource, err)
			}
		}

		// Per-mount read-only support. The host-side mount-s3 stays
		// writable (it's shared across every sandbox referencing this
		// volume); we enforce read-only at the bind layer instead, so
		// each sandbox gets its own RW/RO view independent of any other.
		bindMode := ""
		if vol.ReadOnly {
			bindMode = ":ro"
		}
		d.logger.DebugContext(ctx, "binding volume subpath",
			"volumeId", volumeIdPrefixed,
			"subpath", subpathStr,
			"mountPath", vol.MountPath,
			"readOnly", vol.ReadOnly,
		)
		volumeMountPathBinds = append(volumeMountPathBinds, fmt.Sprintf("%s/:%s/%s", bindSource, vol.MountPath, bindMode))
	}

	return volumeMountPathBinds, nil
}

func (d *DockerClient) ensureVolumeMounted(ctx context.Context, volumeId string, mountPath string, mounter volume.Mounter) error {
	d.volumeMutexesMutex.Lock()
	volumeMutex, exists := d.volumeMutexes[volumeId]
	if !exists {
		volumeMutex = &sync.Mutex{}
		d.volumeMutexes[volumeId] = volumeMutex
	}
	d.volumeMutexesMutex.Unlock()

	volumeMutex.Lock()
	defer volumeMutex.Unlock()

	if mounter.IsMounted(mountPath) {
		d.logger.DebugContext(ctx, "volume already mounted", "volumeId", volumeId, "mountPath", mountPath)
		return nil
	}

	_, statErr := os.Stat(mountPath)
	dirExisted := statErr == nil

	err := os.MkdirAll(mountPath, 0755)
	if err != nil {
		return fmt.Errorf("failed to create mount directory %s: %s", mountPath, err)
	}

	err = mounter.Mount(ctx, volumeId, mountPath)
	if err != nil {
		if !dirExisted {
			removeErr := os.Remove(mountPath)
			if removeErr != nil {
				d.logger.WarnContext(ctx, "failed to remove mount directory", "path", mountPath, "error", removeErr)
			}
		}
		return fmt.Errorf("failed to mount volume %s to %s: %w", volumeId, mountPath, err)
	}

	err = mounter.WaitUntilReady(ctx, mountPath)
	if err != nil {
		if !dirExisted {
			umountErr := mounter.Unmount(ctx, mountPath)
			if umountErr != nil {
				d.logger.WarnContext(ctx, "failed to unmount during cleanup", "path", mountPath, "error", umountErr)
			}
			removeErr := os.Remove(mountPath)
			if removeErr != nil {
				d.logger.WarnContext(ctx, "failed to remove mount directory during cleanup", "path", mountPath, "error", removeErr)
			}
		}
		return fmt.Errorf("mount %s not ready after mounting: %w", mountPath, err)
	}

	d.logger.InfoContext(ctx, "mounted volume", "volumeId", volumeId, "mountPath", mountPath)
	return nil
}

// unmountVolume unmounts the volume at the given host path. Only host-side
// mounts exist on disk (the in-container backend never creates a host
// mountpoint), so we always delegate to the default mounter here.
func (d *DockerClient) unmountVolume(ctx context.Context, mountPath string) error {
	return d.defaultVolumeMounter.Unmount(ctx, mountPath)
}

// isDirectoryMounted checks whether a path is an active mountpoint.
func (d *DockerClient) isDirectoryMounted(path string) bool {
	return d.defaultVolumeMounter.IsMounted(path)
}
