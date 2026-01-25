// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package cloudhypervisor

import (
	"context"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// Resize resizes a VM's CPU and memory allocation
// Cloud Hypervisor supports live resizing for both vCPUs and memory
func (c *Client) Resize(ctx context.Context, sandboxId string, cpus int, memoryMB uint64) error {
	mutex := c.getSandboxMutex(sandboxId)
	mutex.Lock()
	defer mutex.Unlock()

	log.Infof("Resizing sandbox %s: cpus=%d, memoryMB=%d", sandboxId, cpus, memoryMB)

	// Get current state
	info, err := c.GetInfo(ctx, sandboxId)
	if err != nil {
		return fmt.Errorf("failed to get VM info: %w", err)
	}

	if info.State != VmStateRunning && info.State != VmStatePaused {
		return fmt.Errorf("VM must be running or paused to resize (current state: %s)", info.State)
	}

	// Build resize config
	resizeConfig := ResizeConfig{}

	if cpus > 0 {
		resizeConfig.DesiredVcpus = &cpus
	}

	if memoryMB > 0 {
		memoryBytes := memoryMB * 1024 * 1024
		resizeConfig.DesiredRam = &memoryBytes
	}

	// For remote mode, use ch-remote
	if c.IsRemote() {
		return c.resizeRemote(ctx, sandboxId, cpus, memoryMB)
	}

	// Local mode: use API
	if _, err := c.apiRequest(ctx, sandboxId, http.MethodPut, "vm.resize", resizeConfig); err != nil {
		return fmt.Errorf("failed to resize VM: %w", err)
	}

	log.Infof("Sandbox %s resized successfully", sandboxId)
	return nil
}

// resizeRemote resizes a VM via ch-remote over SSH
func (c *Client) resizeRemote(ctx context.Context, sandboxId string, cpus int, memoryMB uint64) error {
	socketPath := c.getSocketPath(sandboxId)

	// Build resize command
	cmdStr := fmt.Sprintf("ch-remote --api-socket %s resize", socketPath)

	if cpus > 0 {
		cmdStr += fmt.Sprintf(" --cpus %d", cpus)
	}

	if memoryMB > 0 {
		cmdStr += fmt.Sprintf(" --memory %d", memoryMB*1024*1024)
	}

	log.Debugf("Executing ch-remote resize: %s", cmdStr)

	output, err := c.runCommandOutput(ctx, "sh", "-c", cmdStr)
	if err != nil {
		return fmt.Errorf("ch-remote resize failed: %w (output: %s)", err, output)
	}

	return nil
}

// ResizeDisk resizes a VM's disk
func (c *Client) ResizeDisk(ctx context.Context, sandboxId, diskId string, newSizeBytes uint64) error {
	mutex := c.getSandboxMutex(sandboxId)
	mutex.Lock()
	defer mutex.Unlock()

	log.Infof("Resizing disk %s for sandbox %s to %d bytes", diskId, sandboxId, newSizeBytes)

	resizeDiskConfig := ResizeDiskConfig{
		DiskId:  diskId,
		NewSize: newSizeBytes,
	}

	if c.IsRemote() {
		socketPath := c.getSocketPath(sandboxId)
		cmdStr := fmt.Sprintf("ch-remote --api-socket %s resize-disk --disk_id %s --size %d",
			socketPath, diskId, newSizeBytes)

		output, err := c.runCommandOutput(ctx, "sh", "-c", cmdStr)
		if err != nil {
			return fmt.Errorf("ch-remote resize-disk failed: %w (output: %s)", err, output)
		}
		return nil
	}

	if _, err := c.apiRequest(ctx, sandboxId, http.MethodPut, "vm.resize-disk", resizeDiskConfig); err != nil {
		return fmt.Errorf("failed to resize disk: %w", err)
	}

	log.Infof("Disk %s resized successfully", diskId)
	return nil
}

// SetVMBalloon sets the balloon size for a VM to reclaim memory
// sizeBytes is the amount of memory to reclaim from the VM (balloon inflation)
// A size of 0 means no balloon (VM gets full memory)
func (c *Client) SetVMBalloon(ctx context.Context, sandboxId string, sizeBytes uint64) error {
	mutex := c.getSandboxMutex(sandboxId)
	mutex.Lock()
	defer mutex.Unlock()

	log.Debugf("Setting balloon for sandbox %s to %d bytes", sandboxId, sizeBytes)

	resizeConfig := ResizeConfig{
		DesiredBalloon: &sizeBytes,
	}

	if c.IsRemote() {
		return c.setBalloonRemote(ctx, sandboxId, sizeBytes)
	}

	if _, err := c.apiRequest(ctx, sandboxId, http.MethodPut, "vm.resize", resizeConfig); err != nil {
		return fmt.Errorf("failed to set balloon: %w", err)
	}

	log.Debugf("Balloon for sandbox %s set to %d bytes", sandboxId, sizeBytes)
	return nil
}

// setBalloonRemote sets balloon via ch-remote over SSH
func (c *Client) setBalloonRemote(ctx context.Context, sandboxId string, sizeBytes uint64) error {
	socketPath := c.getSocketPath(sandboxId)

	cmdStr := fmt.Sprintf("ch-remote --api-socket %s resize --balloon %d", socketPath, sizeBytes)

	log.Debugf("Executing ch-remote balloon resize: %s", cmdStr)

	output, err := c.runCommandOutput(ctx, "sh", "-c", cmdStr)
	if err != nil {
		return fmt.Errorf("ch-remote balloon resize failed: %w (output: %s)", err, output)
	}

	return nil
}
