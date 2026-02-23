// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package cloudhypervisor

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"
)

// SnapshotOptions specifies options for creating a snapshot
type SnapshotOptions struct {
	SandboxId      string // Source sandbox to snapshot
	Name           string // Snapshot name (defaults to sandboxId-timestamp)
	OrganizationId string // Organization ID for namespacing (required to avoid conflicts)
}

// CreateSnapshotFromVM creates a snapshot of a running or paused VM
// This captures both memory state and disk state for instant restoration
func (c *Client) CreateSnapshotFromVM(ctx context.Context, opts SnapshotOptions) (string, error) {
	mutex := c.getSandboxMutex(opts.SandboxId)
	mutex.Lock()
	defer mutex.Unlock()

	// Generate snapshot name if not provided
	if opts.Name == "" {
		opts.Name = fmt.Sprintf("%s-%d", opts.SandboxId, time.Now().Unix())
	}

	// Use orgId/name format for the snapshot path to avoid naming conflicts between organizations
	// The snapshotRef will be: {orgId}/{name}
	var snapshotRef string
	if opts.OrganizationId != "" {
		snapshotRef = filepath.Join(opts.OrganizationId, opts.Name)
	} else {
		snapshotRef = opts.Name
	}
	snapshotPath := filepath.Join(c.config.SnapshotsPath, snapshotRef)

	log.Infof("Creating snapshot %s from sandbox %s", opts.Name, opts.SandboxId)

	// Get current state
	info, err := c.GetInfo(ctx, opts.SandboxId)
	if err != nil {
		return "", fmt.Errorf("failed to get VM info: %w", err)
	}

	// VM must be running or paused to snapshot
	if info.State != VmStateRunning && info.State != VmStatePaused {
		return "", fmt.Errorf("VM must be running or paused to snapshot (current state: %s)", info.State)
	}

	// If running, pause first for consistent snapshot
	wasPaused := info.State == VmStatePaused
	if !wasPaused {
		log.Infof("Pausing VM for consistent snapshot")
		if _, err := c.apiRequest(ctx, opts.SandboxId, http.MethodPut, "vm.pause", nil); err != nil {
			return "", fmt.Errorf("failed to pause VM for snapshot: %w", err)
		}
	}

	// Ensure snapshots directory exists
	if err := c.runCommand(ctx, "mkdir", "-p", c.config.SnapshotsPath); err != nil {
		return "", fmt.Errorf("failed to create snapshots directory: %w", err)
	}

	// Create snapshot directory
	if err := c.runCommand(ctx, "mkdir", "-p", snapshotPath); err != nil {
		return "", fmt.Errorf("failed to create snapshot directory: %w", err)
	}

	// Create snapshot via CH API
	// The destination_url is a file:// URL pointing to the snapshot directory
	snapshotConfig := map[string]string{
		"destination_url": fmt.Sprintf("file://%s", snapshotPath),
	}

	if _, err := c.apiRequest(ctx, opts.SandboxId, http.MethodPut, "vm.snapshot", snapshotConfig); err != nil {
		// Cleanup on failure
		_ = c.runCommand(ctx, "rm", "-rf", snapshotPath)
		// Resume if we paused
		if !wasPaused {
			_, _ = c.apiRequest(ctx, opts.SandboxId, http.MethodPut, "vm.resume", nil)
		}
		return "", fmt.Errorf("failed to create snapshot: %w", err)
	}

	// Flatten disk image to snapshot WHILE VM IS STILL PAUSED
	// Critical: using plain 'cp' would preserve the backing file reference,
	// creating a chain dependency that causes I/O errors during restore.
	// qemu-img convert flattens all layers into a standalone image.
	//
	// Two-step process to bypass Cloud Hypervisor's file lock:
	// 1. Copy the qcow2 overlay file (cp ignores file locks)
	// 2. Convert the copy to flatten the backing chain
	// 3. Remove the temporary copy
	diskPath := c.getDiskPath(opts.SandboxId)
	snapshotDiskPath := filepath.Join(snapshotPath, "disk.qcow2")
	tempDiskPath := filepath.Join(snapshotPath, "disk.qcow2.tmp")

	log.Infof("Flattening disk to snapshot (this may take a moment for large disks)")

	// Step 1: Copy the overlay file (bypasses CH's file lock)
	copyCmd := fmt.Sprintf("cp '%s' '%s'", diskPath, tempDiskPath)
	if err := c.runCommand(ctx, "sh", "-c", copyCmd); err != nil {
		log.Errorf("Failed to copy disk overlay: %v", err)
		_ = c.runCommand(ctx, "rm", "-rf", snapshotPath)
		if !wasPaused {
			_, _ = c.apiRequest(ctx, opts.SandboxId, http.MethodPut, "vm.resume", nil)
		}
		return "", fmt.Errorf("failed to copy disk overlay: %w", err)
	}

	// Step 2: Convert the copy to flatten the backing chain
	convertCmd := fmt.Sprintf("qemu-img convert -O qcow2 '%s' '%s' && rm -f '%s'", tempDiskPath, snapshotDiskPath, tempDiskPath)
	if err := c.runCommand(ctx, "sh", "-c", convertCmd); err != nil {
		// Disk copy is required - cleanup and return error
		log.Errorf("Failed to flatten disk to snapshot: %v", err)
		_ = c.runCommand(ctx, "rm", "-rf", snapshotPath)
		// Resume VM before returning error
		if !wasPaused {
			_, _ = c.apiRequest(ctx, opts.SandboxId, http.MethodPut, "vm.resume", nil)
		}
		return "", fmt.Errorf("failed to flatten disk to snapshot: %w", err)
	}

	// Resume VM after disk flatten is complete
	if !wasPaused {
		log.Infof("Resuming VM after snapshot")
		if _, err := c.apiRequest(ctx, opts.SandboxId, http.MethodPut, "vm.resume", nil); err != nil {
			log.Warnf("Failed to resume VM after snapshot: %v", err)
		}
	}

	log.Infof("Snapshot %s created successfully at %s", opts.Name, snapshotPath)
	return snapshotPath, nil
}

