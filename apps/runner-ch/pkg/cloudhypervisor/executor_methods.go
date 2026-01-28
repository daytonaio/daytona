// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package cloudhypervisor

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/daytonaio/runner-ch/internal"
	"github.com/daytonaio/runner-ch/pkg/api/dto"
	log "github.com/sirupsen/logrus"
)

// Create creates a new VM sandbox using CreateSandboxDTO
// Returns (SandboxInfo, daemonVersion, error)
func (c *Client) Create(ctx context.Context, createDto dto.CreateSandboxDTO) (*SandboxInfo, string, error) {
	opts := CreateOptions{
		SandboxId:  createDto.Id,
		Cpus:       int(createDto.CpuQuota),
		MemoryMB:   uint64(createDto.MemoryQuota * 1024), // MemoryQuota is in GB, convert to MB
		StorageGB:  int(createDto.StorageQuota),
		Snapshot:   createDto.Snapshot,
		GpuDevices: createDto.GpuDevices,
		Metadata:   createDto.Metadata,
	}

	info, err := c.CreateWithOptions(ctx, opts)
	if err != nil {
		return nil, "", err
	}

	return info, internal.Version, nil
}

// Start starts a VM and returns daemon version
func (c *Client) Start(ctx context.Context, sandboxId string, metadata map[string]string) (string, error) {
	err := c.StartVM(ctx, sandboxId)
	if err != nil {
		return "", err
	}
	return internal.Version, nil
}

// UpdateNetworkSettings updates network settings for a sandbox
func (c *Client) UpdateNetworkSettings(ctx context.Context, sandboxId string, settings dto.UpdateNetworkSettingsDTO) error {
	log.Infof("Updating network settings for sandbox %s", sandboxId)
	// TODO: Implement network settings update
	// This would configure iptables rules for network blocking/allowing
	return nil
}

// CreateBackup creates a backup of a sandbox
func (c *Client) CreateBackup(ctx context.Context, sandboxId string, backupDto dto.CreateBackupDTO) error {
	log.Infof("Creating backup for sandbox %s with snapshot %s", sandboxId, backupDto.Snapshot)
	// TODO: Implement backup creation
	// This would snapshot the VM and push to registry
	return nil
}

// BuildSnapshot builds a snapshot from a Dockerfile
func (c *Client) BuildSnapshot(ctx context.Context, request dto.BuildSnapshotRequestDTO) error {
	log.Infof("Building snapshot %s from dockerfile", request.Ref)
	// TODO: Implement snapshot building
	// For Cloud Hypervisor, this might involve:
	// 1. Building a container image from the Dockerfile
	// 2. Converting it to a VM disk image
	return nil
}

// PullSnapshot pulls a snapshot from a registry
// PullSnapshot downloads a snapshot from S3 to the local runner storage
// The ref format is: {orgId}/{snapshotName}
func (c *Client) PullSnapshot(ctx context.Context, request dto.PullSnapshotRequestDTO) error {
	log.Infof("Pulling snapshot %s", request.Ref)

	// Check if S3 is configured
	if c.s3Uploader == nil || !c.s3Uploader.IsConfigured() {
		return fmt.Errorf("S3 is not configured - cannot pull snapshot")
	}

	// Parse the ref to get orgId and snapshotName
	// Expected format: {orgId}/{snapshotName}
	orgId, snapshotName := parseSnapshotRef(request.Ref)
	if orgId == "" || snapshotName == "" {
		return fmt.Errorf("invalid snapshot ref format '%s', expected {orgId}/{snapshotName}", request.Ref)
	}

	// Check if snapshot already exists locally
	localPath := filepath.Join(c.config.SnapshotsPath, request.Ref)
	diskPath := filepath.Join(localPath, "disk.qcow2")
	if exists, _ := c.fileExists(ctx, diskPath); exists {
		log.Infof("Snapshot %s already exists locally at %s", request.Ref, localPath)
		return nil
	}

	// Download from S3
	result, err := c.s3Uploader.DownloadSnapshot(ctx, c.config.SnapshotsPath, orgId, snapshotName)
	if err != nil {
		return fmt.Errorf("failed to download snapshot from S3: %w", err)
	}

	log.Infof("Successfully pulled snapshot %s: %d files, %d bytes in %v",
		request.Ref, result.FileCount, result.TotalSize, result.Duration)

	return nil
}

// ImageInfo contains information about a disk image
type ImageInfo struct {
	Size       int64    // Size in bytes
	Entrypoint []string // Entry point commands
	Cmd        []string // Default command
	Hash       string   // Content hash
}

// GetImageInfo returns information about a disk image/snapshot
func (c *Client) GetImageInfo(ctx context.Context, snapshot string) (*ImageInfo, error) {
	log.Infof("Getting image info for %s", snapshot)

	// Snapshots are directories containing disk.qcow2
	snapshotPath := filepath.Join(c.config.SnapshotsPath, snapshot, "disk.qcow2")

	// Get file size via qemu-img (validate it exists and is valid)
	_, err := c.runCommandOutput(ctx, "qemu-img", "info", "--output=json", snapshotPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get image info: %w", err)
	}

	// Get file size
	sizeOutput, err := c.runCommandOutput(ctx, "stat", "-c", "%s", snapshotPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file size: %w", err)
	}

	var size int64
	fmt.Sscanf(sizeOutput, "%d", &size)

	// Generate a hash from the file
	hashOutput, err := c.runCommandOutput(ctx, "sha256sum", snapshotPath)
	if err != nil {
		log.Warnf("Failed to calculate hash: %v", err)
		hashOutput = "unknown"
	}

	hash := "sha256:"
	if len(hashOutput) >= 64 {
		hash += hashOutput[:64]
	}

	return &ImageInfo{
		Size:       size,
		Entrypoint: []string{}, // N/A for disk images
		Cmd:        []string{}, // N/A for disk images
		Hash:       hash,
	}, nil
}

