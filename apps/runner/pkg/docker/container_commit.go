// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"
	"io"
	"os/exec"
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

			// Check if the committed image contains socket files
			hasSockets, err := d.checkImageForSockets(ctx, imageName)
			if err != nil {
				log.Warnf("Failed to check image for sockets: %v", err)
				// Continue with the commit even if socket check fails
				return nil
			}

			if hasSockets {
				log.Warnf("Image %s contains socket files, using export/import method", imageName)

				containerInspect, err := d.apiClient.ContainerInspect(ctx, containerId)
				if err != nil {
					log.Warnf("Failed to inspect container %s: %v", containerId, err)
					return fmt.Errorf("failed to inspect container %s: %w", containerId, err)
				}

				containerConfig := containerInspect.Config
				containerConfig.Image = imageName

				// Start the container from that image
				c, err := d.apiClient.ContainerCreate(ctx, containerConfig, containerInspect.HostConfig, nil, nil, fmt.Sprintf("socket-fix-%s", containerId))
				if err != nil {
					log.Warnf("Failed to start container from image %s: %v", imageName, err)
					return fmt.Errorf("failed to start container from image %s: %w", imageName, err)
				}
				defer d.apiClient.ContainerRemove(ctx, c.ID, container.RemoveOptions{
					RemoveVolumes: true,
					RemoveLinks:   false,
					Force:         true,
				})
				if err != nil {
					log.Warnf("Failed to remove container %s: %v", c.ID, err)
					return fmt.Errorf("failed to remove container %s: %w", c.ID, err)
				}

				// Start the container
				if err := d.apiClient.ContainerStart(ctx, c.ID, container.StartOptions{}); err != nil {
					log.Warnf("Failed to start container %s: %v", c.ID, err)
					return fmt.Errorf("failed to start container %s: %w", c.ID, err)
				}

				// Remove all sock files from the container
				if _, err := d.execSync(ctx, c.ID, container.ExecOptions{
					Cmd: []string{"find", "/", "-type", "s", "-delete"},
				}, container.ExecStartOptions{}); err != nil {
					log.Warnf("Failed to remove sock files from container %s: %v", c.ID, err)
					return fmt.Errorf("failed to remove sock files from container %s: %w", c.ID, err)
				}

				// Remove the problematic image
				_, err = d.apiClient.ImageRemove(ctx, imageName, image.RemoveOptions{
					Force: true,
				})
				if err != nil {
					log.Warnf("Failed to remove image with sockets: %v", err)
				}

				// Use export/import method to create a clean image
				err = d.exportImportContainer(ctx, c.ID, imageName)
				if err == nil {
					log.Infof("Container %s successfully backed up using export/import method after socket cleanup", containerId)
					return nil
				}

				log.Errorf("Export/import fallback failed for container %s: %v", containerId, err)
				return fmt.Errorf("failed to create image without sockets: %w", err)
			}

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

// checkImageForSockets checks if an image contains socket files by inspecting its GraphDriver data
func (d *DockerClient) checkImageForSockets(ctx context.Context, imageName string) (bool, error) {
	log.Infof("Checking image %s for socket files...", imageName)

	// Inspect the image to get GraphDriver data
	imageInspect, _, err := d.apiClient.ImageInspectWithRaw(ctx, imageName)
	if err != nil {
		return false, fmt.Errorf("failed to inspect image %s: %w", imageName, err)
	}

	// Check if GraphDriver data exists
	if imageInspect.GraphDriver.Data == nil {
		log.Infof("No GraphDriver data found for image %s", imageName)
		return false, nil
	}

	// Get LowerDir and UpperDir from GraphDriver data
	lowerDir, ok := imageInspect.GraphDriver.Data["LowerDir"]
	if !ok {
		log.Infof("No LowerDir found for image %s", imageName)
	}
	upperDir, ok := imageInspect.GraphDriver.Data["UpperDir"]
	if !ok {
		log.Infof("No UpperDir found for image %s", imageName)
	}

	var dirsToCheck []string

	// Add LowerDir directories if it exists (split by colons)
	if lowerDir != "" {
		lowerDirs := strings.Split(lowerDir, ":")
		dirsToCheck = append(dirsToCheck, lowerDirs...)
	}

	// Add UpperDir if it exists
	if upperDir != "" {
		dirsToCheck = append(dirsToCheck, upperDir)
	}

	// Check each directory for socket files
	for _, dir := range dirsToCheck {
		hasSockets, err := d.checkDirectoryForSockets(dir)
		if err != nil {
			log.Warnf("Failed to check directory %s for sockets: %v", dir, err)
			continue
		}

		if hasSockets {
			log.Infof("Found socket files in directory: %s", dir)
			return true, nil
		}
	}

	log.Infof("No socket files found in image %s", imageName)
	return false, nil
}

// checkDirectoryForSockets uses the find command to search for socket files in a directory
func (d *DockerClient) checkDirectoryForSockets(dirPath string) (bool, error) {
	// Use find command to search for socket files
	cmd := exec.Command("find", dirPath, "-type", "s")
	output, err := cmd.Output()

	if err != nil {
		// If find command fails, it might be because no socket files were found
		// Check the exit code to determine if it's an error or just no results
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			// Exit code 1 means no files found, which is not an error
			return false, nil
		}
		return false, fmt.Errorf("find command failed: %w", err)
	}

	// If output is not empty, socket files were found
	socketFiles := strings.TrimSpace(string(output))
	if socketFiles != "" {
		log.Infof("Found socket files in %s: %s", dirPath, socketFiles)
		return true, nil
	}

	return false, nil
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
