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
	log "github.com/sirupsen/logrus"
)

// CleanupOrphanedVolumeMounts removes volume mount directories that are no longer used by any container.
// Throttled to run at most once per volumeCleanupIntervalSec (default 30s).
func (d *DockerClient) CleanupOrphanedVolumeMounts(ctx context.Context) {
	d.volumeCleanupMu.Lock()
	defer d.volumeCleanupMu.Unlock()

	if d.volumeCleanupIntervalSec > 0 && time.Since(d.lastVolumeCleanup) < time.Duration(d.volumeCleanupIntervalSec)*time.Second {
		return
	}
	d.lastVolumeCleanup = time.Now()

	mountDirs, err := filepath.Glob(filepath.Join(getVolumeMountBasePath(), volumeMountPrefix+"*"))
	if err != nil || len(mountDirs) == 0 {
		return
	}

	inUse := d.getInUseVolumeMounts(ctx)

	for _, dir := range mountDirs {
		if !inUse[strings.TrimSuffix(dir, "/")] {
			log.Infof("Cleaning orphaned volume mount: %s", dir)
			d.unmountAndRemoveDir(dir)
		}
	}
}

func (d *DockerClient) getInUseVolumeMounts(ctx context.Context) map[string]bool {
	inUse := make(map[string]bool)
	prefix := filepath.Join(getVolumeMountBasePath(), volumeMountPrefix)

	containers, err := d.apiClient.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return inUse
	}

	// Use Mounts from list response - avoids expensive ContainerInspect calls
	for _, ct := range containers {
		for _, m := range ct.Mounts {
			src := strings.TrimSuffix(m.Source, "/")
			if strings.HasPrefix(src, prefix) {
				inUse[src] = true
			}
		}
	}

	return inUse
}

func (d *DockerClient) unmountAndRemoveDir(path string) {
	if !strings.Contains(path, volumeMountPrefix) {
		return
	}

	if d.isDirectoryMounted(path) {
		if err := exec.Command("umount", path).Run(); err != nil {
			log.Errorf("Failed to unmount %s: %v", path, err)
			return
		}
	}

	if err := os.RemoveAll(path); err != nil {
		log.Errorf("Failed to remove %s: %v", path, err)
	}
}
