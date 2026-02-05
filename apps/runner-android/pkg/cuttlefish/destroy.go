// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package cuttlefish

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
)

// Destroy completely removes a Cuttlefish instance and all its resources
func (c *Client) Destroy(ctx context.Context, sandboxId string) error {
	mutex := c.getSandboxMutex(sandboxId)
	mutex.Lock()
	defer mutex.Unlock()

	log.Infof("Destroying sandbox %s", sandboxId)

	info, exists := c.GetInstance(sandboxId)
	if !exists {
		log.Warnf("Sandbox %s not found, nothing to destroy", sandboxId)
		return nil
	}

	instanceNum := info.InstanceNum

	// Stop the instance if running
	state := c.getInstanceState(ctx, instanceNum)
	if state == InstanceStateRunning {
		log.Infof("Stopping instance %d before destroy", instanceNum)
		if err := c.stopInstance(ctx, instanceNum); err != nil {
			log.Warnf("Failed to stop instance during destroy: %v", err)
		}
	}

	// Kill any remaining processes
	killCmd := fmt.Sprintf("pkill -9 -f 'cuttlefish.*instance_nums.*%d' || true", instanceNum)
	_, _ = c.runShellScript(ctx, killCmd)

	// Remove runtime directory
	runtimeDir := c.getRuntimeDir(instanceNum)
	if err := c.runCommand(ctx, "rm", "-rf", runtimeDir); err != nil {
		log.Warnf("Failed to remove runtime directory %s: %v", runtimeDir, err)
	}

	// Remove instance data directory
	instanceDir := c.getInstanceDir(sandboxId)
	if err := c.runCommand(ctx, "rm", "-rf", instanceDir); err != nil {
		log.Warnf("Failed to remove instance directory %s: %v", instanceDir, err)
	}

	// Remove from tracking
	c.mutex.Lock()
	delete(c.instances, sandboxId)
	delete(c.instanceNums, instanceNum)
	c.mutex.Unlock()

	// Save updated mappings
	if err := c.saveInstanceMappings(); err != nil {
		log.Warnf("Failed to save instance mappings: %v", err)
	}

	// Clean up sandbox mutex
	c.sandboxMuMu.Lock()
	delete(c.sandboxMutex, sandboxId)
	c.sandboxMuMu.Unlock()

	log.Infof("Sandbox %s destroyed successfully", sandboxId)
	return nil
}

// Delete is an alias for Destroy
func (c *Client) Delete(ctx context.Context, sandboxId string) error {
	return c.Destroy(ctx, sandboxId)
}

// RemoveDestroyed cleans up any remaining resources for a destroyed sandbox
func (c *Client) RemoveDestroyed(ctx context.Context, sandboxId string) error {
	log.Infof("Removing destroyed sandbox %s resources", sandboxId)

	// Get instance info if it exists
	info, exists := c.GetInstance(sandboxId)
	if exists {
		// Kill any remaining processes
		killCmd := fmt.Sprintf("pkill -9 -f 'cuttlefish.*instance_nums.*%d' || true", info.InstanceNum)
		_, _ = c.runShellScript(ctx, killCmd)

		// Remove runtime directory
		runtimeDir := c.getRuntimeDir(info.InstanceNum)
		_ = c.runCommand(ctx, "rm", "-rf", runtimeDir)
	}

	// Remove instance data directory
	instanceDir := c.getInstanceDir(sandboxId)
	_ = c.runCommand(ctx, "rm", "-rf", instanceDir)

	return nil
}

// splitLines splits a string into non-empty lines
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			line := s[start:i]
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
			if len(line) > 0 {
				lines = append(lines, line)
			}
			start = i + 1
		}
	}
	if start < len(s) {
		line := s[start:]
		if len(line) > 0 {
			lines = append(lines, line)
		}
	}
	return lines
}
