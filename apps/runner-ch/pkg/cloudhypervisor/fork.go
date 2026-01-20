// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package cloudhypervisor

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// ForkOptions specifies options for forking a VM
type ForkOptions struct {
	SourceSandboxId string // Source VM to fork from
	NewSandboxId    string // ID for the new forked VM
	Prefault        bool   // Prefault memory pages for faster access
}

// ForkVM creates a new VM as a CoW copy of an existing running/paused VM
// This is optimized for fast forking:
// - Disk uses qcow2 overlay (instant, no copy)
// - Memory state is captured via vm.snapshot and restored
// - Network namespace provides isolation with same internal IP
//
// The fork process:
// 1. Pause source VM (if running)
// 2. Create temporary snapshot via vm.snapshot API (memory + device state)
// 3. Create qcow2 overlay disk pointing to source VM's disk as backing file
// 4. Create new network namespace for child
// 5. Start new cloud-hypervisor process in namespace
// 6. Call vm.restore API with snapshot path
// 7. Resume source VM
// 8. Resume child VM
// 9. Cleanup temporary snapshot (memory state only)
func (c *Client) ForkVM(ctx context.Context, opts ForkOptions) (*SandboxInfo, error) {
	sourceMutex := c.getSandboxMutex(opts.SourceSandboxId)
	sourceMutex.Lock()
	defer sourceMutex.Unlock()

	newMutex := c.getSandboxMutex(opts.NewSandboxId)
	newMutex.Lock()
	defer newMutex.Unlock()

	log.Infof("Forking sandbox %s to %s", opts.SourceSandboxId, opts.NewSandboxId)

	// Check if source VM exists and get its state
	sourceInfo, err := c.GetInfo(ctx, opts.SourceSandboxId)
	if err != nil {
		return nil, fmt.Errorf("failed to get source VM info: %w", err)
	}

	// VM must be running or paused to fork
	if sourceInfo.State != VmStateRunning && sourceInfo.State != VmStatePaused {
		return nil, fmt.Errorf("source VM must be running or paused to fork (current state: %s)", sourceInfo.State)
	}

	// Check if target sandbox already exists
	targetSocketPath := c.getSocketPath(opts.NewSandboxId)
	exists, _ := c.fileExists(ctx, targetSocketPath)
	if exists {
		return nil, fmt.Errorf("target sandbox %s already exists", opts.NewSandboxId)
	}

	// Track if we paused the source VM (so we can resume it on cleanup)
	wasPaused := sourceInfo.State == VmStatePaused

	// Step 1: Pause source VM if running
	if !wasPaused {
		log.Infof("Pausing source VM for fork")
		if _, err := c.apiRequest(ctx, opts.SourceSandboxId, http.MethodPut, "vm.pause", nil); err != nil {
			return nil, fmt.Errorf("failed to pause source VM: %w", err)
		}
	}

	// Ensure we resume source VM on any error
	defer func() {
		if !wasPaused {
			log.Infof("Resuming source VM after fork")
			if _, err := c.apiRequest(ctx, opts.SourceSandboxId, http.MethodPut, "vm.resume", nil); err != nil {
				log.Warnf("Failed to resume source VM: %v", err)
			}
		}
	}()

	// Step 2: Create temporary snapshot (memory + device state)
	snapshotName := fmt.Sprintf("fork-%s-%d", opts.SourceSandboxId, time.Now().UnixNano())
	snapshotPath := filepath.Join(c.config.SnapshotsPath, snapshotName)

	log.Infof("Creating temporary snapshot %s for fork", snapshotName)

	// Ensure snapshots directory exists
	if err := c.runCommand(ctx, "mkdir", "-p", c.config.SnapshotsPath); err != nil {
		return nil, fmt.Errorf("failed to create snapshots directory: %w", err)
	}

	// Create snapshot directory
	if err := c.runCommand(ctx, "mkdir", "-p", snapshotPath); err != nil {
		return nil, fmt.Errorf("failed to create snapshot directory: %w", err)
	}

	// Create snapshot via CH API (captures memory + device state)
	snapshotConfig := map[string]string{
		"destination_url": fmt.Sprintf("file://%s", snapshotPath),
	}

	if _, err := c.apiRequest(ctx, opts.SourceSandboxId, http.MethodPut, "vm.snapshot", snapshotConfig); err != nil {
		_ = c.runCommand(ctx, "rm", "-rf", snapshotPath)
		return nil, fmt.Errorf("failed to create snapshot: %w", err)
	}

	// Step 3: Create qcow2 overlay disk (instant CoW, no full copy)
	sourceDiskPath := c.getDiskPath(opts.SourceSandboxId)
	targetSandboxDir := c.getSandboxDir(opts.NewSandboxId)
	targetDiskPath := c.getDiskPath(opts.NewSandboxId)

	log.Infof("Creating CoW overlay disk for fork")

	// Create target sandbox directory and qcow2 overlay in one command
	createDiskCmd := fmt.Sprintf(
		`mkdir -p "%s" && qemu-img create -f qcow2 -F qcow2 -b "%s" "%s"`,
		targetSandboxDir, sourceDiskPath, targetDiskPath)

	if output, err := c.runShellScript(ctx, createDiskCmd); err != nil {
		_ = c.runCommand(ctx, "rm", "-rf", snapshotPath)
		_ = c.runCommand(ctx, "rm", "-rf", targetSandboxDir)
		return nil, fmt.Errorf("failed to create overlay disk: %w (output: %s)", err, output)
	}

	// Step 4: Create network namespace for child VM
	netns, err := c.netnsPool.Create(ctx, opts.NewSandboxId)
	if err != nil {
		_ = c.runCommand(ctx, "rm", "-rf", snapshotPath)
		_ = c.runCommand(ctx, "rm", "-rf", targetSandboxDir)
		return nil, fmt.Errorf("failed to create network namespace: %w", err)
	}
	log.Infof("Created network namespace %s for fork", netns.NamespaceName)

	// Store guest IP for proxy routing
	ipFilePath := filepath.Join(targetSandboxDir, "ip")
	_ = c.runCommand(ctx, "sh", "-c", fmt.Sprintf("echo '%s' > %s", netns.GuestIP, ipFilePath))
	GetIPCache().Set(opts.NewSandboxId, netns.GuestIP)

	// Step 5: Start cloud-hypervisor process in namespace
	targetSocketPath = c.getSocketPath(opts.NewSandboxId)
	logPath := filepath.Join(targetSandboxDir, "cloud-hypervisor.log")

	log.Infof("Starting cloud-hypervisor for fork in namespace %s", netns.NamespaceName)

	// Start CH in namespace and wait for socket
	startCmd := fmt.Sprintf(`
# Start cloud-hypervisor in namespace background
nohup ip netns exec %s cloud-hypervisor --api-socket %s > %s 2>&1 &
pid=$!

# Wait for socket with fast polling
timeout=600
elapsed=0
while [ $elapsed -lt $timeout ]; do
    if [ -S '%s' ]; then
        sleep 0.2
        echo "READY"
        exit 0
    fi
    sleep 0.05
    elapsed=$((elapsed + 1))
done
echo "TIMEOUT"
kill $pid 2>/dev/null || true
exit 1
`, netns.NamespaceName, targetSocketPath, logPath, targetSocketPath)

	output, err := c.runShellScript(ctx, startCmd)
	if err != nil || strings.Contains(output, "TIMEOUT") {
		c.cleanupFork(ctx, opts.NewSandboxId, snapshotPath)
		return nil, fmt.Errorf("failed to start CH for fork: %v (output: %s)", err, output)
	}

	// Step 6: Restore VM from snapshot
	log.Infof("Restoring forked VM from snapshot")

	restoreConfig := RestoreConfig{
		SourceUrl: fmt.Sprintf("file://%s", snapshotPath),
		Prefault:  opts.Prefault,
	}

	if _, err := c.apiRequest(ctx, opts.NewSandboxId, http.MethodPut, "vm.restore", restoreConfig); err != nil {
		c.cleanupFork(ctx, opts.NewSandboxId, snapshotPath)
		return nil, fmt.Errorf("failed to restore forked VM: %w", err)
	}

	// Step 7: Resume child VM if it restored as paused
	childInfo, err := c.GetInfo(ctx, opts.NewSandboxId)
	if err != nil {
		log.Warnf("Failed to get child VM info after restore: %v", err)
	} else if childInfo.State == VmStatePaused {
		log.Infof("Resuming forked VM")
		if _, err := c.apiRequest(ctx, opts.NewSandboxId, http.MethodPut, "vm.resume", nil); err != nil {
			log.Warnf("Failed to resume forked VM: %v", err)
		}
	}

	// Step 8: Cleanup temporary snapshot (memory files are large, disk is not needed)
	// Keep the snapshot directory for debugging, but remove memory files
	log.Infof("Cleaning up temporary snapshot memory files")
	_ = c.runCommand(ctx, "sh", "-c", fmt.Sprintf("rm -f %s/memory* 2>/dev/null || true", snapshotPath))

	// Store parent ID for tracking
	parentFilePath := filepath.Join(targetSandboxDir, "parent")
	_ = c.runCommand(ctx, "sh", "-c", fmt.Sprintf("echo '%s' > %s", opts.SourceSandboxId, parentFilePath))

	log.Infof("Fork completed: %s -> %s", opts.SourceSandboxId, opts.NewSandboxId)

	return c.GetSandboxInfo(ctx, opts.NewSandboxId)
}

