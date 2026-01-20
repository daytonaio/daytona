// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package cloudhypervisor

import (
	"context"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// Destroy completely removes a VM and all its resources
func (c *Client) Destroy(ctx context.Context, sandboxId string) error {
	mutex := c.getSandboxMutex(sandboxId)
	mutex.Lock()
	defer mutex.Unlock()

	log.Infof("Destroying sandbox %s", sandboxId)

	// Release IP back to pool
	c.ipPool.Release(sandboxId)
	GetIPCache().Delete(sandboxId)

	socketPath := c.getSocketPath(sandboxId)
	socketExists, _ := c.fileExists(ctx, socketPath)

	if socketExists {
		// Try to get VM state first
		info, err := c.GetInfo(ctx, sandboxId)
		if err != nil {
			log.Warnf("Failed to get VM info during destroy: %v", err)
		} else {
			// If running or paused, shut it down first
			if info.State == VmStateRunning || info.State == VmStatePaused {
				log.Infof("Shutting down VM before destroy")
				_, _ = c.apiRequest(ctx, sandboxId, http.MethodPut, "vm.shutdown", nil)
			}
		}

		// Delete the VM through API
		if _, err := c.apiRequest(ctx, sandboxId, http.MethodPut, "vm.delete", nil); err != nil {
			log.Warnf("Failed to delete VM through API: %v", err)
		}

		// Shutdown the VMM
		if _, err := c.apiRequest(ctx, sandboxId, http.MethodPut, "vmm.shutdown", nil); err != nil {
			log.Warnf("Failed to shutdown VMM: %v", err)
		}
	}

	// Kill any remaining process
	if err := c.killVMProcess(ctx, sandboxId); err != nil {
		log.Warnf("Failed to kill VM process: %v", err)
	}

	// Release TAP interface (back to pool or delete)
	if c.tapPool.IsEnabled() {
		if err := c.tapPool.Release(ctx, sandboxId); err != nil {
			log.Warnf("Failed to release TAP to pool: %v", err)
		}
	} else {
		tapName := c.getTapName(sandboxId)
		if err := c.deleteTapInterface(ctx, tapName); err != nil {
			log.Warnf("Failed to delete TAP interface: %v", err)
		}
	}

	// Remove sandbox directory
	sandboxDir := c.getSandboxDir(sandboxId)
	if err := c.runCommand(ctx, "rm", "-rf", sandboxDir); err != nil {
		log.Warnf("Failed to remove sandbox directory: %v", err)
	}

	// Remove socket
	if err := c.runCommand(ctx, "rm", "-f", socketPath); err != nil {
		log.Warnf("Failed to remove socket: %v", err)
	}

	// Clear HTTP client for this sandbox
	c.httpMutex.Lock()
	delete(c.httpClients, socketPath)
	c.httpMutex.Unlock()

	log.Infof("Sandbox %s destroyed successfully", sandboxId)
	return nil
}

// Delete is an alias for Destroy
func (c *Client) Delete(ctx context.Context, sandboxId string) error {
	return c.Destroy(ctx, sandboxId)
}

// RemoveDestroyed cleans up any remaining resources for a destroyed sandbox
// This is useful for cleaning up orphaned resources
func (c *Client) RemoveDestroyed(ctx context.Context, sandboxId string) error {
	log.Infof("Removing destroyed sandbox %s resources", sandboxId)

	// Kill any remaining process
	_ = c.killVMProcess(ctx, sandboxId)

	// Delete TAP interface
	tapName := c.getTapName(sandboxId)
	_ = c.deleteTapInterface(ctx, tapName)

	// Remove sandbox directory
	sandboxDir := c.getSandboxDir(sandboxId)
	_ = c.runCommand(ctx, "rm", "-rf", sandboxDir)

	// Remove socket
	socketPath := c.getSocketPath(sandboxId)
	_ = c.runCommand(ctx, "rm", "-f", socketPath)

	return nil
}

// List returns all sandbox IDs
func (c *Client) List(ctx context.Context) ([]string, error) {
	output, err := c.runCommandOutput(ctx, "ls", "-1", c.config.SocketsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list sockets: %w", err)
	}

	var sandboxIds []string
	for _, line := range splitLines(output) {
		if len(line) > 5 && line[len(line)-5:] == ".sock" {
			sandboxIds = append(sandboxIds, line[:len(line)-5])
		}
	}

	return sandboxIds, nil
}

// ListWithInfo returns all sandboxes with their info
func (c *Client) ListWithInfo(ctx context.Context) ([]*SandboxInfo, error) {
	ids, err := c.List(ctx)
	if err != nil {
		return nil, err
	}

	var sandboxes []*SandboxInfo
	for _, id := range ids {
		info, err := c.GetSandboxInfo(ctx, id)
		if err != nil {
			log.Warnf("Failed to get info for sandbox %s: %v", id, err)
			continue
		}
		sandboxes = append(sandboxes, info)
	}

	return sandboxes, nil
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
