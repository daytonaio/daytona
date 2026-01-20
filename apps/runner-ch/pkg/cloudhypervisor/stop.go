// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package cloudhypervisor

import (
	"context"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

// Stop stops a VM - by default uses pause for fast resume capability
func (c *Client) Stop(ctx context.Context, sandboxId string) error {
	// Use pause by default for instant resume
	return c.Pause(ctx, sandboxId)
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
