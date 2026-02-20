// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
)

// normalizePath removes all trailing slashes from a path to ensure consistent comparison.
// This handles cases where Docker API might return paths with trailing slashes (single or multiple).
func normalizePath(path string) string {
	return strings.TrimRight(path, "/")
}

// CleanupOrphanedVolumeMounts removes volume mount directories that are no longer used by any container.
// Throttled to run at most once per volumeCleanupIntervalSec (default 30s).
func (d *DockerClient) CleanupOrphanedVolumeMounts(ctx context.Context) {
	d.volumeCleanupMutex.Lock()
	defer d.volumeCleanupMutex.Unlock()

	if d.volumeCleanupIntervalSec > 0 && time.Since(d.lastVolumeCleanup) < time.Duration(d.volumeCleanupIntervalSec)*time.Second {
		return
	}
	d.lastVolumeCleanup = time.Now()

	dryRun := d.volumeCleanupDryRun
	d.logger.InfoContext(ctx, "Volume cleanup", "dry-run", dryRun)

	mountDirs, err := filepath.Glob(filepath.Join(getVolumeMountBasePath(), volumeMountPrefix+"*"))
	if err != nil || len(mountDirs) == 0 {
		return
	}

	inUse, err := d.getInUseVolumeMounts(ctx)
	if err != nil {
		d.logger.ErrorContext(ctx, "Volume cleanup aborted", "error", err)
		return
	}

	for _, dir := range mountDirs {
		if !inUse[normalizePath(dir)] {
			if dryRun {
				d.logger.InfoContext(ctx, "[DRY-RUN] Would clean orphaned volume mount", "path", dir)
			} else {
				d.logger.InfoContext(ctx, "Cleaning orphaned volume mount", "path", dir)
				d.unmountAndRemoveDir(ctx, dir)
			}
		}
	}
}

func (d *DockerClient) getInUseVolumeMounts(ctx context.Context) (map[string]bool, error) {
	inUse := make(map[string]bool)
	prefix := filepath.Join(getVolumeMountBasePath(), volumeMountPrefix)

	containers, err := d.apiClient.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, err
	}

	// Use Mounts from list response - avoids expensive ContainerInspect calls
	for _, ct := range containers {
		for _, m := range ct.Mounts {
			src := normalizePath(m.Source)
			if strings.HasPrefix(src, prefix) {
				inUse[src] = true
			}
		}
	}

	return inUse, nil
}

func (d *DockerClient) unmountAndRemoveDir(ctx context.Context, path string) {
	base := filepath.Join(getVolumeMountBasePath(), volumeMountPrefix)
	cleanPath := filepath.Clean(path)
	if !strings.HasPrefix(cleanPath, base) {
		return
	}

	if d.isDirectoryMounted(cleanPath) {
		if err := exec.Command("umount", cleanPath).Run(); err != nil {
			d.logger.ErrorContext(ctx, "Failed to unmount directory", "path", cleanPath, "error", err)
			return
		}
	}

	if err := os.RemoveAll(cleanPath); err != nil {
		d.logger.ErrorContext(ctx, "Failed to remove directory", "path", cleanPath, "error", err)
	}
}
