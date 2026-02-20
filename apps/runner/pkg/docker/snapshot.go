// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"

	"github.com/daytonaio/runner/pkg/api/dto"

	log "github.com/sirupsen/logrus"
)

// CreateSnapshot creates a snapshot from a container by committing it,
// tagging with the organization ID and snapshot name, and pushing to the registry.
func (d *DockerClient) CreateSnapshot(ctx context.Context, snapshotDto dto.CreateSnapshotDTO) (*dto.CreateSnapshotResponseDTO, error) {
	log.Infof("Creating snapshot '%s' for container %s (org: %s)...", snapshotDto.Name, snapshotDto.SandboxId, snapshotDto.OrganizationId)

	// Check if registry is provided
	if snapshotDto.Registry == nil {
		return nil, fmt.Errorf("registry is required for creating snapshot")
	}

	// Build the snapshot path: {registryUrl}/{project}/{orgId}/{name}:latest
	var snapshotPath string
	if snapshotDto.Registry.Project != nil && *snapshotDto.Registry.Project != "" {
		snapshotPath = fmt.Sprintf("%s/%s/%s/%s:latest",
			snapshotDto.Registry.Url,
			*snapshotDto.Registry.Project,
			snapshotDto.OrganizationId,
			snapshotDto.Name)
	} else {
		snapshotPath = fmt.Sprintf("%s/%s/%s:latest",
			snapshotDto.Registry.Url,
			snapshotDto.OrganizationId,
			snapshotDto.Name)
	}

	// Use a temporary local image name for the commit
	localImageName := fmt.Sprintf("snapshot-%s-%s:latest", snapshotDto.OrganizationId, snapshotDto.Name)

	// Commit the container to a local image
	err := d.commitContainer(ctx, snapshotDto.SandboxId, localImageName)
	if err != nil {
		log.Errorf("Error committing container %s: %v", snapshotDto.SandboxId, err)
		return nil, fmt.Errorf("failed to commit container: %w", err)
	}

	// Tag the local image with the registry reference
	err = d.TagImage(ctx, localImageName, snapshotPath)
	if err != nil {
		log.Errorf("Error tagging image %s as %s: %v", localImageName, snapshotPath, err)
		// Clean up local image on failure
		_ = d.RemoveImage(ctx, localImageName, true)
		return nil, fmt.Errorf("failed to tag image: %w", err)
	}

	// Push the tagged image to the registry
	err = d.PushImage(ctx, snapshotPath, snapshotDto.Registry)
	if err != nil {
		log.Errorf("Error pushing image %s: %v", snapshotPath, err)
		// Clean up local images on failure
		_ = d.RemoveImage(ctx, snapshotPath, true)
		_ = d.RemoveImage(ctx, localImageName, true)
		return nil, fmt.Errorf("failed to push image: %w", err)
	}

	log.Infof("Snapshot '%s' created and pushed successfully as %s", snapshotDto.Name, snapshotPath)

	// Get image info for the response
	imageInfo, err := d.GetImageInfo(ctx, snapshotPath)
	if err != nil {
		log.Warnf("Failed to get image info for %s: %v", snapshotPath, err)
		// Continue - we can still return partial response
	}

	// Remove intermediate commit tag; keep the registry-tagged image locally
	// so the runner can use it for sandbox creation without re-pulling
	err = d.RemoveImage(ctx, localImageName, true)
	if err != nil {
		log.Warnf("Error removing intermediate image %s: %v", localImageName, err)
	}

	response := &dto.CreateSnapshotResponseDTO{
		Name:         snapshotDto.Name,
		SnapshotPath: snapshotPath,
	}

	if imageInfo != nil {
		response.SizeGB = float64(imageInfo.Size) / (1024 * 1024 * 1024) // Convert bytes to GB
		response.SizeBytes = imageInfo.Size
		response.Hash = dto.HashWithoutPrefix(imageInfo.Hash)
	}

	return response, nil
}
