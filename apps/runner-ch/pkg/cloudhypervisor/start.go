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

// StartVM boots or resumes a VM
// If the VM is paused, it resumes. If it's shut off, it boots.
// If the CH process died (socket missing), it attempts recovery.
func (c *Client) StartVM(ctx context.Context, sandboxId string) error {
	mutex := c.getSandboxMutex(sandboxId)
	mutex.Lock()
	defer mutex.Unlock()

	log.Infof("Starting sandbox %s", sandboxId)

	// Check if socket exists - if not, the CH process died and we need recovery
	socketPath := c.getSocketPath(sandboxId)
	socketExists, _ := c.fileExists(ctx, socketPath)
	if !socketExists {
		// Check if disk exists (sandbox was created but CH died)
		diskPath := c.getDiskPath(sandboxId)
		diskExists, _ := c.fileExists(ctx, diskPath)
		if diskExists {
			log.Warnf("Sandbox %s socket missing but disk exists - attempting recovery", sandboxId)
			// Unlock mutex before calling RecoverSandbox (which will re-lock)
			mutex.Unlock()
			err := c.RecoverSandbox(ctx, sandboxId)
			mutex.Lock()
			if err != nil {
				return fmt.Errorf("failed to recover sandbox: %w", err)
			}
			log.Infof("Sandbox %s recovered successfully", sandboxId)
			return nil
		}
		return fmt.Errorf("sandbox %s not found (no socket or disk)", sandboxId)
	}

	// Get current state
	info, err := c.GetInfo(ctx, sandboxId)
	if err != nil {
		return fmt.Errorf("failed to get VM info: %w", err)
	}

	switch info.State {
	case VmStateRunning:
		log.Infof("Sandbox %s is already running", sandboxId)
		return nil

	case VmStatePaused:
		// Resume from pause
		log.Infof("Resuming paused sandbox %s", sandboxId)
		if _, err := c.apiRequest(ctx, sandboxId, http.MethodPut, "vm.resume", nil); err != nil {
			return fmt.Errorf("failed to resume VM: %w", err)
		}
		// Clean up checkpoint after successful resume (no longer needed)
		if c.hasCheckpoint(ctx, sandboxId) {
			log.Debugf("Cleaning up checkpoint for %s after resume", sandboxId)
			_ = c.deleteCheckpoint(ctx, sandboxId)
		}

	case VmStateCreated, VmStateShutdown:
		// Boot the VM
		log.Infof("Booting sandbox %s", sandboxId)
		if _, err := c.apiRequest(ctx, sandboxId, http.MethodPut, "vm.boot", nil); err != nil {
			return fmt.Errorf("failed to boot VM: %w", err)
		}

	default:
		return fmt.Errorf("cannot start VM in state: %s", info.State)
	}

	// Wait for running state
	if err := c.waitForState(ctx, sandboxId, VmStateRunning, 60*time.Second); err != nil {
		return fmt.Errorf("failed waiting for running state: %w", err)
	}

	log.Infof("Sandbox %s started successfully", sandboxId)
	return nil
}

// Resume resumes a paused VM (alias for Start when paused)
func (c *Client) Resume(ctx context.Context, sandboxId string) error {
	mutex := c.getSandboxMutex(sandboxId)
	mutex.Lock()
	defer mutex.Unlock()

	log.Infof("Resuming sandbox %s", sandboxId)

	// Get current state
	info, err := c.GetInfo(ctx, sandboxId)
	if err != nil {
		return fmt.Errorf("failed to get VM info: %w", err)
	}

	if info.State == VmStateRunning {
		log.Infof("Sandbox %s is already running", sandboxId)
		return nil
	}

	if info.State != VmStatePaused {
		return fmt.Errorf("VM is not paused (state: %s), cannot resume", info.State)
	}

	if _, err := c.apiRequest(ctx, sandboxId, http.MethodPut, "vm.resume", nil); err != nil {
		return fmt.Errorf("failed to resume VM: %w", err)
	}

	// Wait for running state
	if err := c.waitForState(ctx, sandboxId, VmStateRunning, 30*time.Second); err != nil {
		return fmt.Errorf("failed waiting for running state: %w", err)
	}

	log.Infof("Sandbox %s resumed successfully", sandboxId)
	return nil
}

// Boot boots a VM from created/shutdown state
func (c *Client) Boot(ctx context.Context, sandboxId string) error {
	mutex := c.getSandboxMutex(sandboxId)
	mutex.Lock()
	defer mutex.Unlock()

	log.Infof("Booting sandbox %s", sandboxId)

	// Get current state
	info, err := c.GetInfo(ctx, sandboxId)
	if err != nil {
		return fmt.Errorf("failed to get VM info: %w", err)
	}

	if info.State == VmStateRunning {
		log.Infof("Sandbox %s is already running", sandboxId)
		return nil
	}

	if info.State != VmStateCreated && info.State != VmStateShutdown {
		return fmt.Errorf("VM is not in bootable state (state: %s)", info.State)
	}

	if _, err := c.apiRequest(ctx, sandboxId, http.MethodPut, "vm.boot", nil); err != nil {
		return fmt.Errorf("failed to boot VM: %w", err)
	}

	// Wait for running state
	if err := c.waitForState(ctx, sandboxId, VmStateRunning, 60*time.Second); err != nil {
		return fmt.Errorf("failed waiting for running state: %w", err)
	}

	log.Infof("Sandbox %s booted successfully", sandboxId)
	return nil
}

// waitForState waits for a VM to reach a specific state
func (c *Client) waitForState(ctx context.Context, sandboxId string, targetState VmState, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		info, err := c.GetInfo(ctx, sandboxId)
		if err != nil {
			log.Warnf("Error getting VM info while waiting for state: %v", err)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		if info.State == targetState {
			return nil
		}

		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for state %s", targetState)
}

// Reboot performs a graceful reboot of the VM
func (c *Client) Reboot(ctx context.Context, sandboxId string) error {
	mutex := c.getSandboxMutex(sandboxId)
	mutex.Lock()
	defer mutex.Unlock()

	log.Infof("Rebooting sandbox %s", sandboxId)

	if _, err := c.apiRequest(ctx, sandboxId, http.MethodPut, "vm.reboot", nil); err != nil {
		return fmt.Errorf("failed to reboot VM: %w", err)
	}

	// Wait for running state after reboot
	if err := c.waitForState(ctx, sandboxId, VmStateRunning, 120*time.Second); err != nil {
		return fmt.Errorf("failed waiting for running state after reboot: %w", err)
	}

	log.Infof("Sandbox %s rebooted successfully", sandboxId)
	return nil
}
