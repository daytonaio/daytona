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
		MemoryMB:   uint64(createDto.MemoryQuota),
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
func (c *Client) PullSnapshot(ctx context.Context, request dto.PullSnapshotRequestDTO) error {
	log.Infof("Pulling snapshot %s", request.Ref)
	// TODO: Implement snapshot pulling
	// This would download the snapshot image to the snapshots directory
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

	// Find the snapshot file
	snapshotPath := filepath.Join(c.config.SnapshotsPath, snapshot)
	if !hasImageExtension(snapshotPath) {
		snapshotPath += ".qcow2"
	}

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

// RemoveImage removes a snapshot/image
func (c *Client) RemoveImage(ctx context.Context, ref string, force bool) error {
	log.Infof("Removing image %s (force=%v)", ref, force)

	snapshotPath := filepath.Join(c.config.SnapshotsPath, ref)
	if !hasImageExtension(snapshotPath) {
		// Try common extensions
		for _, ext := range []string{".qcow2", ".raw", ".img"} {
			testPath := snapshotPath + ext
			exists, _ := c.fileExists(ctx, testPath)
			if exists {
				snapshotPath = testPath
				break
			}
		}
	}

	return c.runCommand(ctx, "rm", "-f", snapshotPath)
}

// PushSnapshot pushes a snapshot to a registry
func (c *Client) PushSnapshot(ctx context.Context, request dto.PushSnapshotRequestDTO) (any, error) {
	log.Infof("Pushing snapshot %s", request.Ref)
	// TODO: Implement snapshot pushing to registry
	return nil, nil
}

// CreateSnapshot creates a snapshot of a sandbox
func (c *Client) CreateSnapshot(ctx context.Context, request dto.CreateSnapshotRequestDTO) (*dto.CreateSnapshotResponseDTO, error) {
	log.Infof("Creating snapshot '%s' from sandbox '%s' (live=%v)", request.Name, request.SandboxId, request.Live)

	// Use the internal CreateSnapshotFromVM with proper options
	snapshotPath, err := c.CreateSnapshotFromVM(ctx, SnapshotOptions{
		SandboxId: request.SandboxId,
		Name:      request.Name,
	})
	if err != nil {
		return nil, err
	}

	// Get size of the snapshot
	var sizeBytes int64
	diskPath := filepath.Join(snapshotPath, "disk.raw")
	sizeOutput, err := c.runCommandOutput(ctx, "stat", "-c", "%s", diskPath)
	if err == nil {
		fmt.Sscanf(sizeOutput, "%d", &sizeBytes)
	}

	return &dto.CreateSnapshotResponseDTO{
		Name:         request.Name,
		SnapshotPath: snapshotPath,
		SizeBytes:    sizeBytes,
		LiveMode:     request.Live,
	}, nil
}
