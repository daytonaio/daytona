// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/docker/docker/api/types/container"
	log "github.com/sirupsen/logrus"
)

// normalizePath removes all trailing slashes from a path to ensure consistent comparison.
// This handles cases where Docker API might return paths with trailing slashes (single or multiple).
func normalizePath(path string) string {
	return strings.TrimRight(path, "/")
}

// CleanupOrphanedVolumeMounts removes volume mount directories that are no longer used by any container.
// Throttled to run at most once per volumeCleanupInterval (default 30s).
// Skips directories within exclusion period to avoid race conditions during sandbox creation.
func (d *DockerClient) CleanupOrphanedVolumeMounts(ctx context.Context) {
	d.volumeCleanupMutex.Lock()
	defer d.volumeCleanupMutex.Unlock()

	if d.volumeCleanupInterval > 0 && time.Since(d.lastVolumeCleanup) < d.volumeCleanupInterval {
		return
	}
	d.lastVolumeCleanup = time.Now()

	dryRun := d.volumeCleanupDryRun
	log.Infof("Volume cleanup dry-run: %v", dryRun)

	volumeMountBasePath := getVolumeMountBasePath()
	mountDirs, err := filepath.Glob(filepath.Join(volumeMountBasePath, volumeMountPrefix+"*"))
	if err != nil || len(mountDirs) == 0 {
		return
	}

	inUse, err := d.getInUseVolumeMounts(ctx)
	if err != nil {
		log.Errorf("Volume cleanup aborted: %v", err)
		return
	}

	exclusionPeriod := d.volumeCleanupExclusionPeriod

	for _, dir := range mountDirs {
		if inUse[normalizePath(dir)] {
			continue
		}
		if d.isRecentlyCreated(dir, exclusionPeriod) {
			continue
		}
		if dryRun {
			log.Infof("[DRY-RUN] Would clean orphaned volume mount: %s", dir)
			continue
		}
		log.Infof("Cleaning orphaned volume mount: %s", dir)
		d.unmountAndRemoveDir(dir)
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

func (d *DockerClient) unmountAndRemoveDir(path string) {
	mountBasePath := getVolumeMountBasePath()
	volumeMountPath := filepath.Join(mountBasePath, volumeMountPrefix)
	cleanPath := filepath.Clean(path)
	if !strings.HasPrefix(cleanPath, volumeMountPath) {
		return
	}

	if d.isDirectoryMounted(cleanPath) {
		if err := exec.Command("umount", cleanPath).Run(); err != nil {
			log.Errorf("Failed to unmount %s: %v", cleanPath, err)
			return
		}
		// Was FUSE mounted, data is on S3 - safe to remove
		if err := os.RemoveAll(cleanPath); err != nil {
			log.Errorf("Failed to remove %s: %v", cleanPath, err)
		}
		return
	}

	// Not mounted - might have unsynced local data
	if isDirEmpty(cleanPath) {
		if err := os.Remove(cleanPath); err != nil {
			log.Errorf("Failed to remove %s: %v", cleanPath, err)
		}
		return
	}

	timestamp := time.Now().Unix()
	garbagePath := filepath.Join(mountBasePath, fmt.Sprintf("garbage-%d-%s", timestamp, strings.TrimPrefix(filepath.Base(cleanPath), volumeMountPrefix)))
	log.Debugf("Renaming non-empty volume directory: %s", garbagePath)
	if err := os.Rename(cleanPath, garbagePath); err != nil {
		log.Errorf("Failed to rename %s: %v", cleanPath, err)
	}
}

func isDirEmpty(path string) bool {
	entries, err := os.ReadDir(path)
	if err != nil {
		log.Errorf("Failed to read directory %s: %v", path, err)
		return false
	}
	return len(entries) == 0
}

func (d *DockerClient) isRecentlyCreated(path string, exclusionPeriod time.Duration) bool {
	if exclusionPeriod <= 0 {
		return false
	}
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	// Use ctime (inode change time) instead of mtime (content modification time)
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		// Fallback to mtime if syscall.Stat_t is not available
		return time.Since(info.ModTime()) < exclusionPeriod
	}
	ctime := time.Unix(stat.Ctim.Sec, stat.Ctim.Nsec)
	return time.Since(ctime) < exclusionPeriod
}
