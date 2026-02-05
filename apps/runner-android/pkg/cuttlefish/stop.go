// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package cuttlefish

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
)

// Stop stops a Cuttlefish instance
func (c *Client) Stop(ctx context.Context, sandboxId string) error {
	mutex := c.getSandboxMutex(sandboxId)
	mutex.Lock()
	defer mutex.Unlock()

	log.Infof("Stopping sandbox %s", sandboxId)

	info, exists := c.GetInstance(sandboxId)
	if !exists {
		return fmt.Errorf("sandbox %s not found", sandboxId)
	}

	// Check current state
	state := c.getInstanceState(ctx, info.InstanceNum)
	if state == InstanceStateStopped {
		log.Infof("Sandbox %s is already stopped", sandboxId)
		return nil
	}

	// Stop the instance
	if err := c.stopInstance(ctx, info.InstanceNum); err != nil {
		return fmt.Errorf("failed to stop instance: %w", err)
	}

	// Update state
	c.mutex.Lock()
	info.State = InstanceStateStopped
	c.mutex.Unlock()

	log.Infof("Sandbox %s stopped successfully", sandboxId)
	return nil
}

// Pause is an alias for Stop (Cuttlefish doesn't support pause)
func (c *Client) Pause(ctx context.Context, sandboxId string) error {
	return c.Stop(ctx, sandboxId)
}

// PauseVM is an alias for Stop
func (c *Client) PauseVM(ctx context.Context, sandboxId string) error {
	return c.Stop(ctx, sandboxId)
}

// Shutdown is an alias for Stop
func (c *Client) Shutdown(ctx context.Context, sandboxId string) error {
	return c.Stop(ctx, sandboxId)
}

// ForceStop stops the instance forcefully
func (c *Client) ForceStop(ctx context.Context, sandboxId string) error {
	mutex := c.getSandboxMutex(sandboxId)
	mutex.Lock()
	defer mutex.Unlock()

	log.Infof("Force stopping sandbox %s", sandboxId)

	info, exists := c.GetInstance(sandboxId)
	if !exists {
		return fmt.Errorf("sandbox %s not found", sandboxId)
	}

	// Kill processes associated with this instance
	killCmd := fmt.Sprintf("pkill -9 -f 'cuttlefish.*instance_nums.*%d' || true", info.InstanceNum)
	_, _ = c.runShellScript(ctx, killCmd)

	// Also try stop_cvd
	_ = c.stopInstance(ctx, info.InstanceNum)

	// Update state
	c.mutex.Lock()
	info.State = InstanceStateStopped
	c.mutex.Unlock()

	log.Infof("Sandbox %s force stopped", sandboxId)
	return nil
}
