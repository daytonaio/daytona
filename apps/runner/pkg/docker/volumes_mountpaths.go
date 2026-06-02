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
	"github.com/google/uuid"
	"github.com/daytonaio/runner/pkg/volume"
)

const volumeMountPrefix = "daytona-volume-"

// volumeId becomes part of the host mount path and the S3 bucket name, so require
// the canonical lowercase UUID form (rejects braced/URN/dashless/uppercase variants,
// which uuid.Parse would otherwise accept).
func isValidVolumeId(volumeId string) bool {
	parsed, err := uuid.Parse(volumeId)
	if err != nil {
		return false
	}
	return parsed.String() == volumeId
}

func getVolumeMountBasePath() string {
	if config.GetEnvironment() == "development" {
		return "/tmp"
	}
	return "/mnt"
}

func (d *DockerClient) getVolumesMountPathBinds(ctx context.Context, volumes []dto.VolumeDTO, mounter volume.Mounter) ([]string, error) {
	// In-container mounters create no host mounts or binds; the daemon mounts
	// inside the sandbox.
	if _, ok := mounter.(volume.InContainerMounter); ok {
		return nil, nil
	}

	// Phase 1: fan out FUSE mounts for unique volumes in parallel.
	uniqueMounts := make(map[string]string, len(volumes)) // volumeIdPrefixed -> baseMountPath
	mountBase := filepath.Clean(getVolumeMountBasePath())
	for _, vol := range volumes {
		if !isValidVolumeId(vol.VolumeId) {
			return nil, fmt.Errorf("invalid volumeId %q: must be a volume UUID", vol.VolumeId)
		}
		volumeIdPrefixed := fmt.Sprintf("%s%s", volumeMountPrefix, vol.VolumeId)
		if _, ok := uniqueMounts[volumeIdPrefixed]; !ok {
			baseMountPath := filepath.Join(getVolumeMountBasePath(), volumeIdPrefixed)
			// Defense in depth: the path must stay a direct child of mountBase so a
			// traversal string can never escape it or collide with another volume.
			if filepath.Dir(baseMountPath) != mountBase || filepath.Base(baseMountPath) != volumeIdPrefixed {
				return nil, fmt.Errorf("invalid volumeId %q: resolves outside volume mount base", vol.VolumeId)
			}
			uniqueMounts[volumeIdPrefixed] = baseMountPath
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
			if err := d.ensureVolumeMounted(mountCtx, volumeId, mountPath, mounter); err != nil {
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
			if !strings.HasPrefix(filepath.Clean(bindSource), filepath.Clean(baseMountPath)) {
				return nil, fmt.Errorf("invalid subpath %q: resolves outside volume mount", *vol.Subpath)
			}
			err := os.MkdirAll(bindSource, 0755)
			if err != nil {
				return nil, fmt.Errorf("failed to create subpath directory %s: %s", bindSource, err)
			}
		}

		// The host-side mount-s3 stays writable (it's shared across every
		// sandbox using this volume); enforce read-only at the bind layer so
		// each sandbox gets its own independent RW/RO view.
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

// unmountVolume unmounts the host path. Only host-side mounts exist on disk
// (the in-container backend never creates one), so we always use the default
// mounter.
func (d *DockerClient) unmountVolume(ctx context.Context, mountPath string) error {
	return d.defaultVolumeMounter.Unmount(ctx, mountPath)
}

// isDirectoryMounted checks whether a path is an active mountpoint.
func (d *DockerClient) isDirectoryMounted(path string) bool {
	return d.defaultVolumeMounter.IsMounted(path)
}
