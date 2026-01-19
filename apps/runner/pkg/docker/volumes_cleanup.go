// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/daytonaio/runner/cmd/runner/config"
	"github.com/docker/docker/api/types/container"
	log "github.com/sirupsen/logrus"
)

const volumeMountPrefix = "daytona-volume-"

// CleanupOrphanedVolumeMounts removes volume mount directories that are no longer used by any container
func (d *DockerClient) CleanupOrphanedVolumeMounts(ctx context.Context) {
	basePath := "/mnt"
	if config.GetEnvironment() == "development" {
		basePath = "/tmp"
	}

	mountDirs, err := filepath.Glob(filepath.Join(basePath, volumeMountPrefix+"*"))
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

	containers, err := d.apiClient.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return inUse
	}

	for _, ct := range containers {
		info, err := d.apiClient.ContainerInspect(ctx, ct.ID)
		if err != nil || info.HostConfig == nil {
			continue
		}
		for _, bind := range info.HostConfig.Binds {
			parts := strings.Split(bind, ":")
			if len(parts) >= 1 {
				inUse[strings.TrimSuffix(parts[0], "/")] = true
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