// RestoreOptions specifies options for restoring from a snapshot
type RestoreOptions struct {
	SnapshotPath string // Path to snapshot (or snapshot name)
	SandboxId    string // New sandbox ID (required)
	Prefault     bool   // Prefault memory pages for faster access
}

// Restore creates a new VM from a snapshot (live fork)
// This is the fast path for creating new sandboxes from a base image
func (c *Client) Restore(ctx context.Context, opts RestoreOptions) (*SandboxInfo, error) {
	mutex := c.getSandboxMutex(opts.SandboxId)
	mutex.Lock()
	defer mutex.Unlock()

	// Resolve snapshot path
	snapshotPath := opts.SnapshotPath
	if !filepath.IsAbs(snapshotPath) {
		snapshotPath = filepath.Join(c.config.SnapshotsPath, snapshotPath)
	}

	log.Infof("Restoring sandbox %s from snapshot %s", opts.SandboxId, snapshotPath)

	// Check if snapshot exists
	exists, err := c.fileExists(ctx, snapshotPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check snapshot existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("snapshot not found: %s", snapshotPath)
	}

	// Create sandbox directory
	sandboxDir := c.getSandboxDir(opts.SandboxId)
	if err := c.runCommand(ctx, "mkdir", "-p", sandboxDir); err != nil {
		return nil, fmt.Errorf("failed to create sandbox directory: %w", err)
	}

	// Copy disk from snapshot (check both qcow2 and legacy raw format)
	snapshotDiskPath := filepath.Join(snapshotPath, "disk.qcow2")
	if exists, _ := c.fileExists(ctx, snapshotDiskPath); !exists {
		// Fall back to legacy raw format
		snapshotDiskPath = filepath.Join(snapshotPath, "disk.raw")
	}
	diskPath := c.getDiskPath(opts.SandboxId)

	diskExists, _ := c.fileExists(ctx, snapshotDiskPath)
	if diskExists {
		log.Infof("Copying disk from snapshot")
		if err := c.runCommand(ctx, "cp", snapshotDiskPath, diskPath); err != nil {
			c.cleanupSandbox(ctx, opts.SandboxId)
			return nil, fmt.Errorf("failed to copy disk from snapshot: %w", err)
		}
	}

	// Get TAP interface (from pool if enabled, otherwise create)
	var tapName string
	if c.tapPool.IsEnabled() {
		var err error
		tapName, err = c.tapPool.Acquire(ctx, opts.SandboxId)
		if err != nil {
			c.cleanupSandbox(ctx, opts.SandboxId)
			return nil, fmt.Errorf("failed to acquire TAP from pool: %w", err)
		}
		log.Infof("Acquired TAP %s from pool for fork %s", tapName, opts.SandboxId)
	} else {
		tapName = c.getTapName(opts.SandboxId)
		if err := c.createTapInterface(ctx, tapName); err != nil {
			c.cleanupSandbox(ctx, opts.SandboxId)
			return nil, fmt.Errorf("failed to create TAP interface: %w", err)
		}
	}

	// Start cloud-hypervisor process
	if err := c.startVMProcess(ctx, opts.SandboxId); err != nil {
		c.cleanupSandbox(ctx, opts.SandboxId)
		return nil, fmt.Errorf("failed to start VM process: %w", err)
	}

	// Wait for socket
	if err := c.waitForSocket(ctx, opts.SandboxId, 30*time.Second); err != nil {
		c.cleanupSandbox(ctx, opts.SandboxId)
		return nil, fmt.Errorf("failed to wait for socket: %w", err)
	}

	// Build restore config
	restoreConfig := RestoreConfig{
		SourceUrl: fmt.Sprintf("file://%s", snapshotPath),
		Prefault:  opts.Prefault,
	}

	// Need to provide net_fds for the new TAP device
	// This is done through the config, but CH might need the TAP to be set up differently
	// For now, we'll try without net_fds and see if CH can figure it out

	log.Infof("Restoring VM from snapshot (prefault=%v)", opts.Prefault)

	// Call restore API
	if _, err := c.apiRequest(ctx, opts.SandboxId, http.MethodPut, "vm.restore", restoreConfig); err != nil {
		c.cleanupSandbox(ctx, opts.SandboxId)
		return nil, fmt.Errorf("failed to restore VM: %w", err)
	}

	// Wait for running state
	if err := c.waitForState(ctx, opts.SandboxId, VmStateRunning, 60*time.Second); err != nil {
		// Try to resume if it restored as paused
		info, _ := c.GetInfo(ctx, opts.SandboxId)
		if info != nil && info.State == VmStatePaused {
			log.Infof("VM restored as paused, resuming")
			if _, err := c.apiRequest(ctx, opts.SandboxId, http.MethodPut, "vm.resume", nil); err != nil {
				log.Warnf("Failed to resume after restore: %v", err)
			}
		}
	}

	log.Infof("Sandbox %s restored successfully from snapshot", opts.SandboxId)
	return c.GetSandboxInfo(ctx, opts.SandboxId)
}