// RemoveImage removes a snapshot/image directory and its S3 artifacts
// The ref format is: {orgId}/{snapshotName}
func (c *Client) RemoveImage(ctx context.Context, ref string, force bool) error {
	log.Infof("Removing image %s (force=%v)", ref, force)

	// Parse the ref to get orgId and snapshotName
	// Expected format: {orgId}/{snapshotName}
	orgId, snapshotName := parseSnapshotRef(ref)

	// Delete from S3 if configured
	if c.s3Uploader != nil && c.s3Uploader.IsConfigured() {
		if orgId != "" && snapshotName != "" {
			log.Infof("Deleting snapshot from S3: orgId=%s, snapshotName=%s", orgId, snapshotName)
			if err := c.s3Uploader.DeleteSnapshot(ctx, orgId, snapshotName); err != nil {
				log.Warnf("Failed to delete snapshot from S3 (will continue with local cleanup): %v", err)
				// Continue with local cleanup even if S3 deletion fails
			} else {
				log.Infof("Successfully deleted snapshot from S3")
			}
		} else {
			log.Warnf("Could not parse orgId/snapshotName from ref '%s', skipping S3 deletion", ref)
		}
	}

	// Local snapshot directory is at: {snapshotsPath}/{orgId}/{snapshotName}
	snapshotDir := filepath.Join(c.config.SnapshotsPath, ref)

	// Remove the entire snapshot directory (local cache)
	return c.runCommand(ctx, "rm", "-rf", snapshotDir)
}

// parseSnapshotRef extracts organizationId and snapshotName from a snapshot ref
// Expected format: {orgId}/{snapshotName}
// Returns empty strings if the ref doesn't match the expected format
func parseSnapshotRef(ref string) (orgId, snapshotName string) {
	// Split by "/"
	parts := splitPath(ref)

	// We need exactly orgId and snapshotName (2 parts)
	if len(parts) >= 2 {
		return parts[0], parts[1]
	}

	return "", ""
}

// splitPath splits a path by "/" separator
func splitPath(path string) []string {
	var parts []string
	start := 0
	for i := 0; i <= len(path); i++ {
		if i == len(path) || path[i] == '/' {
			if i > start {
				parts = append(parts, path[start:i])
			}
			start = i + 1
		}
	}
	return parts
}

// PushSnapshot pushes a snapshot to a registry
func (c *Client) PushSnapshot(ctx context.Context, request dto.PushSnapshotRequestDTO) (any, error) {
	log.Infof("Pushing snapshot %s", request.Ref)
	// TODO: Implement snapshot pushing to registry
	return nil, nil
}

// CreateSnapshot creates a snapshot of a sandbox and uploads it to S3
func (c *Client) CreateSnapshot(ctx context.Context, request dto.CreateSnapshotRequestDTO) (*dto.CreateSnapshotResponseDTO, error) {
	log.Infof("Creating snapshot '%s' from sandbox '%s' (live=%v, org=%s)", request.Name, request.SandboxId, request.Live, request.OrganizationId)

	// Use the internal CreateSnapshotFromVM with proper options
	// The local path will be: /var/lib/cloud-hypervisor/snapshots/{orgId}/{name}
	snapshotPath, err := c.CreateSnapshotFromVM(ctx, SnapshotOptions{
		SandboxId:      request.SandboxId,
		Name:           request.Name,
		OrganizationId: request.OrganizationId,
	})
	if err != nil {
		return nil, err
	}

	// The snapshot ref is in format: {orgId}/{name}
	// This is used for both local storage and S3
	snapshotRef := filepath.Join(request.OrganizationId, request.Name)

	// Get size of the snapshot (check qcow2 first, then legacy raw)
	var sizeBytes int64
	diskPath := filepath.Join(snapshotPath, "disk.qcow2")
	if exists, _ := c.fileExists(ctx, diskPath); !exists {
		diskPath = filepath.Join(snapshotPath, "disk.raw")
	}
	sizeOutput, err := c.runCommandOutput(ctx, "stat", "-c", "%s", diskPath)
	if err == nil {
		fmt.Sscanf(sizeOutput, "%d", &sizeBytes)
	}

	response := &dto.CreateSnapshotResponseDTO{
		Name:         request.Name,
		SnapshotPath: snapshotRef, // Return the ref format, not the full local path
		SizeBytes:    sizeBytes,
		LiveMode:     request.Live,
	}

	// Upload to S3 if configured
	if c.s3Uploader != nil && c.s3Uploader.IsConfigured() {
		log.Infof("Uploading snapshot '%s' to S3 (org: %s)", request.Name, request.OrganizationId)

		uploadResult, err := c.s3Uploader.UploadSnapshot(ctx, snapshotPath, request.OrganizationId, request.Name)
		if err != nil {
			log.Errorf("Failed to upload snapshot to S3: %v", err)
			return nil, fmt.Errorf("snapshot created locally but failed to upload to S3: %w", err)
		}

		response.S3Path = uploadResult.S3Path
		response.SizeBytes = uploadResult.TotalSize // Update with actual uploaded size
		log.Infof("Snapshot uploaded to S3: %s (%d bytes in %v)",
			uploadResult.S3Path, uploadResult.TotalSize, uploadResult.Duration)
	} else {
		log.Warn("S3 not configured - snapshot stored locally only")
	}

	return response, nil
}
