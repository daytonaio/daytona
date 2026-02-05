// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package cuttlefish

import (
	"context"
	"fmt"

	"github.com/daytonaio/runner-android/internal"
	"github.com/daytonaio/runner-android/pkg/api/dto"
	log "github.com/sirupsen/logrus"
)

// Create creates a new Cuttlefish instance using CreateSandboxDTO
// Returns (SandboxInfo, daemonVersion, error)
func (c *Client) Create(ctx context.Context, createDto dto.CreateSandboxDTO) (*SandboxInfo, string, error) {
	opts := CreateOptions{
		SandboxId: createDto.Id,
		Cpus:      int(createDto.CpuQuota),
		MemoryMB:  uint64(createDto.MemoryQuota * 1024), // MemoryQuota is in GB, convert to MB
		DiskGB:    int(createDto.StorageQuota),
		Snapshot:  createDto.Snapshot,
		Metadata:  createDto.Metadata,
	}

	info, err := c.CreateWithOptions(ctx, opts)
	if err != nil {
		return nil, "", err
	}

	return info, internal.Version, nil
}

// Start starts a Cuttlefish instance and returns daemon version
func (c *Client) Start(ctx context.Context, sandboxId string, metadata map[string]string) (string, error) {
	err := c.StartVM(ctx, sandboxId)
	if err != nil {
		return "", err
	}
	return internal.Version, nil
}

// UpdateNetworkSettings updates network settings for a sandbox
// Note: This is a no-op for Cuttlefish as it manages its own networking
func (c *Client) UpdateNetworkSettings(ctx context.Context, sandboxId string, settings dto.UpdateNetworkSettingsDTO) error {
	log.Infof("Updating network settings for sandbox %s (no-op for Cuttlefish)", sandboxId)
	return nil
}

// CreateBackup creates a backup of a sandbox
// For Cuttlefish, this creates a snapshot that can be used to create new instances
func (c *Client) CreateBackup(ctx context.Context, sandboxId string, backupDto dto.CreateBackupDTO) error {
	log.Infof("CreateBackup for Cuttlefish sandbox %s (use CreateSnapshot instead)", sandboxId)
	return nil
}

// BuildSnapshot builds a snapshot from a Dockerfile
// Note: Not applicable for Cuttlefish (uses Android system images)
func (c *Client) BuildSnapshot(ctx context.Context, request dto.BuildSnapshotRequestDTO) error {
	log.Infof("BuildSnapshot not supported for Cuttlefish (uses Android system images)")
	return nil
}

// PullSnapshot pulls a snapshot/system image
// Note: For Cuttlefish, this would involve downloading Android images
func (c *Client) PullSnapshot(ctx context.Context, request dto.PullSnapshotRequestDTO) error {
	log.Infof("PullSnapshot: %s (not yet implemented for Cuttlefish)", request.Snapshot)
	return nil
}

// GetImageInfo returns information about a system image/snapshot
func (c *Client) GetImageInfo(ctx context.Context, snapshot string) (*ImageInfo, error) {
	log.Infof("Getting image info for %s", snapshot)

	info, err := c.GetSnapshotInfo(ctx, snapshot)
	if err != nil {
		return &ImageInfo{
			Size:       0,
			Entrypoint: []string{},
			Cmd:        []string{},
			Hash:       "unknown",
		}, nil
	}

	return &ImageInfo{
		Size:       info.Metadata.SizeBytes,
		Entrypoint: []string{},
		Cmd:        []string{},
		Hash:       fmt.Sprintf("%s-%d", info.Path, info.Metadata.CreatedAt.Unix()),
	}, nil
}

// ImageInfo contains information about a system image
type ImageInfo struct {
	Size       int64
	Entrypoint []string
	Cmd        []string
	Hash       string
}

// RemoveImage removes a snapshot
func (c *Client) RemoveImage(ctx context.Context, ref string, force bool) error {
	log.Infof("RemoveImage: %s", ref)
	return c.DeleteSnapshot(ctx, ref)
}

