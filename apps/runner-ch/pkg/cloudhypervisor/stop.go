// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package cloudhypervisor

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"
)

// Stop stops a VM by creating a memory checkpoint for later restoration
// This preserves full VM state (memory + devices) so it can survive CH restarts
func (c *Client) Stop(ctx context.Context, sandboxId string) error {
	mutex := c.getSandboxMutex(sandboxId)
	mutex.Lock()
	defer mutex.Unlock()

	log.Infof("Stopping sandbox %s with memory checkpoint", sandboxId)

	// Get current state
	info, err := c.GetInfo(ctx, sandboxId)
	if err != nil {
		return fmt.Errorf("failed to get VM info: %w", err)
	}

	// If already stopped/paused and checkpoint exists, we're done
	if info.State == VmStatePaused || info.State == VmStateShutdown {
		if c.hasCheckpoint(ctx, sandboxId) {
			log.Infof("Sandbox %s already stopped with checkpoint", sandboxId)
			return nil
		}
	}

	if info.State != VmStateRunning && info.State != VmStatePaused {
		return fmt.Errorf("VM is not running or paused (state: %s), cannot stop", info.State)
	}

	// Step 1: Pause the VM if running
	if info.State == VmStateRunning {
		log.Infof("Pausing VM for checkpoint...")
		if _, err := c.apiRequest(ctx, sandboxId, http.MethodPut, "vm.pause", nil); err != nil {
			return fmt.Errorf("failed to pause VM: %w", err)
		}
		if err := c.waitForState(ctx, sandboxId, VmStatePaused, 30*time.Second); err != nil {
			return fmt.Errorf("failed waiting for paused state: %w", err)
		}
	}

	// Step 2: Create checkpoint (memory snapshot)
	checkpointPath := c.getCheckpointPath(sandboxId)
	log.Infof("Creating memory checkpoint at %s", checkpointPath)

	if err := c.createCheckpoint(ctx, sandboxId, checkpointPath); err != nil {
		// If checkpoint fails, still leave VM paused (better than nothing)
		log.Warnf("Failed to create checkpoint: %v (VM remains paused)", err)
		return nil
	}

	log.Infof("Sandbox %s stopped with memory checkpoint", sandboxId)
	return nil
}

// getCheckpointPath returns the checkpoint directory path for a sandbox
func (c *Client) getCheckpointPath(sandboxId string) string {
	return filepath.Join(c.getSandboxDir(sandboxId), "checkpoint")
}

// hasCheckpoint checks if a checkpoint exists for a sandbox
func (c *Client) hasCheckpoint(ctx context.Context, sandboxId string) bool {
	checkpointPath := c.getCheckpointPath(sandboxId)
	// Check for state.json which is always created by vm.snapshot
	statePath := filepath.Join(checkpointPath, "state.json")
	exists, _ := c.fileExists(ctx, statePath)
	return exists
}

// createCheckpoint creates a memory checkpoint using vm.snapshot
func (c *Client) createCheckpoint(ctx context.Context, sandboxId, checkpointPath string) error {
	// Ensure checkpoint directory exists (clean it first to avoid stale data)
	cleanupCmd := fmt.Sprintf("rm -rf %s && mkdir -p %s", checkpointPath, checkpointPath)
	if _, err := c.runShellScript(ctx, cleanupCmd); err != nil {
		return fmt.Errorf("failed to prepare checkpoint directory: %w", err)
	}

	// Create snapshot via CH API
	snapshotConfig := map[string]string{
		"destination_url": fmt.Sprintf("file://%s", checkpointPath),
	}

	if _, err := c.apiRequest(ctx, sandboxId, http.MethodPut, "vm.snapshot", snapshotConfig); err != nil {
		// Cleanup on failure
		_ = c.runCommand(ctx, "rm", "-rf", checkpointPath)
		return fmt.Errorf("vm.snapshot failed: %w", err)
	}

	log.Infof("Checkpoint created successfully at %s", checkpointPath)
	return nil
}

