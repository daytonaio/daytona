// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package cloudhypervisor

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// CloneOptions specifies options for cloning a VM
type CloneOptions struct {
	SourceSandboxId string // Source VM to clone from
	NewSandboxId    string // ID for the new cloned VM
}

// CloneVM creates an independent copy of a VM with its own disk overlay and memory state
// Unlike ForkVM which uses qcow2 overlay pointing to source disk, CloneVM:
// - Copies the disk overlay file (fast, only copies R/W changes)
// - Creates a new qcow2 with the same base image as source
// - Clones memory state via snapshot/restore (preserves running applications)
// - Results in a fully independent sandbox that doesn't depend on source
//
// The clone process:
// 1. Pause source VM
// 2. Create memory snapshot via vm.snapshot API
// 3. Copy disk overlay and create new qcow2 with same backing file
// 4. Create network namespace for clone
// 5. Restore clone from snapshot (with memory state)
// 6. Resume source VM
func (c *Client) CloneVM(ctx context.Context, opts CloneOptions) (*SandboxInfo, error) {
	sourceMutex := c.getSandboxMutex(opts.SourceSandboxId)
	sourceMutex.Lock()
	defer sourceMutex.Unlock()

	newMutex := c.getSandboxMutex(opts.NewSandboxId)
	newMutex.Lock()
	defer newMutex.Unlock()

	log.Infof("Cloning sandbox %s to %s (with memory state)", opts.SourceSandboxId, opts.NewSandboxId)

	// Check if target sandbox already exists
	targetSocketPath := c.getSocketPath(opts.NewSandboxId)
	exists, _ := c.fileExists(ctx, targetSocketPath)
	if exists {
		return nil, fmt.Errorf("target sandbox %s already exists", opts.NewSandboxId)
	}

	// Verify the source sandbox directory and disk exist
	sourceDiskPath := c.getDiskPath(opts.SourceSandboxId)
	exists, err := c.fileExists(ctx, sourceDiskPath)
	if err != nil || !exists {
		return nil, fmt.Errorf("source sandbox disk not found at %s", sourceDiskPath)
	}

	// Get source VM info (which includes the config)
	sourceInfo, err := c.GetInfo(ctx, opts.SourceSandboxId)
	if err != nil {
		return nil, fmt.Errorf("failed to get source VM info: %w", err)
	}

	// VM must be running or paused to clone with memory state
	if sourceInfo.State != VmStateRunning && sourceInfo.State != VmStatePaused {
		return nil, fmt.Errorf("source VM must be running or paused to clone (current state: %s)", sourceInfo.State)
	}

	// Track if we paused the source VM (so we can resume it on cleanup)
	wasPaused := sourceInfo.State == VmStatePaused

	// Step 1: Pause source VM if running
	if !wasPaused {
		log.Infof("Pausing source VM for clone")
		if _, err := c.apiRequest(ctx, opts.SourceSandboxId, http.MethodPut, "vm.pause", nil); err != nil {
			return nil, fmt.Errorf("failed to pause source VM: %w", err)
		}
	}

	// Ensure we resume source VM on any error
	defer func() {
		if !wasPaused {
			log.Infof("Resuming source VM after clone")
			if _, err := c.apiRequest(ctx, opts.SourceSandboxId, http.MethodPut, "vm.resume", nil); err != nil {
				log.Warnf("Failed to resume source VM: %v", err)
			}
		}
	}()

	// Step 2: Create temporary snapshot (memory + device state)
	snapshotName := fmt.Sprintf("clone-%s-%d", opts.SourceSandboxId, time.Now().UnixNano())
	snapshotPath := filepath.Join(c.config.SnapshotsPath, snapshotName)

	log.Infof("Creating memory snapshot for clone")

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

	// Step 3: Get the backing file from source disk and create independent copy
	targetSandboxDir := c.getSandboxDir(opts.NewSandboxId)
	targetDiskPath := c.getDiskPath(opts.NewSandboxId)

	log.Infof("Creating disk copy for clone")

	// Get the backing file path from source disk
	backingFileCmd := fmt.Sprintf(`qemu-img info -U --output=json "%s" | jq -r '.["backing-filename"] // empty'`, sourceDiskPath)
	backingFile, err := c.runShellScript(ctx, backingFileCmd)
	if err != nil {
		_ = c.runCommand(ctx, "rm", "-rf", snapshotPath)
		return nil, fmt.Errorf("failed to get backing file: %w", err)
	}
	backingFile = strings.TrimSpace(backingFile)

	// Create target sandbox directory
	if err := c.runCommand(ctx, "mkdir", "-p", targetSandboxDir); err != nil {
		_ = c.runCommand(ctx, "rm", "-rf", snapshotPath)
		return nil, fmt.Errorf("failed to create target sandbox directory: %w", err)
	}

	// Copy disk file directly using cp (bypasses qemu locking)
	// The copy will have the same backing file reference as the source
	log.Infof("Copying disk overlay file")
	if backingFile != "" {
		log.Infof("Source disk has backing file: %s", backingFile)
	}

	copyCmd := fmt.Sprintf(`cp "%s" "%s"`, sourceDiskPath, targetDiskPath)
	if output, err := c.runShellScript(ctx, copyCmd); err != nil {
		_ = c.runCommand(ctx, "rm", "-rf", snapshotPath)
		_ = c.runCommand(ctx, "rm", "-rf", targetSandboxDir)
		return nil, fmt.Errorf("failed to copy disk: %w (output: %s)", err, output)
	}

	log.Infof("Disk copy completed for clone")

	// Step 3b: Patch the snapshot config to use the new disk path
	log.Infof("Patching snapshot config to use new disk")
	patchConfigCmd := fmt.Sprintf(`
config_file="%s/config.json"
if [ -f "$config_file" ]; then
    jq '.disks[0].path = "%s"' "$config_file" > "$config_file.new" && mv "$config_file.new" "$config_file"
    echo "Patched disk path in snapshot config"
else
    echo "No config.json found in snapshot"
fi
`, snapshotPath, targetDiskPath)

	if output, err := c.runShellScript(ctx, patchConfigCmd); err != nil {
		log.Warnf("Failed to patch snapshot config: %v (output: %s)", err, output)
	}

	// Step 4: Create network namespace for clone
	netns, err := c.netnsPool.Create(ctx, opts.NewSandboxId)
	if err != nil {
		_ = c.runCommand(ctx, "rm", "-rf", snapshotPath)
		_ = c.runCommand(ctx, "rm", "-rf", targetSandboxDir)
		return nil, fmt.Errorf("failed to create network namespace: %w", err)
	}
	log.Infof("Created network namespace %s for clone", netns.NamespaceName)

	// Store guest IP for proxy routing
	ipFilePath := filepath.Join(targetSandboxDir, "ip")
	_ = c.runCommand(ctx, "sh", "-c", fmt.Sprintf("echo '%s' > %s", netns.GuestIP, ipFilePath))
	GetIPCache().Set(opts.NewSandboxId, netns.GuestIP)

	// Step 5: Start cloud-hypervisor process in namespace
	log.Infof("Starting cloud-hypervisor for clone in namespace %s", netns.NamespaceName)

	targetSocketPath = c.getSocketPath(opts.NewSandboxId)
	logPath := filepath.Join(targetSandboxDir, "cloud-hypervisor.log")

	startCmd := fmt.Sprintf(`
nohup ip netns exec %s cloud-hypervisor --api-socket %s > %s 2>&1 &
pid=$!

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
		c.cleanupClone(ctx, opts.NewSandboxId, snapshotPath)
		return nil, fmt.Errorf("failed to start CH for clone: %v (output: %s)", err, output)
	}

	// Step 6: Restore VM from snapshot (with memory state)
	log.Infof("Restoring cloned VM from snapshot")

	if !c.IsRemote() {
		// Get the network device ID from source VM config for proper FD mapping
		netId := "_net0"
		if sourceInfo != nil && sourceInfo.Config != nil && len(sourceInfo.Config.Net) > 0 {
			if sourceInfo.Config.Net[0].Id != "" {
				netId = sourceInfo.Config.Net[0].Id
			}
		}

		// Local mode: Use FD passing for true live clone
		if err := c.restoreWithNetFds(ctx, opts.NewSandboxId, snapshotPath, netns, netId, false); err != nil {
			c.cleanupClone(ctx, opts.NewSandboxId, snapshotPath)
			return nil, fmt.Errorf("failed to restore cloned VM: %w", err)
		}
	} else {
		// Remote mode: Fall back to standard restore
		restoreConfig := RestoreConfig{
			SourceUrl: fmt.Sprintf("file://%s", snapshotPath),
			Prefault:  false,
		}
		if _, err := c.apiRequest(ctx, opts.NewSandboxId, http.MethodPut, "vm.restore", restoreConfig); err != nil {
			c.cleanupClone(ctx, opts.NewSandboxId, snapshotPath)
			return nil, fmt.Errorf("failed to restore cloned VM: %w", err)
		}
	}

	// Step 7: Resume clone if it restored as paused
	cloneInfo, err := c.GetInfo(ctx, opts.NewSandboxId)
	if err != nil {
		log.Warnf("Failed to get clone VM info after restore: %v", err)
	} else if cloneInfo.State == VmStatePaused {
		log.Infof("Resuming cloned VM")
		if _, err := c.apiRequest(ctx, opts.NewSandboxId, http.MethodPut, "vm.resume", nil); err != nil {
			log.Warnf("Failed to resume cloned VM: %v", err)
		}
	}

	// Step 8: Clean up temporary snapshot memory files (keep config for debugging)
	log.Infof("Cleaning up temporary snapshot memory files")
	_ = c.runCommand(ctx, "sh", "-c", fmt.Sprintf("rm -f %s/memory* 2>/dev/null || true", snapshotPath))

	// Store source ID for tracking
	sourceFilePath := filepath.Join(targetSandboxDir, "source")
	_ = c.runCommand(ctx, "sh", "-c", fmt.Sprintf("echo '%s' > %s", opts.SourceSandboxId, sourceFilePath))

	// Save config.json for future operations
	if cloneInfo != nil && cloneInfo.Config != nil {
		configPath := c.getConfigPath(opts.NewSandboxId)
		configJSON, err := json.MarshalIndent(cloneInfo.Config, "", "  ")
		if err == nil {
			_ = c.writeFile(ctx, configPath, configJSON)
		}
	}

	log.Infof("Clone completed: %s -> %s (with memory state)", opts.SourceSandboxId, opts.NewSandboxId)

	return c.GetSandboxInfo(ctx, opts.NewSandboxId)
}

// cleanupClone cleans up resources on clone failure
func (c *Client) cleanupClone(ctx context.Context, sandboxId, snapshotPath string) {
	log.Warnf("Cleaning up failed clone for %s", sandboxId)

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

// GetSourceSandboxId returns the source sandbox ID for a cloned VM
func (c *Client) GetSourceSandboxId(ctx context.Context, sandboxId string) (string, error) {
	sourceFilePath := filepath.Join(c.getSandboxDir(sandboxId), "source")
	output, err := c.runShellScript(ctx, fmt.Sprintf("cat %s 2>/dev/null", sourceFilePath))
	if err != nil {
		return "", nil // Not a clone, no source
	}
	return strings.TrimSpace(output), nil
}

// IsClone returns true if the sandbox is a clone of another sandbox
func (c *Client) IsClone(ctx context.Context, sandboxId string) bool {
	source, _ := c.GetSourceSandboxId(ctx, sandboxId)
	return source != ""
}