// cleanupFork cleans up resources on fork failure
func (c *Client) cleanupFork(ctx context.Context, sandboxId, snapshotPath string) {
	log.Warnf("Cleaning up failed fork for %s", sandboxId)

	// Kill VM process
	_ = c.killVMProcess(ctx, sandboxId)

	// Delete network namespace
	_ = c.netnsPool.Delete(ctx, sandboxId)

	// Remove IP from cache
	GetIPCache().Delete(sandboxId)

	// Remove sandbox directory
	sandboxDir := c.getSandboxDir(sandboxId)
	_ = c.runCommand(ctx, "rm", "-rf", sandboxDir)

	// Remove socket
	socketPath := c.getSocketPath(sandboxId)
	_ = c.runCommand(ctx, "rm", "-f", socketPath)

	// Remove snapshot
	if snapshotPath != "" {
		_ = c.runCommand(ctx, "rm", "-rf", snapshotPath)
	}
}

// GetParentSandboxId returns the parent sandbox ID for a forked VM
func (c *Client) GetParentSandboxId(ctx context.Context, sandboxId string) (string, error) {
	parentFilePath := filepath.Join(c.getSandboxDir(sandboxId), "parent")
	output, err := c.runShellScript(ctx, fmt.Sprintf("cat %s 2>/dev/null", parentFilePath))
	if err != nil {
		return "", nil // Not a fork, no parent
	}
	return strings.TrimSpace(output), nil
}

// IsFork returns true if the sandbox is a fork of another sandbox
func (c *Client) IsFork(ctx context.Context, sandboxId string) bool {
	parent, _ := c.GetParentSandboxId(ctx, sandboxId)
	return parent != ""
}
