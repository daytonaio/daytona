// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package cuttlefish

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

// StartVM starts a stopped Cuttlefish instance
func (c *Client) StartVM(ctx context.Context, sandboxId string) error {
	mutex := c.getSandboxMutex(sandboxId)
	mutex.Lock()
	defer mutex.Unlock()

	log.Infof("Starting sandbox %s", sandboxId)

	// Get instance info
	info, exists := c.GetInstance(sandboxId)
	if !exists {
		return fmt.Errorf("sandbox %s not found", sandboxId)
	}

	// Check current state
	state := c.getInstanceState(ctx, info.InstanceNum)
	if state == InstanceStateRunning {
		log.Infof("Sandbox %s is already running", sandboxId)
		return nil
	}

	// Try to start the existing CVD group first (faster than re-creating)
	if state == InstanceStateStopped {
		if err := c.startExistingInstance(ctx, info.InstanceNum); err != nil {
			log.Warnf("Failed to start existing instance, will try re-creating: %v", err)
		} else {
			// Successfully started existing instance
			goto waitForADB
		}
	}

	// Re-launch the instance (use snapshot from metadata if available)
	{
		snapshot := ""
		if info.Metadata != nil {
			snapshot = info.Metadata["snapshot"]
		}
		if err := c.launchInstance(ctx, info, snapshot); err != nil {
			return fmt.Errorf("failed to start instance: %w", err)
		}
	}

waitForADB:

	// Wait for ADB to be ready
	if err := c.waitForADB(ctx, info.InstanceNum, 120*time.Second); err != nil {
		log.Warnf("ADB not ready after start: %v (continuing anyway)", err)
	}

	// Update state
	c.mutex.Lock()
	info.State = InstanceStateRunning
	c.mutex.Unlock()

	// Reset health monitor state for this sandbox
	if c.healthMonitor != nil {
		c.healthMonitor.ResetSandboxState(sandboxId)
	}

	log.Infof("Sandbox %s started successfully", sandboxId)
	return nil
}

// startExistingInstance starts a stopped CVD group using cvd start
func (c *Client) startExistingInstance(ctx context.Context, instanceNum int) error {
	groupName := fmt.Sprintf("cvd_%d", instanceNum)
	log.Infof("Starting existing CVD group %s", groupName)

	// Use cvd -group_name <group> start to restart a stopped instance
	startCmd := fmt.Sprintf("HOME=%s %s -group_name %s start 2>&1",
		c.config.CVDHome,
		c.config.CVDPath,
		groupName,
	)

	log.Debugf("Running start command: %s", startCmd)

	output, err := c.runShellScript(ctx, startCmd)
	if err != nil {
		return fmt.Errorf("cvd start failed: %w (output: %s)", err, output)
	}

	log.Infof("CVD group %s started successfully", groupName)
	return nil
}

// Resume is an alias for StartVM (Cuttlefish doesn't have pause/resume)
func (c *Client) Resume(ctx context.Context, sandboxId string) error {
	return c.StartVM(ctx, sandboxId)
}

// Boot is an alias for StartVM
func (c *Client) Boot(ctx context.Context, sandboxId string) error {
	return c.StartVM(ctx, sandboxId)
}

// Reboot restarts a Cuttlefish instance via ADB
func (c *Client) Reboot(ctx context.Context, sandboxId string) error {
	mutex := c.getSandboxMutex(sandboxId)
	mutex.Lock()
	defer mutex.Unlock()

	log.Infof("Rebooting sandbox %s", sandboxId)

	info, exists := c.GetInstance(sandboxId)
	if !exists {
		return fmt.Errorf("sandbox %s not found", sandboxId)
	}

	// Use ADB to reboot the device
	rebootCmd := fmt.Sprintf("%s -s %s reboot", c.config.ADBPath, info.ADBSerial)
	if _, err := c.runShellScript(ctx, rebootCmd); err != nil {
		return fmt.Errorf("failed to reboot via ADB: %w", err)
	}

	// Wait for device to come back
	time.Sleep(5 * time.Second)

	// Wait for ADB to be ready again
	if err := c.waitForADB(ctx, info.InstanceNum, 120*time.Second); err != nil {
		return fmt.Errorf("device not ready after reboot: %w", err)
	}

	log.Infof("Sandbox %s rebooted successfully", sandboxId)
	return nil
}

// waitForState waits for an instance to reach a specific state
func (c *Client) waitForState(ctx context.Context, sandboxId string, targetState InstanceState, timeout time.Duration) error {
	info, exists := c.GetInstance(sandboxId)
	if !exists {
		return fmt.Errorf("sandbox %s not found", sandboxId)
	}

	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		state := c.getInstanceState(ctx, info.InstanceNum)
		if state == targetState {
			return nil
		}

		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for state %s", targetState)
}
