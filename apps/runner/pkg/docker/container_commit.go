// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"

	log "github.com/sirupsen/logrus"
)

func (d *DockerClient) commitContainer(ctx context.Context, containerId, imageName string) error {
	const maxRetries = 3

	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Infof("Committing container %s (attempt %d/%d)...", containerId, attempt, maxRetries)

		commitResp, err := d.apiClient.ContainerCommit(ctx, containerId, container.CommitOptions{
			Reference: imageName,
			Pause:     false,
		})
		if err == nil {
			log.Infof("Container %s committed successfully with image ID: %s", containerId, commitResp.ID)
			return nil
		}

		// Check if the error is related to "failed to get digest" and try export/import fallback
		if strings.Contains(err.Error(), "Error response from daemon: failed to get digest") {
			log.Warnf("Commit failed with digest error, attempting export/import fallback for container %s", containerId)

			err = d.exportImportContainer(ctx, containerId, imageName)
			if err == nil {
				log.Infof("Container %s successfully backed up using export/import method", containerId)
				return nil
			}

			log.Errorf("Export/import fallback also failed for container %s: %v", containerId, err)
		}

		if attempt < maxRetries {
			log.Warnf("Failed to commit container %s (attempt %d/%d): %v", containerId, attempt, maxRetries, err)
			continue
		}

		return fmt.Errorf("failed to commit container after %d attempts: %w", maxRetries, err)
	}

	return nil
}

func (d *DockerClient) exportImportContainer(ctx context.Context, containerId, imageName string) error {
	log.Infof("Exporting container %s and importing as image %s...", containerId, imageName)

	// First, inspect the container to get its configuration
	containerInfo, err := d.apiClient.ContainerInspect(ctx, containerId)
	if err != nil {
		return fmt.Errorf("failed to inspect container %s: %w", containerId, err)
	}

	// Export the container
	exportReader, err := d.apiClient.ContainerExport(ctx, containerId)
	if err != nil {
		return fmt.Errorf("failed to export container %s: %w", containerId, err)
	}
	defer exportReader.Close()

	// Prepare import options with container configuration
	importOptions := image.ImportOptions{
		Message: fmt.Sprintf("Imported from container %s", containerId),
	}

	// Build the configuration changes to preserve CMD, ENTRYPOINT, ENV, etc.
	var changes []string

	// Preserve CMD if it exists
	if len(containerInfo.Config.Cmd) > 0 {
		cmdStr := buildDockerfileCmd(containerInfo.Config.Cmd)
		changes = append(changes, fmt.Sprintf("CMD %s", cmdStr))
	}

	// Preserve ENTRYPOINT if it exists
	if len(containerInfo.Config.Entrypoint) > 0 {
		entrypointStr := buildDockerfileCmd(containerInfo.Config.Entrypoint)
		changes = append(changes, fmt.Sprintf("ENTRYPOINT %s", entrypointStr))
	}

	// Preserve environment variables
	if len(containerInfo.Config.Env) > 0 {
		for _, env := range containerInfo.Config.Env {
			changes = append(changes, fmt.Sprintf("ENV %s", env))
		}
	}

	// Preserve working directory
	if containerInfo.Config.WorkingDir != "" {
		changes = append(changes, fmt.Sprintf("WORKDIR %s", containerInfo.Config.WorkingDir))
	}

	// Preserve exposed ports
	if len(containerInfo.Config.ExposedPorts) > 0 {
		for port := range containerInfo.Config.ExposedPorts {
			changes = append(changes, fmt.Sprintf("EXPOSE %s", string(port)))
		}
	}

	// Preserve user
	if containerInfo.Config.User != "" {
		changes = append(changes, fmt.Sprintf("USER %s", containerInfo.Config.User))
	}

	// Apply the changes
	importOptions.Changes = changes

	log.Infof("Applying configuration changes: %v", changes)

	importResponse, err := d.apiClient.ImageImport(ctx, image.ImportSource{
		Source:     exportReader,
		SourceName: "-",
	}, imageName, importOptions)
	if err != nil {
		return fmt.Errorf("failed to import container %s as image %s: %w", containerId, imageName, err)
	}
	defer importResponse.Close()

	// Read the import response to completion
	_, err = io.ReadAll(importResponse)
	if err != nil {
		return fmt.Errorf("failed to read import response for container %s: %w", containerId, err)
	}

	log.Infof("Container %s successfully exported and imported as image %s with preserved configuration", containerId, imageName)
	return nil
}

// buildDockerfileCmd converts a slice of command arguments to a properly formatted Dockerfile CMD/ENTRYPOINT string
func buildDockerfileCmd(cmd []string) string {
	if len(cmd) == 0 {
		return ""
	}

	// Use JSON array format for better compatibility
	var quotedArgs []string
	for _, arg := range cmd {
		// Escape quotes and backslashes in the argument
		escaped := strings.ReplaceAll(arg, "\\", "\\\\")
		escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
		quotedArgs = append(quotedArgs, fmt.Sprintf("\"%s\"", escaped))
	}

	return fmt.Sprintf("[%s]", strings.Join(quotedArgs, ", "))
}
