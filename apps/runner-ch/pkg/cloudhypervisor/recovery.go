// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package cloudhypervisor

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// SandboxConfig stores the configuration needed to recover a VM
// This is persisted to disk so VMs can be recovered after CH restart
type SandboxConfig struct {
	SandboxId      string            `json:"sandboxId"`
	Cpus           int               `json:"cpus"`
	MemoryMB       uint64            `json:"memoryMB"`
	StorageGB      int               `json:"storageGB"`
	Snapshot       string            `json:"snapshot,omitempty"`
	OrganizationId string            `json:"organizationId,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
	NetNSId        int               `json:"netnsId,omitempty"`
	ExternalIP     string            `json:"externalIP,omitempty"`
	GuestIP        string            `json:"guestIP,omitempty"`
	CreatedAt      time.Time         `json:"createdAt"`
	LastState      VmState           `json:"lastState,omitempty"`
}

// SaveSandboxConfig persists the sandbox configuration to disk
func (c *Client) SaveSandboxConfig(ctx context.Context, sandboxId string, config SandboxConfig) error {
	configPath := filepath.Join(c.getSandboxDir(sandboxId), "config.json")

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write config file
	script := fmt.Sprintf("cat > %s << 'EOF'\n%s\nEOF", configPath, string(data))
	if _, err := c.runShellScript(ctx, script); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	log.Debugf("Saved sandbox config for %s", sandboxId)
	return nil
}

// LoadSandboxConfig loads the sandbox configuration from disk
func (c *Client) LoadSandboxConfig(ctx context.Context, sandboxId string) (*SandboxConfig, error) {
	configPath := filepath.Join(c.getSandboxDir(sandboxId), "config.json")

	output, err := c.runShellScript(ctx, fmt.Sprintf("cat %s 2>/dev/null", configPath))
	if err != nil {
		return nil, fmt.Errorf("config file not found: %w", err)
	}

	var config SandboxConfig
	if err := json.Unmarshal([]byte(output), &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}

// UpdateSandboxState updates the last known state in the config
func (c *Client) UpdateSandboxState(ctx context.Context, sandboxId string, state VmState) error {
	config, err := c.LoadSandboxConfig(ctx, sandboxId)
	if err != nil {
		// Config doesn't exist, nothing to update
		return nil
	}

	config.LastState = state
	return c.SaveSandboxConfig(ctx, sandboxId, *config)
}

// RecoverOrphanedSandboxes scans for sandbox directories without running CH processes
// and attempts to recover them. This should be called on runner startup.
func (c *Client) RecoverOrphanedSandboxes(ctx context.Context) error {
	log.Info("Scanning for orphaned sandboxes to recover...")

	// List all sandbox directories
	sandboxDirs, err := c.listSandboxDirectories(ctx)
	if err != nil {
		return fmt.Errorf("failed to list sandbox directories: %w", err)
	}

	// List running VMs (by socket)
	runningSockets, err := c.listRunningSockets(ctx)
	if err != nil {
		return fmt.Errorf("failed to list running sockets: %w", err)
	}

	runningSet := make(map[string]bool)
	for _, sock := range runningSockets {
		// Extract sandbox ID from socket name (e.g., "abc123.sock" -> "abc123")
		sandboxId := strings.TrimSuffix(sock, ".sock")
		runningSet[sandboxId] = true
	}

	recovered := 0
	failed := 0

	for _, sandboxId := range sandboxDirs {
		// Skip if already running
		if runningSet[sandboxId] {
			log.Debugf("Sandbox %s is already running, skipping", sandboxId)
			continue
		}

		// Skip special directories
		if sandboxId == ".stats" || sandboxId == "" {
			continue
		}

		// Check if disk exists (required for recovery)
		diskPath := c.getDiskPath(sandboxId)
		exists, _ := c.fileExists(ctx, diskPath)
		if !exists {
			log.Warnf("Sandbox %s has no disk, skipping recovery", sandboxId)
			continue
		}

		log.Infof("Found orphaned sandbox %s, attempting recovery...", sandboxId)

		if err := c.recoverSandbox(ctx, sandboxId); err != nil {
			log.Errorf("Failed to recover sandbox %s: %v", sandboxId, err)
			failed++
		} else {
			log.Infof("Successfully recovered sandbox %s", sandboxId)
			recovered++
		}
	}

	log.Infof("Recovery complete: %d recovered, %d failed", recovered, failed)
	return nil
}

// recoverSandbox attempts to recover a single orphaned sandbox
// If a checkpoint exists (from Stop with memory snapshot), it restores from that
// Otherwise, it performs a cold boot
func (c *Client) recoverSandbox(ctx context.Context, sandboxId string) error {
	// Try to load saved config
	config, err := c.LoadSandboxConfig(ctx, sandboxId)
	if err != nil {
		// No config - try to recover with defaults
		log.Warnf("No config found for %s, using defaults", sandboxId)
		config = &SandboxConfig{
			SandboxId: sandboxId,
			Cpus:      c.config.DefaultCpus,
			MemoryMB:  c.config.DefaultMemoryMB,
			StorageGB: 20,
		}
	}

	// Step 1: Recover or recreate network namespace
	netns, err := c.recoverNetworkNamespace(ctx, sandboxId, config)
	if err != nil {
		return fmt.Errorf("failed to recover network namespace: %w", err)
	}

	// Check if we have a checkpoint (memory snapshot from Stop)
	hasCheckpoint := c.hasCheckpoint(ctx, sandboxId)
	if hasCheckpoint {
		log.Infof("Found checkpoint for %s - restoring with memory state", sandboxId)
		if err := c.restoreFromCheckpoint(ctx, sandboxId, config, netns); err != nil {
			log.Warnf("Checkpoint restore failed: %v - falling back to cold boot", err)
			// Delete the bad checkpoint
			_ = c.deleteCheckpoint(ctx, sandboxId)
			hasCheckpoint = false
		}
	}

	if !hasCheckpoint {
		// Cold boot path
		log.Infof("No checkpoint for %s - performing cold boot", sandboxId)

		// Step 2: Start cloud-hypervisor process in the namespace
		if err := c.startRecoveredVM(ctx, sandboxId, config, netns); err != nil {
			return fmt.Errorf("failed to start recovered VM: %w", err)
		}

		// Step 3: Wait for VM to be ready
		if err := c.waitForSocket(ctx, sandboxId, 30*time.Second); err != nil {
			return fmt.Errorf("VM socket not ready: %w", err)
		}

		// Step 4: Boot the VM (it starts in Created state)
		if _, err := c.apiRequest(ctx, sandboxId, "PUT", "vm.boot", nil); err != nil {
			return fmt.Errorf("failed to boot VM: %w", err)
		}

		// Step 5: Wait for running state
		if err := c.waitForState(ctx, sandboxId, VmStateRunning, 60*time.Second); err != nil {
			return fmt.Errorf("VM failed to reach running state: %w", err)
		}
	}

	// Step 6: Update IP cache
	if netns != nil {
		GetIPCache().Set(sandboxId, netns.ExternalIP)
	}

	return nil
}

// restoreFromCheckpoint restores a VM from a memory checkpoint
func (c *Client) restoreFromCheckpoint(ctx context.Context, sandboxId string, config *SandboxConfig, netns *NetNamespace) error {
	socketPath := c.getSocketPath(sandboxId)
	logPath := filepath.Join(c.getSandboxDir(sandboxId), "cloud-hypervisor.log")
	checkpointPath := c.getCheckpointPath(sandboxId)

	log.Infof("Restoring %s from checkpoint at %s", sandboxId, checkpointPath)

	// Step 1: Start cloud-hypervisor in the namespace
	startScript := fmt.Sprintf(`
# Start cloud-hypervisor in namespace background
nohup ip netns exec %s cloud-hypervisor --api-socket %s > %s 2>&1 &
pid=$!

# Wait for socket with fast polling
timeout=100
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
`, netns.NamespaceName, socketPath, logPath, socketPath)

	output, err := c.runShellScript(ctx, startScript)
	if err != nil || strings.Contains(output, "TIMEOUT") {
		return fmt.Errorf("failed to start CH: %v (output: %s)", err, output)
	}

	// Step 2: Restore from checkpoint using vm.restore
	restoreConfig := RestoreConfig{
		SourceUrl: fmt.Sprintf("file://%s", checkpointPath),
		Prefault:  true, // Prefault pages for faster access
	}

	if _, err := c.apiRequest(ctx, sandboxId, "PUT", "vm.restore", restoreConfig); err != nil {
		return fmt.Errorf("vm.restore failed: %w", err)
	}

	// Step 3: Wait for VM to be running or paused (restore puts it in paused state)
	time.Sleep(500 * time.Millisecond) // Give it a moment to restore

	info, err := c.GetInfo(ctx, sandboxId)
	if err != nil {
		return fmt.Errorf("failed to get VM state after restore: %w", err)
	}

	// Step 4: Resume if paused (checkpoints restore to paused state)
	if info.State == VmStatePaused {
		log.Infof("VM restored as paused, resuming...")
		if _, err := c.apiRequest(ctx, sandboxId, "PUT", "vm.resume", nil); err != nil {
			return fmt.Errorf("failed to resume after restore: %w", err)
		}
	}

	// Step 5: Wait for running state
	if err := c.waitForState(ctx, sandboxId, VmStateRunning, 30*time.Second); err != nil {
		return fmt.Errorf("VM failed to reach running state after restore: %w", err)
	}

	// Step 6: Clean up checkpoint (no longer needed after successful restore)
	log.Infof("Checkpoint restore successful, cleaning up checkpoint")
	if err := c.deleteCheckpoint(ctx, sandboxId); err != nil {
		log.Warnf("Failed to clean up checkpoint: %v", err)
	}

	log.Infof("Sandbox %s restored from checkpoint successfully", sandboxId)
	return nil
}

// recoverNetworkNamespace recovers or recreates the network namespace for a sandbox
func (c *Client) recoverNetworkNamespace(ctx context.Context, sandboxId string, config *SandboxConfig) (*NetNamespace, error) {
	shortId := sandboxId[:8]
	nsName := fmt.Sprintf("ns-%s", shortId)

	// Check if namespace already exists
	output, err := c.runShellScript(ctx, fmt.Sprintf("ip netns list | grep -w %s || true", nsName))
	if err == nil && strings.Contains(output, nsName) {
		log.Infof("Network namespace %s still exists", nsName)

		// Try to get existing namespace info from pool
		netns := c.netnsPool.Get(sandboxId)
		if netns != nil {
			return netns, nil
		}

		// Namespace exists but not in pool - reconstruct info
		// Read external IP from namespace
		ipOutput, _ := c.runShellScript(ctx, fmt.Sprintf(
			"ip netns exec %s ip -4 addr show veth-%s-host 2>/dev/null | grep inet | awk '{print $2}' | cut -d/ -f1",
			nsName, shortId))
		externalIP := strings.TrimSpace(ipOutput)
		if externalIP == "" {
			externalIP = config.ExternalIP
		}

		return &NetNamespace{
			NamespaceName: nsName,
			TapName:       "tap0", // TAP inside namespace is always tap0
			ExternalIP:    externalIP,
			GuestIP:       "192.168.0.2",
			GatewayIP:     "192.168.0.1",
		}, nil
	}

	// Namespace doesn't exist - need to recreate it
	log.Infof("Recreating network namespace for %s", sandboxId)

	// Allocate a new namespace
	netns, err := c.netnsPool.Create(ctx, sandboxId)
	if err != nil {
		return nil, fmt.Errorf("failed to create namespace: %w", err)
	}

	// Update config with new namespace info
	config.NetNSId = netns.ExternalNum
	config.ExternalIP = netns.ExternalIP
	config.GuestIP = netns.GuestIP
	if err := c.SaveSandboxConfig(ctx, sandboxId, *config); err != nil {
		log.Warnf("Failed to update config: %v", err)
	}

	return netns, nil
}

// startRecoveredVM starts a cloud-hypervisor process for a recovered sandbox
func (c *Client) startRecoveredVM(ctx context.Context, sandboxId string, config *SandboxConfig, netns *NetNamespace) error {
	socketPath := c.getSocketPath(sandboxId)
	logPath := filepath.Join(c.getSandboxDir(sandboxId), "cloud-hypervisor.log")
	diskPath := c.getDiskPath(sandboxId)

	// Build VM configuration
	vmConfig := c.buildVMConfig(sandboxId, diskPath, config.Cpus, config.MemoryMB, netns)

	// Serialize config for CH
	configJSON, err := json.Marshal(vmConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal VM config: %w", err)
	}

	// Write config to temp file
	configPath := filepath.Join(c.getSandboxDir(sandboxId), "vm-config.json")
	writeScript := fmt.Sprintf("cat > %s << 'VMCONFIG'\n%s\nVMCONFIG", configPath, string(configJSON))
	if _, err := c.runShellScript(ctx, writeScript); err != nil {
		return fmt.Errorf("failed to write VM config: %w", err)
	}

	// Start cloud-hypervisor in the namespace
	startScript := fmt.Sprintf(`
# Start cloud-hypervisor in namespace background
nohup ip netns exec %s cloud-hypervisor --api-socket %s > %s 2>&1 &
pid=$!

# Wait for socket with fast polling
timeout=100
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
`, netns.NamespaceName, socketPath, logPath, socketPath)

	output, err := c.runShellScript(ctx, startScript)
	if err != nil || strings.Contains(output, "TIMEOUT") {
		return fmt.Errorf("failed to start CH: %v (output: %s)", err, output)
	}

	// Create VM with config
	if _, err := c.apiRequest(ctx, sandboxId, "PUT", "vm.create", vmConfig); err != nil {
		return fmt.Errorf("failed to create VM: %w", err)
	}

	return nil
}

// buildVMConfig builds a VmConfig for recovery
func (c *Client) buildVMConfig(sandboxId, diskPath string, cpus int, memoryMB uint64, netns *NetNamespace) VmConfig {
	// Use the namespace's TAP name (tap0 inside the namespace)
	// NOT tap-{sandboxId} which is a different interface
	tapName := "tap0" // Default for namespace-based networking
	if netns != nil && netns.TapName != "" {
		tapName = netns.TapName
	}

	config := VmConfig{
		Payload: &PayloadConfig{
			Kernel:    c.config.KernelPath,
			Cmdline:   "console=ttyS0 root=/dev/vda1 rw",
			Initramfs: c.config.InitramfsPath,
		},
		Cpus: &CpusConfig{
			BootVcpus: cpus,
			MaxVcpus:  cpus,
		},
		Memory: &MemoryConfig{
			Size: memoryMB * 1024 * 1024,
		},
		Disks: []DiskConfig{
			{
				Path: diskPath,
			},
		},
		Net: []NetConfig{
			{
				Tap: tapName,
			},
		},
		Serial: &ConsoleConfig{
			Mode: "Tty",
		},
		Console: &ConsoleConfig{
			Mode: "Off",
		},
		Rng: &RngConfig{
			Src: "/dev/urandom",
		},
	}

	return config
}

// listSandboxDirectories returns all sandbox directory names
func (c *Client) listSandboxDirectories(ctx context.Context) ([]string, error) {
	output, err := c.runShellScript(ctx, fmt.Sprintf("ls -1 %s 2>/dev/null || true", c.config.SandboxesPath))
	if err != nil {
		return nil, err
	}

	var dirs []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			dirs = append(dirs, line)
		}
	}
	return dirs, nil
}

// listRunningSockets returns all socket files in the sockets directory
func (c *Client) listRunningSockets(ctx context.Context) ([]string, error) {
	output, err := c.runShellScript(ctx, fmt.Sprintf("ls -1 %s 2>/dev/null || true", c.config.SocketsPath))
	if err != nil {
		return nil, err
	}

	var sockets []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasSuffix(line, ".sock") {
			sockets = append(sockets, line)
		}
	}
	return sockets, nil
}

// CanRecover checks if a sandbox can be recovered
func (c *Client) CanRecover(ctx context.Context, sandboxId string) (bool, string) {
	// Check if socket exists (VM is running)
	socketPath := c.getSocketPath(sandboxId)
	if exists, _ := c.fileExists(ctx, socketPath); exists {
		return false, "VM is already running"
	}

	// Check if disk exists
	diskPath := c.getDiskPath(sandboxId)
	if exists, _ := c.fileExists(ctx, diskPath); !exists {
		return false, "disk not found"
	}

	return true, ""
}

// RecoverSandbox is the public API to recover a specific sandbox
// This can be called when StartVM fails due to missing socket
func (c *Client) RecoverSandbox(ctx context.Context, sandboxId string) error {
	mutex := c.getSandboxMutex(sandboxId)
	mutex.Lock()
	defer mutex.Unlock()

	canRecover, reason := c.CanRecover(ctx, sandboxId)
	if !canRecover {
		return fmt.Errorf("cannot recover sandbox: %s", reason)
	}

	return c.recoverSandbox(ctx, sandboxId)
}

// IsSocketMissing checks if the VM socket is missing (CH process died)
func (c *Client) IsSocketMissing(ctx context.Context, sandboxId string) bool {
	socketPath := c.getSocketPath(sandboxId)
	exists, _ := c.fileExists(ctx, socketPath)
	return !exists
}

// Local file operations for local mode
func (c *Client) readLocalFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (c *Client) writeLocalFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}