// deleteCheckpoint removes a checkpoint
func (c *Client) deleteCheckpoint(ctx context.Context, sandboxId string) error {
	checkpointPath := c.getCheckpointPath(sandboxId)
	if _, err := c.runShellScript(ctx, fmt.Sprintf("rm -rf %s", checkpointPath)); err != nil {
		return fmt.Errorf("failed to delete checkpoint: %w", err)
	}
	return nil
}

// Pause suspends a running VM (fast, keeps memory in place)
// This is the preferred way to "stop" for fast resume
func (c *Client) Pause(ctx context.Context, sandboxId string) error {
	mutex := c.getSandboxMutex(sandboxId)
	mutex.Lock()
	defer mutex.Unlock()

	log.Infof("Pausing sandbox %s", sandboxId)

	// Get current state
	info, err := c.GetInfo(ctx, sandboxId)
	if err != nil {
		return fmt.Errorf("failed to get VM info: %w", err)
	}

	if info.State == VmStatePaused {
		log.Infof("Sandbox %s is already paused", sandboxId)
		return nil
	}

	if info.State != VmStateRunning {
		return fmt.Errorf("VM is not running (state: %s), cannot pause", info.State)
	}

	if _, err := c.apiRequest(ctx, sandboxId, http.MethodPut, "vm.pause", nil); err != nil {
		return fmt.Errorf("failed to pause VM: %w", err)
	}

	// Wait for paused state
	if err := c.waitForState(ctx, sandboxId, VmStatePaused, 30*time.Second); err != nil {
		return fmt.Errorf("failed waiting for paused state: %w", err)
	}

	log.Infof("Sandbox %s paused successfully", sandboxId)
	return nil
}

// Shutdown performs a graceful shutdown of the VM
// The VM will need to be booted again (not just resumed)
func (c *Client) Shutdown(ctx context.Context, sandboxId string) error {
	mutex := c.getSandboxMutex(sandboxId)
	mutex.Lock()
	defer mutex.Unlock()

	log.Infof("Shutting down sandbox %s", sandboxId)

	// Get current state
	info, err := c.GetInfo(ctx, sandboxId)
	if err != nil {
		return fmt.Errorf("failed to get VM info: %w", err)
	}

	if info.State == VmStateShutdown {
		log.Infof("Sandbox %s is already shut down", sandboxId)
		return nil
	}

	if info.State != VmStateRunning && info.State != VmStatePaused {
		return fmt.Errorf("VM is not running or paused (state: %s), cannot shutdown", info.State)
	}

	// Send power button signal for graceful shutdown
	if _, err := c.apiRequest(ctx, sandboxId, http.MethodPut, "vm.power-button", nil); err != nil {
		log.Warnf("Power button failed, trying direct shutdown: %v", err)
		// Try direct shutdown
		if _, err := c.apiRequest(ctx, sandboxId, http.MethodPut, "vm.shutdown", nil); err != nil {
			return fmt.Errorf("failed to shutdown VM: %w", err)
		}
	}

	// Wait for shutdown state (with timeout)
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	if err := c.waitForState(ctx, sandboxId, VmStateShutdown, 60*time.Second); err != nil {
		// If graceful shutdown times out, force shutdown
		log.Warnf("Graceful shutdown timed out, forcing: %v", err)
		if _, err := c.apiRequest(context.Background(), sandboxId, http.MethodPut, "vm.shutdown", nil); err != nil {
			return fmt.Errorf("failed to force shutdown VM: %w", err)
		}
	}

	log.Infof("Sandbox %s shut down successfully", sandboxId)
	return nil
}

// ForceStop immediately stops the VM without graceful shutdown
func (c *Client) ForceStop(ctx context.Context, sandboxId string) error {
	mutex := c.getSandboxMutex(sandboxId)
	mutex.Lock()
	defer mutex.Unlock()

	log.Infof("Force stopping sandbox %s", sandboxId)

	if _, err := c.apiRequest(ctx, sandboxId, http.MethodPut, "vm.shutdown", nil); err != nil {
		return fmt.Errorf("failed to force stop VM: %w", err)
	}

	log.Infof("Sandbox %s force stopped", sandboxId)
	return nil
}

// PauseVM is an alias for Pause - used by executor methods
func (c *Client) PauseVM(ctx context.Context, sandboxId string) error {
	return c.Pause(ctx, sandboxId)
}