// Fork creates a new VM as a copy of an existing running/paused VM
// This is a convenience method that snapshots and restores in one operation
func (c *Client) Fork(ctx context.Context, sourceSandboxId, newSandboxId string) (*SandboxInfo, error) {
	log.Infof("Forking sandbox %s to %s", sourceSandboxId, newSandboxId)

	// Create a temporary snapshot
	snapshotName := fmt.Sprintf("fork-%s-%d", sourceSandboxId, time.Now().UnixNano())
	snapshotPath, err := c.CreateSnapshotFromVM(ctx, SnapshotOptions{
		SandboxId: sourceSandboxId,
		Name:      snapshotName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot for fork: %w", err)
	}

	// Restore from snapshot
	info, err := c.Restore(ctx, RestoreOptions{
		SnapshotPath: snapshotPath,
		SandboxId:    newSandboxId,
		Prefault:     true, // Prefault for faster access
	})
	if err != nil {
		// Cleanup snapshot on failure
		_ = c.DeleteSnapshot(ctx, snapshotName)
		return nil, fmt.Errorf("failed to restore from snapshot: %w", err)
	}

	// Optionally delete the temporary snapshot
	// (keeping it would allow multiple forks from the same point)
	// _ = c.DeleteSnapshot(ctx, snapshotName)

	log.Infof("Fork completed: %s -> %s", sourceSandboxId, newSandboxId)
	return info, nil
}

// DeleteSnapshot removes a snapshot
func (c *Client) DeleteSnapshot(ctx context.Context, name string) error {
	snapshotPath := filepath.Join(c.config.SnapshotsPath, name)

	log.Infof("Deleting snapshot %s", name)

	if err := c.runCommand(ctx, "rm", "-rf", snapshotPath); err != nil {
		return fmt.Errorf("failed to delete snapshot: %w", err)
	}

	log.Infof("Snapshot %s deleted", name)
	return nil
}

// ListSnapshots returns all available snapshots
func (c *Client) ListSnapshots(ctx context.Context) ([]string, error) {
	output, err := c.runCommandOutput(ctx, "ls", "-1", c.config.SnapshotsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list snapshots: %w", err)
	}

	return splitLines(output), nil
}

// GetSnapshotInfo returns information about a snapshot
func (c *Client) GetSnapshotInfo(ctx context.Context, name string) (*SnapshotInfo, error) {
	snapshotPath := filepath.Join(c.config.SnapshotsPath, name)

	exists, err := c.fileExists(ctx, snapshotPath)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("snapshot not found: %s", name)
	}

	// Check for config.json in snapshot
	configPath := filepath.Join(snapshotPath, "config")
	configData, err := c.readFile(ctx, configPath)

	info := &SnapshotInfo{
		Name: name,
		Path: snapshotPath,
	}

	if err == nil {
		var vmConfig VmConfig
		if json.Unmarshal(configData, &vmConfig) == nil {
			info.VmConfig = &vmConfig
		}
	}

	// Get disk size if present (check qcow2 first, then legacy raw)
	diskPath := filepath.Join(snapshotPath, "disk.qcow2")
	if exists, _ := c.fileExists(ctx, diskPath); !exists {
		diskPath = filepath.Join(snapshotPath, "disk.raw")
	}
	sizeOutput, err := c.runCommandOutput(ctx, "stat", "-c", "%s", diskPath)
	if err == nil {
		var size int64
		fmt.Sscanf(sizeOutput, "%d", &size)
		info.DiskSizeBytes = size
	}

	return info, nil
}

// SnapshotInfo contains information about a snapshot
type SnapshotInfo struct {
	Name          string    `json:"name"`
	Path          string    `json:"path"`
	DiskSizeBytes int64     `json:"diskSizeBytes,omitempty"`
	VmConfig      *VmConfig `json:"vmConfig,omitempty"`
	CreatedAt     time.Time `json:"createdAt,omitempty"`
}