// PushSnapshot pushes a snapshot to a registry
func (c *Client) PushSnapshot(ctx context.Context, request dto.PushSnapshotRequestDTO) (any, error) {
	log.Infof("PushSnapshot: %s (not implemented for Cuttlefish)", request.Ref)
	return nil, nil
}

// CreateSnapshot creates a snapshot of a running sandbox
// The snapshot can be used to create new sandboxes with the same state
func (c *Client) CreateSnapshot(ctx context.Context, request dto.CreateSnapshotRequestDTO) (*dto.CreateSnapshotResponseDTO, error) {
	log.Infof("Creating snapshot '%s' from sandbox %s for org %s", request.Name, request.SandboxId, request.OrganizationId)

	// Extract org ID from request or metadata
	orgId := request.OrganizationId
	if orgId == "" {
		// Try to get org ID from sandbox metadata
		info, exists := c.GetInstance(request.SandboxId)
		if exists && info.Metadata != nil {
			orgId = info.Metadata["orgId"]
		}
	}

	if orgId == "" {
		return nil, fmt.Errorf("organization ID is required for creating custom snapshots")
	}

	// Create the snapshot (description is empty since the DTO doesn't have it)
	snapshotInfo, err := c.CreateSnapshotFromInstance(ctx, request.SandboxId, orgId, request.Name, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot: %w", err)
	}

	return &dto.CreateSnapshotResponseDTO{
		Name:         snapshotInfo.Metadata.Name,
		SnapshotPath: snapshotInfo.Path,
		SizeBytes:    snapshotInfo.Metadata.SizeBytes,
		LiveMode:     request.Live,
	}, nil
}

// RecoverOrphanedSandboxes recovers sandboxes that may have been orphaned
func (c *Client) RecoverOrphanedSandboxes(ctx context.Context) error {
	log.Info("Checking for orphaned Cuttlefish instances...")

	// Reload instance mappings
	if err := c.loadInstanceMappings(); err != nil {
		log.Warnf("Failed to reload instance mappings: %v", err)
	}

	// Check each registered instance
	c.mutex.RLock()
	instances := make([]*InstanceInfo, 0, len(c.instances))
	for _, info := range c.instances {
		instances = append(instances, info)
	}
	c.mutex.RUnlock()

	recovered := 0
	for _, info := range instances {
		state := c.getInstanceState(ctx, info.InstanceNum)
		c.mutex.Lock()
		info.State = state
		c.mutex.Unlock()

		if state == InstanceStateRunning {
			log.Infof("Instance %d (%s) is running", info.InstanceNum, info.SandboxId)
		} else {
			log.Infof("Instance %d (%s) is stopped", info.InstanceNum, info.SandboxId)
		}
		recovered++
	}

	log.Infof("Recovery check completed: %d instances found", recovered)
	return nil
}

// InitializeIPPool is a no-op for Cuttlefish (it manages its own IPs)
func (c *Client) InitializeIPPool(ctx context.Context) error {
	return nil
}

// InitializeNetNSPool is a no-op for Cuttlefish
func (c *Client) InitializeNetNSPool(ctx context.Context) error {
	return nil
}

// ForkVM is not supported for Cuttlefish
func (c *Client) ForkVM(ctx context.Context, opts ForkOptions) (*SandboxInfo, error) {
	log.Warn("ForkVM not supported for Cuttlefish")
	return nil, nil
}

// ForkOptions for ForkVM (not supported)
type ForkOptions struct {
	SourceSandboxId string
	NewSandboxId    string
	SourceStopped   bool
}

// CloneVM is not supported for Cuttlefish
func (c *Client) CloneVM(ctx context.Context, opts CloneOptions) (*SandboxInfo, error) {
	log.Warn("CloneVM not supported for Cuttlefish")
	return nil, nil
}

// CloneOptions for CloneVM (not supported)
type CloneOptions struct {
	SourceSandboxId string
	NewSandboxId    string
}
