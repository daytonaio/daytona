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
	"time"

	"github.com/daytonaio/runner/cmd/runner/config"
	"github.com/docker/docker/api/types/container"
	log "github.com/sirupsen/logrus"
)

// RunMountCleanup runs the scheduled checks
func (d *DockerClient) RunMountCleanup(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	// Run immediately on startup
	d.checkAndCleanMounts(ctx)

	for {
		select {
		case <-ticker.C:
			d.checkAndCleanMounts(ctx)
		case <-ctx.Done():
			log.Println("Mount cleanup stopped")
			return
		}
	}
}

// checkAndCleanMounts checks all mount directories and cleans unused ones
func (d *DockerClient) checkAndCleanMounts(ctx context.Context) {
	log.Println("Starting mount check and cleanup")

	// Get all mount directories matching our criteria
	mountDirs, err := d.getMountDirectories()
	if err != nil {
		log.Printf("Error finding mount directories: %v", err)
		return
	}

	if len(mountDirs) == 0 {
		log.Println("No mount directories found")
		return
	}

	log.Printf("Found %d mount directories", len(mountDirs))

	// Check each directory
	for _, dir := range mountDirs {
		isUsed, err := d.isDirectoryUsedByContainers(ctx, dir)
		if err != nil {
			log.Printf("Error checking directory %s: %v", dir, err)
			continue
		}

		if isUsed {
			log.Printf("Directory %s is in use by containers", dir)
		} else {
			log.Printf("Directory %s is not in use, unmounting and deleting", dir)
			err := d.unmountAndDeleteDirectory(dir)
			if err != nil {
				log.Printf("Error unmounting/deleting %s: %v", dir, err)
			} else {
				log.Printf("Successfully unmounted and deleted %s", dir)
			}
		}
	}

	log.Println("Mount check and cleanup completed")
}

// isDirectoryUsedByContainers checks if any running container is using the specified directory
func (d *DockerClient) isDirectoryUsedByContainers(ctx context.Context, dirPath string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// List running containers
	containers, err := d.apiClient.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		return false, fmt.Errorf("failed to list containers: %w", err)
	}

	// Check each container's mounts
	for _, container := range containers {
		// Get detailed info about the container
		inspect, err := d.ContainerInspect(ctx, container.ID)
		if err != nil {
			return false, fmt.Errorf("failed to inspect container %s: %w", container.ID, err)
		}

		// Check mounts
		for _, mount := range inspect.Mounts {
			// Check if the mount source is the directory we're looking for
			// or if it's a subdirectory of the mount source or vice versa
			if mount.Source == dirPath || strings.HasPrefix(dirPath, mount.Source+"/") || strings.HasPrefix(mount.Source, dirPath+"/") {
				return true, nil
			}
		}
	}

	return false, nil
}

// getMountDirectories finds all directories under the base path with specific prefix that are mount points
func (d *DockerClient) getMountDirectories() ([]string, error) {
	basePath := "/mnt"
	if config.GetNodeEnv() == "development" {
		basePath = "/tmp"
	}

	prefix := "daytona-volume-"

	// Check if base path exists
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("base path %s does not exist", basePath)
	}

	// Get mount information
	output, err := exec.Command("mount").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get mount information: %w", err)
	}

	mountOutput := string(output)
	var mountDirs []string

	// List the directories directly under the base path
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read base directory: %w", err)
	}

	// Check each entry for the prefix and mount status
	for _, entry := range entries {
		// Skip if not a directory
		if !entry.IsDir() {
			continue
		}

		// Skip if doesn't have the required prefix
		if !strings.HasPrefix(entry.Name(), prefix) {
			continue
		}

		// Construct the full path
		fullPath := filepath.Join(basePath, entry.Name())

		// Check if this path appears in the mount output
		if strings.Contains(mountOutput, fullPath+" ") {
			mountDirs = append(mountDirs, fullPath)
		}
	}

	return mountDirs, nil
}

// unmountAndDeleteDirectory unmounts and then deletes the directory
func (d *DockerClient) unmountAndDeleteDirectory(dirPath string) error {
	// Unmount the directory
	cmd := exec.Command("umount", dirPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to unmount directory: %w", err)
	}

	// Delete the directory
	if err := os.RemoveAll(dirPath); err != nil {
		return fmt.Errorf("failed to delete directory: %w", err)
	}

	return nil
}
