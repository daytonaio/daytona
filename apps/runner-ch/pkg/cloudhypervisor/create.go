// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package cloudhypervisor

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// CreateOptions specifies options for creating a new VM
type CreateOptions struct {
	SandboxId  string
	Cpus       int
	MemoryMB   uint64
	StorageGB  int
	Snapshot   string            // Base snapshot to use (optional)
	GpuDevices []string          // VFIO device paths for GPU passthrough
	Metadata   map[string]string // Custom metadata
	KernelArgs string            // Additional kernel command line arguments
}

// CreateWithOptions creates a new VM sandbox using CreateOptions
func (c *Client) CreateWithOptions(ctx context.Context, opts CreateOptions) (*SandboxInfo, error) {
	mutex := c.getSandboxMutex(opts.SandboxId)
	mutex.Lock()
	defer mutex.Unlock()

	log.Infof("Creating sandbox %s: cpus=%d, memoryMB=%d, storageGB=%d, snapshot=%s",
		opts.SandboxId, opts.Cpus, opts.MemoryMB, opts.StorageGB, opts.Snapshot)

	// Check if sandbox already exists (by socket OR disk)
	// This prevents accidental deletion of existing sandboxes when CH process died
	socketPath := c.getSocketPath(opts.SandboxId)
	socketExists, _ := c.fileExists(ctx, socketPath)
	if socketExists {
		log.Infof("Sandbox %s already exists (socket found), returning existing info", opts.SandboxId)
		return c.GetSandboxInfo(ctx, opts.SandboxId)
	}

	// Also check if disk exists - sandbox may exist but CH process died
	diskPath := c.getDiskPath(opts.SandboxId)
	diskExists, _ := c.fileExists(ctx, diskPath)
	if diskExists {
		log.Warnf("Sandbox %s disk exists but socket missing - sandbox needs recovery, not creation", opts.SandboxId)
		return nil, fmt.Errorf("sandbox %s exists but needs recovery (disk found, socket missing) - use start instead of create", opts.SandboxId)
	}

	// Check if this is a warm snapshot (has memory state for instant restore)
	if opts.Snapshot != "" && c.isWarmSnapshot(ctx, opts.Snapshot) {
		log.Infof("Detected warm snapshot %s - using instant restore", opts.Snapshot)
		return c.createFromWarmSnapshot(ctx, opts)
	}

	// Set defaults
	if opts.Cpus <= 0 {
		opts.Cpus = c.config.DefaultCpus
	}
	if opts.MemoryMB <= 0 {
		opts.MemoryMB = c.config.DefaultMemoryMB
	}
	if opts.StorageGB <= 0 {
		opts.StorageGB = 20
	}

	// Enforce minimum memory of 1 GB
	const minMemoryMB uint64 = 1024
	if opts.MemoryMB < minMemoryMB {
		log.Warnf("Memory %d MB is below minimum %d MB, using minimum", opts.MemoryMB, minMemoryMB)
		opts.MemoryMB = minMemoryMB
	}

	// Cloud Hypervisor's virtio-mem requires memory to be aligned to 128 MiB
	// Round up to the nearest 128 MiB (131072 KB = 128 * 1024)
	const alignmentMB uint64 = 128
	if opts.MemoryMB%alignmentMB != 0 {
		opts.MemoryMB = ((opts.MemoryMB / alignmentMB) + 1) * alignmentMB
		log.Infof("Memory aligned to %d MB (128 MiB boundary)", opts.MemoryMB)
	}

	// Create sandbox directory and disk in one batched SSH call (for speed)
	diskPath, err := c.createDiskBatched(ctx, opts)
	if err != nil {
		c.cleanupSandbox(ctx, opts.SandboxId)
		return nil, fmt.Errorf("failed to create disk: %w", err)
	}

	// Create network namespace for this sandbox
	// Each VM gets its own isolated namespace with fixed internal IP (192.168.0.2)
	netns, err := c.netnsPool.Create(ctx, opts.SandboxId)
	if err != nil {
		c.cleanupSandbox(ctx, opts.SandboxId)
		return nil, fmt.Errorf("failed to create network namespace: %w", err)
	}
	log.Infof("Created network namespace %s for %s (external: %s)", netns.NamespaceName, opts.SandboxId, netns.ExternalIP)

	// TAP is created inside the namespace (tap0), use that name
	tapName := netns.TapName

	// Use fixed guest IP (same for all VMs, namespace provides isolation)
	ip := netns.GuestIP
	log.Infof("Using fixed guest IP %s for sandbox %s (namespace: %s)", ip, opts.SandboxId, netns.NamespaceName)

	// Store namespace external IP for proxy routing
	ipFilePath := filepath.Join(c.getSandboxDir(opts.SandboxId), "ip")
	_ = c.runCommand(ctx, "sh", "-c", fmt.Sprintf("echo '%s' > %s", ip, ipFilePath))
	GetIPCache().Set(opts.SandboxId, ip)

	// Create cloud-init ISO with fixed internal network configuration
	// This is REQUIRED for network namespaces - there's no DHCP server, so VM must have static IP
	hostname := opts.SandboxId
	if len(hostname) > 12 {
		hostname = hostname[:12]
	}
	cloudInitISO, err := c.createCloudInitISO(ctx, opts.SandboxId, CloudInitConfig{
		IP:       ip,
		Gateway:  netns.GatewayIP,
		Netmask:  GuestNetmask,
		Hostname: hostname,
	})
	if err != nil {
		c.cleanupSandbox(ctx, opts.SandboxId)
		return nil, fmt.Errorf("failed to create cloud-init ISO (required for network config): %w", err)
	}

	// Generate MAC address
	mac := c.generateMAC(opts.SandboxId)

	// Build VM configuration
	vmConfig := c.buildVmConfig(opts, diskPath, tapName, mac, cloudInitISO)

	// Write config to file
	configPath := c.getConfigPath(opts.SandboxId)
	configJSON, err := json.MarshalIndent(vmConfig, "", "  ")
	if err != nil {
		c.cleanupSandbox(ctx, opts.SandboxId)
		return nil, fmt.Errorf("failed to marshal VM config: %w", err)
	}
	if err := c.writeFile(ctx, configPath, configJSON); err != nil {
		c.cleanupSandbox(ctx, opts.SandboxId)
		return nil, fmt.Errorf("failed to write VM config: %w", err)
	}

	// Start Cloud Hypervisor process inside the network namespace and wait for socket
	if err := c.startVMProcessInNamespaceAndWait(ctx, opts.SandboxId, netns, 30*time.Second); err != nil {
		c.cleanupSandbox(ctx, opts.SandboxId)
		return nil, fmt.Errorf("failed to start VM: %w", err)
	}

	// Create the VM using the API
	if _, err := c.apiRequest(ctx, opts.SandboxId, http.MethodPut, "vm.create", vmConfig); err != nil {
		c.cleanupSandbox(ctx, opts.SandboxId)
		return nil, fmt.Errorf("failed to create VM: %w", err)
	}

	// Boot the VM
	if _, err := c.apiRequest(ctx, opts.SandboxId, http.MethodPut, "vm.boot", nil); err != nil {
		c.cleanupSandbox(ctx, opts.SandboxId)
		return nil, fmt.Errorf("failed to boot VM: %w", err)
	}

	// Wait for daemon to be reachable (the VM needs time to boot and start the daemon)
	if err := c.waitForDaemon(ctx, opts.SandboxId, netns.ExternalIP, 30*time.Second); err != nil {
		log.Warnf("Daemon health check failed for %s: %v (continuing anyway)", opts.SandboxId, err)
	}

	log.Infof("Sandbox %s created successfully with IP %s", opts.SandboxId, ip)

	return c.GetSandboxInfo(ctx, opts.SandboxId)
}

// waitForIP waits for the VM to obtain an IP address via DHCP
// createDiskBatched creates sandbox directory and disk image in a single SSH call
// This reduces SSH connection overhead from ~4 calls to 1 call
func (c *Client) createDiskBatched(ctx context.Context, opts CreateOptions) (string, error) {
	sandboxDir := c.getSandboxDir(opts.SandboxId)
	diskPath := c.getDiskPath(opts.SandboxId)

	var baseImage string
	if opts.Snapshot != "" {
		// Use snapshot as base - snapshots are directories containing disk.qcow2
		baseImage = filepath.Join(c.config.SnapshotsPath, opts.Snapshot, "disk.qcow2")
	} else {
		baseImage = c.config.BaseImagePath
	}

	log.Infof("Creating disk from base image: %s", baseImage)

	// Batch all operations into a single SSH command:
	// 1. Check if disk already exists (skip if yes)
	// 2. Check if base image exists (fail if no)
	// 3. Create sandbox directory
	// 4. Create qcow2 overlay disk
	log.Infof("Creating disk: base=%s disk=%s size=%dG", baseImage, diskPath, opts.StorageGB)

	batchCmd := fmt.Sprintf(
		`if [ -f "%s" ]; then echo "EXISTS"; exit 0; fi; `+
			`if [ ! -f "%s" ]; then echo "BASE_NOT_FOUND"; exit 1; fi; `+
			`mkdir -p "%s" && `+
			`qemu-img create -f qcow2 -F qcow2 -b "%s" "%s" %dG && `+
			`echo "CREATED"`,
		diskPath, baseImage, sandboxDir, baseImage, diskPath, opts.StorageGB)

	output, err := c.runShellScript(ctx, batchCmd)
	output = strings.TrimSpace(output)

	if err != nil {
		if strings.Contains(output, "BASE_NOT_FOUND") {
			return "", fmt.Errorf("base image not found: %s", baseImage)
		}
		return "", fmt.Errorf("failed to create disk: %w (output: %s)", err, output)
	}

	if strings.Contains(output, "EXISTS") {
		log.Infof("Disk %s already exists", diskPath)
	} else {
		log.Infof("Disk created: %s (qcow2 overlay, %dGB quota)", diskPath, opts.StorageGB)
	}

	return diskPath, nil
}

// createDisk creates the disk image for a sandbox (non-batched version for compatibility)
func (c *Client) createDisk(ctx context.Context, opts CreateOptions) (string, error) {
	diskPath := c.getDiskPath(opts.SandboxId)

	// Check if disk already exists
	exists, _ := c.fileExists(ctx, diskPath)
	if exists {
		log.Infof("Disk %s already exists", diskPath)
		return diskPath, nil
	}

	var baseImage string
	if opts.Snapshot != "" {
		// Use snapshot as base - snapshots are directories containing disk.qcow2
		baseImage = filepath.Join(c.config.SnapshotsPath, opts.Snapshot, "disk.qcow2")
	} else {
		baseImage = c.config.BaseImagePath
	}

	// Check if base image exists
	baseExists, err := c.fileExists(ctx, baseImage)
	if err != nil {
		return "", fmt.Errorf("failed to check base image: %w", err)
	}
	if !baseExists {
		return "", fmt.Errorf("base image not found: %s", baseImage)
	}

	log.Infof("Creating disk from base image: %s", baseImage)

	// Create sandbox directory first
	sandboxDir := c.getSandboxDir(opts.SandboxId)
	if err := c.runCommand(ctx, "mkdir", "-p", sandboxDir); err != nil {
		return "", fmt.Errorf("failed to create sandbox directory: %w", err)
	}

	// Create qcow2 overlay with backing file
	createCmd := fmt.Sprintf("qemu-img create -f qcow2 -F qcow2 -b '%s' '%s' %dG",
		baseImage, diskPath, opts.StorageGB)

	log.Debugf("Running: %s", createCmd)
	if err := c.runCommand(ctx, "sh", "-c", createCmd); err != nil {
		return "", fmt.Errorf("failed to create overlay disk: %w", err)
	}

	log.Infof("Disk created: %s (qcow2 overlay, %dGB quota)", diskPath, opts.StorageGB)
	return diskPath, nil
}

// createTapInterface creates a TAP network interface
func (c *Client) createTapInterface(ctx context.Context, tapName string) error {
	log.Infof("Creating TAP interface %s", tapName)

	// Try using the helper script first
	if c.config.TapCreateScript != "" {
		exists, _ := c.fileExists(ctx, c.config.TapCreateScript)
		if exists {
			return c.runCommand(ctx, c.config.TapCreateScript, tapName)
		}
	}

	// Manual TAP creation
	cmds := [][]string{
		{"ip", "tuntap", "add", tapName, "mode", "tap"},
		{"ip", "link", "set", tapName, "master", c.config.BridgeName},
		{"ip", "link", "set", tapName, "up"},
	}

	for _, cmd := range cmds {
		if err := c.runCommand(ctx, cmd[0], cmd[1:]...); err != nil {
			// Try to cleanup on failure
			_ = c.runCommand(ctx, "ip", "tuntap", "del", tapName, "mode", "tap")
			return fmt.Errorf("failed to run %v: %w", cmd, err)
		}
	}

	return nil
}

// generateMAC generates a deterministic MAC address from sandbox ID
func (c *Client) generateMAC(sandboxId string) string {
	hash := sha256.Sum256([]byte(sandboxId))
	// Use locally administered, unicast MAC address
	// 02:xx:xx:xx:xx:xx format
	return fmt.Sprintf("02:%02x:%02x:%02x:%02x:%02x",
		hash[0], hash[1], hash[2], hash[3], hash[4])
}

// buildVmConfig creates a VM configuration
func (c *Client) buildVmConfig(opts CreateOptions, diskPath, tapName, mac, cloudInitISO string) *VmConfig {
	// Build payload config
	// Cloud Hypervisor supports two boot methods:
	// 1. Firmware boot (hypervisor-fw): boots like a BIOS, but doesn't set root= properly
	// 2. Direct kernel boot: requires bzImage kernel + initramfs, works reliably
	// We use kernel boot by default since firmware boot fails to set root= parameter
	cmdline := "console=ttyS0 root=LABEL=cloudimg-rootfs rw"
	if opts.KernelArgs != "" {
		cmdline += " " + opts.KernelArgs
	}

	payload := &PayloadConfig{
		Kernel:    c.config.KernelPath,
		Initramfs: c.config.InitramfsPath,
		Cmdline:   cmdline,
	}

	// Build disk list
	// Note: Direct (O_DIRECT) is disabled because it doesn't work properly
	// with qcow2 files that have backing files. The backing chain reads fail
	// with O_DIRECT enabled, causing kernel panic "unable to mount root fs".
	disks := []DiskConfig{
		{
			Path: diskPath,
		},
	}

	// Add cloud-init ISO as second disk if available
	if cloudInitISO != "" {
		disks = append(disks, DiskConfig{
			Path:     cloudInitISO,
			Readonly: true,
		})
	}

	config := &VmConfig{
		Payload: payload,
		Cpus: &CpusConfig{
			BootVcpus: opts.Cpus,
			MaxVcpus:  opts.Cpus * 2, // Allow hotplug up to 2x
		},
		Memory: &MemoryConfig{
			Size:          opts.MemoryMB * 1024 * 1024, // Convert MB to bytes (already 128 MiB aligned)
			HotplugMethod: "VirtioMem",
			HotplugSize:   ptrUint64(opts.MemoryMB * 1024 * 1024), // Hotplug pool (128 MiB aligned)
			Thp:           true,
		},
		Disks: disks,
		Net: []NetConfig{
			{
				Id:  "_net0", // Explicit ID for fork restore compatibility
				Tap: tapName,
				Mac: mac,
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
		// Memory ballooning for dynamic memory management
		// Allows host to reclaim unused memory from guest VMs
		// DeflateOnOom: auto-deflate if guest runs out of memory
		// FreePageReporting: guest proactively reports free pages to host
		Balloon: &BalloonConfig{
			Size:              0,    // Start with no inflation (VM gets full memory)
			DeflateOnOom:      true, // Auto-deflate if guest needs memory
			FreePageReporting: true, // Guest reports free pages proactively
		},
	}

	// Add GPU devices if specified
	if len(opts.GpuDevices) > 0 {
		config.Devices = make([]DeviceConfig, len(opts.GpuDevices))
		for i, device := range opts.GpuDevices {
			config.Devices[i] = DeviceConfig{
				Path:  device,
				Iommu: true,
				Id:    fmt.Sprintf("gpu%d", i),
			}
		}
		config.Iommu = true
	}

	return config
}

// startVMProcess starts the cloud-hypervisor process
func (c *Client) startVMProcess(ctx context.Context, sandboxId string) error {
	socketPath := c.getSocketPath(sandboxId)
	logPath := filepath.Join(c.getSandboxDir(sandboxId), "cloud-hypervisor.log")

	log.Infof("Starting cloud-hypervisor process for %s with socket %s", sandboxId, socketPath)

	// Build command
	cmdStr := fmt.Sprintf("cloud-hypervisor --api-socket %s > %s 2>&1 &",
		socketPath, logPath)

	if c.IsRemote() {
		// Run via SSH with nohup to keep process running
		cmd := exec.CommandContext(ctx, "ssh",
			"-i", c.config.SSHKeyPath,
			"-o", "StrictHostKeyChecking=accept-new",
			"-o", "BatchMode=yes",
			c.config.SSHHost,
			fmt.Sprintf("nohup %s", cmdStr),
		)
		return cmd.Run()
	}

	// Local execution
	cmd := exec.CommandContext(ctx, "sh", "-c", cmdStr)
	return cmd.Run()
}

// startVMProcessAndWait starts CH process and waits for socket in a single SSH call
// This saves an SSH round-trip compared to calling startVMProcess + waitForSocket separately
// DEPRECATED: Use startVMProcessInNamespaceAndWait for namespace-based networking
func (c *Client) startVMProcessAndWait(ctx context.Context, sandboxId string, timeout time.Duration) error {
	socketPath := c.getSocketPath(sandboxId)
	logPath := filepath.Join(c.getSandboxDir(sandboxId), "cloud-hypervisor.log")
	timeoutSec := int(timeout.Seconds())

	log.Infof("Starting cloud-hypervisor process for %s with socket %s", sandboxId, socketPath)

	if c.IsRemote() {
		// Combined: start CH + wait for socket in one SSH call
		// This saves ~2-3 seconds of SSH overhead
		combinedCmd := fmt.Sprintf(`
# Start cloud-hypervisor in background
nohup cloud-hypervisor --api-socket %s > %s 2>&1 &
pid=$!

# Wait for socket with fast polling
timeout=%d
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
`, socketPath, logPath, timeoutSec*20, socketPath) // *20 because we sleep 0.05s

		output, err := c.runSSHCommand(ctx, combinedCmd)
		output = strings.TrimSpace(output)

		if err != nil || strings.Contains(output, "TIMEOUT") {
			return fmt.Errorf("failed to start CH or timeout waiting for socket: %v (output: %s)", err, output)
		}

		log.Infof("cloud-hypervisor started and socket ready for %s", sandboxId)
		return nil
	}

	// Local mode: start process then wait
	if err := c.startVMProcess(ctx, sandboxId); err != nil {
		return fmt.Errorf("failed to start VM process: %w", err)
	}
	return c.waitForSocket(ctx, sandboxId, timeout)
}

// startVMProcessInNamespaceAndWait starts CH process inside a network namespace and waits for socket
// This is the preferred method for namespace-based networking
func (c *Client) startVMProcessInNamespaceAndWait(ctx context.Context, sandboxId string, netns *NetNamespace, timeout time.Duration) error {
	socketPath := c.getSocketPath(sandboxId)
	logPath := filepath.Join(c.getSandboxDir(sandboxId), "cloud-hypervisor.log")
	timeoutSec := int(timeout.Seconds())

	log.Infof("Starting cloud-hypervisor process for %s in namespace %s with socket %s", sandboxId, netns.NamespaceName, socketPath)

	// Combined: start CH in namespace + wait for socket in one SSH call
	combinedCmd := fmt.Sprintf(`
# Start cloud-hypervisor in namespace background
nohup ip netns exec %s cloud-hypervisor --api-socket %s > %s 2>&1 &
pid=$!

# Wait for socket with fast polling
timeout=%d
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
`, netns.NamespaceName, socketPath, logPath, timeoutSec*20, socketPath) // *20 because we sleep 0.05s

	output, err := c.runShellScript(ctx, combinedCmd)
	output = strings.TrimSpace(output)

	if err != nil || strings.Contains(output, "TIMEOUT") {
		return fmt.Errorf("failed to start CH in namespace or timeout waiting for socket: %v (output: %s)", err, output)
	}

	log.Infof("cloud-hypervisor started in namespace %s and socket ready for %s", netns.NamespaceName, sandboxId)
	return nil
}

// waitForSocket waits for the API socket to be ready
// Uses remote-side polling to avoid multiple SSH round-trips
func (c *Client) waitForSocket(ctx context.Context, sandboxId string, timeout time.Duration) error {
	socketPath := c.getSocketPath(sandboxId)
	timeoutSec := int(timeout.Seconds())

	log.Infof("Waiting for API socket %s", socketPath)

	if c.IsRemote() {
		// Remote polling: single SSH call that polls on the remote side
		pollCmd := fmt.Sprintf(`
timeout=%d
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
exit 1
`, timeoutSec*20, socketPath) // *20 because we sleep 0.05s

		output, err := c.runCommandOutput(ctx, "sh", "-c", pollCmd)
		output = strings.TrimSpace(output)

		if err != nil || strings.Contains(output, "TIMEOUT") {
			return fmt.Errorf("timeout waiting for socket %s", socketPath)
		}
		return nil
	}

	// Local polling (original logic)
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		exists, _ := c.fileExists(ctx, socketPath)
		if exists {
			time.Sleep(300 * time.Millisecond)
			return nil
		}

		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for socket %s", socketPath)
}

// cleanupSandbox removes sandbox resources on failure
func (c *Client) cleanupSandbox(ctx context.Context, sandboxId string) {
	log.Warnf("Cleaning up sandbox %s after failure", sandboxId)

	// Kill VM process
	_ = c.killVMProcess(ctx, sandboxId)

	// Delete network namespace (this also cleans up TAP and veth interfaces)
	_ = c.netnsPool.Delete(ctx, sandboxId)

	// Release IP from cache
	GetIPCache().Delete(sandboxId)

	// Remove sandbox directory
	sandboxDir := c.getSandboxDir(sandboxId)
	_ = c.runCommand(ctx, "rm", "-rf", sandboxDir)

	// Remove socket
	socketPath := c.getSocketPath(sandboxId)
	_ = c.runCommand(ctx, "rm", "-f", socketPath)
}

// killVMProcess kills the cloud-hypervisor process for a sandbox
func (c *Client) killVMProcess(ctx context.Context, sandboxId string) error {
	socketPath := c.getSocketPath(sandboxId)

	// Find and kill process by socket
	cmdStr := fmt.Sprintf("pkill -f 'cloud-hypervisor.*%s' || true", socketPath)
	return c.runCommand(ctx, "sh", "-c", cmdStr)
}

// deleteTapInterface removes a TAP network interface
func (c *Client) deleteTapInterface(ctx context.Context, tapName string) error {
	log.Infof("Deleting TAP interface %s", tapName)

	// Try using the helper script first
	if c.config.TapDeleteScript != "" {
		exists, _ := c.fileExists(ctx, c.config.TapDeleteScript)
		if exists {
			return c.runCommand(ctx, c.config.TapDeleteScript, tapName)
		}
	}

	// Manual deletion
	return c.runCommand(ctx, "ip", "tuntap", "del", tapName, "mode", "tap")
}

// GetSandboxInfo returns information about a sandbox
func (c *Client) GetSandboxInfo(ctx context.Context, sandboxId string) (*SandboxInfo, error) {
	vmInfo, err := c.GetInfo(ctx, sandboxId)
	if err != nil {
		return nil, err
	}

	info := &SandboxInfo{
		Id:         sandboxId,
		State:      vmInfo.State,
		SocketPath: c.getSocketPath(sandboxId),
		DiskPath:   c.getDiskPath(sandboxId),
		TapDevice:  c.getTapName(sandboxId),
		CreatedAt:  time.Now(), // Would need to track this separately
	}

	if vmInfo.Config != nil {
		if vmInfo.Config.Cpus != nil {
			info.Vcpus = vmInfo.Config.Cpus.BootVcpus
		}
		if vmInfo.Config.Memory != nil {
			info.MemoryMB = vmInfo.Config.Memory.Size / (1024 * 1024)
		}
	}

	// Get IP from pool (instant lookup)
	if ip := c.ipPool.Get(sandboxId); ip != "" {
		info.IpAddress = ip
	} else if ip := GetIPCache().Get(sandboxId); ip != "" {
		// Fallback to cache
		info.IpAddress = ip
	} else {
		// Last resort: try to read from stored file
		ipFilePath := filepath.Join(c.getSandboxDir(sandboxId), "ip")
		if output, err := c.runShellScript(ctx, fmt.Sprintf("cat %s 2>/dev/null", ipFilePath)); err == nil {
			if ip := strings.TrimSpace(output); ip != "" {
				info.IpAddress = ip
				GetIPCache().Set(sandboxId, ip)
			}
		}
	}

	return info, nil
}

// hasImageExtension checks if a path has a disk image extension
func hasImageExtension(path string) bool {
	ext := filepath.Ext(path)
	return ext == ".qcow2" || ext == ".raw" || ext == ".img"
}

// ptrUint64 returns a pointer to a uint64
func ptrUint64(v uint64) *uint64 {
	return &v
}

// isWarmSnapshot checks if a snapshot is a "warm" snapshot with memory state
// Warm snapshots contain: disk.qcow2, memory-ranges, state.json, config.json
// They allow instant VM restore without booting (3-4 seconds vs 15+ seconds)
func (c *Client) isWarmSnapshot(ctx context.Context, snapshotName string) bool {
	// Warm snapshots are directories containing memory state
	snapshotDir := filepath.Join(c.config.SnapshotsPath, snapshotName)

	// Check if it's a directory with memory-ranges file
	checkCmd := fmt.Sprintf(`[ -d "%s" ] && [ -f "%s/memory-ranges" ] && [ -f "%s/disk.qcow2" ] && echo "WARM" || echo "COLD"`,
		snapshotDir, snapshotDir, snapshotDir)

	output, err := c.runShellScript(ctx, checkCmd)
	if err != nil {
		return false
	}

	return strings.TrimSpace(output) == "WARM"
}

// createFromWarmSnapshot creates a VM by restoring from a warm snapshot
// This is much faster than cold boot (~3-4 seconds vs ~15 seconds) because:
// 1. No kernel boot required
// 2. Memory state is restored directly
// 3. Daemon is immediately available
func (c *Client) createFromWarmSnapshot(ctx context.Context, opts CreateOptions) (*SandboxInfo, error) {
	snapshotDir := filepath.Join(c.config.SnapshotsPath, opts.Snapshot)
	sandboxDir := c.getSandboxDir(opts.SandboxId)
	diskPath := c.getDiskPath(opts.SandboxId)

	log.Infof("Creating sandbox %s from warm snapshot %s (instant restore)", opts.SandboxId, opts.Snapshot)

	// Step 1: Create sandbox directory and CoW disk overlay
	// For warm snapshots, the overlay must be at least as large as the backing file
	// since the memory state references sectors from the original disk size
	goldenDisk := filepath.Join(snapshotDir, "disk.qcow2")

	// Get the backing file's virtual size to ensure overlay is large enough
	sizeCmd := fmt.Sprintf(`qemu-img info --output=json "%s" | jq -r '.["virtual-size"]'`, goldenDisk)
	sizeOutput, err := c.runShellScript(ctx, sizeCmd)
	if err != nil {
		c.cleanupSandbox(ctx, opts.SandboxId)
		return nil, fmt.Errorf("failed to get snapshot disk size: %w", err)
	}

	// Parse the size and convert to GB (round up)
	var backingFileSizeGB int64 = int64(opts.StorageGB)
	if sizeStr := strings.TrimSpace(sizeOutput); sizeStr != "" && sizeStr != "null" {
		if sizeBytes, parseErr := strconv.ParseInt(sizeStr, 10, 64); parseErr == nil {
			backingFileSizeGB = (sizeBytes + (1 << 30) - 1) / (1 << 30) // Round up to GB
		}
	}

	// Use the larger of backing file size or requested size
	diskSizeGB := backingFileSizeGB
	if int64(opts.StorageGB) > diskSizeGB {
		diskSizeGB = int64(opts.StorageGB)
	}
	log.Infof("Creating CoW overlay: backing file=%dGB, requested=%dGB, using=%dGB", backingFileSizeGB, opts.StorageGB, diskSizeGB)

	createDiskCmd := fmt.Sprintf(
		`mkdir -p "%s" && qemu-img create -f qcow2 -F qcow2 -b "%s" "%s" %dG`,
		sandboxDir, goldenDisk, diskPath, diskSizeGB)

	if _, err := c.runShellScript(ctx, createDiskCmd); err != nil {
		c.cleanupSandbox(ctx, opts.SandboxId)
		return nil, fmt.Errorf("failed to create disk overlay: %w", err)
	}
	log.Infof("Created CoW disk overlay from warm snapshot")

	// Step 2: Create network namespace
	netns, err := c.netnsPool.Create(ctx, opts.SandboxId)
	if err != nil {
		c.cleanupSandbox(ctx, opts.SandboxId)
		return nil, fmt.Errorf("failed to create network namespace: %w", err)
	}
	log.Infof("Created network namespace %s for %s", netns.NamespaceName, opts.SandboxId)

	// Store IP for proxy routing
	ip := netns.GuestIP
	ipFilePath := filepath.Join(sandboxDir, "ip")
	_ = c.runCommand(ctx, "sh", "-c", fmt.Sprintf("echo '%s' > %s", ip, ipFilePath))
	GetIPCache().Set(opts.SandboxId, ip)

	// Step 3: Create temporary snapshot directory with patched config
	// The config needs to point to the new disk path
	tempSnapshotDir := filepath.Join(sandboxDir, "temp-snapshot")
	patchCmd := fmt.Sprintf(`
mkdir -p "%s"
cp "%s/memory-ranges" "%s/"
cp "%s/state.json" "%s/"

# Patch config.json to use the new disk path and remove cloud-init disk
cat "%s/config.json" | jq '.disks[0].path = "%s"' > "%s/config.json"
`,
		tempSnapshotDir,
		snapshotDir, tempSnapshotDir,
		snapshotDir, tempSnapshotDir,
		snapshotDir, diskPath, tempSnapshotDir)

	if _, err := c.runShellScript(ctx, patchCmd); err != nil {
		c.cleanupSandbox(ctx, opts.SandboxId)
		return nil, fmt.Errorf("failed to prepare snapshot for restore: %w", err)
	}

	// Step 4: Start cloud-hypervisor process in namespace
	if err := c.startVMProcessInNamespaceAndWait(ctx, opts.SandboxId, netns, 10*time.Second); err != nil {
		c.cleanupSandbox(ctx, opts.SandboxId)
		return nil, fmt.Errorf("failed to start cloud-hypervisor: %w", err)
	}

	// Step 5: Restore VM from snapshot
	restoreConfig := RestoreConfig{
		SourceUrl: fmt.Sprintf("file://%s", tempSnapshotDir),
		Prefault:  false,
	}

	if _, err := c.apiRequest(ctx, opts.SandboxId, http.MethodPut, "vm.restore", restoreConfig); err != nil {
		c.cleanupSandbox(ctx, opts.SandboxId)
		return nil, fmt.Errorf("failed to restore VM from snapshot: %w", err)
	}
	log.Infof("VM restored from warm snapshot")

	// Step 6: Resume the VM
	if _, err := c.apiRequest(ctx, opts.SandboxId, http.MethodPut, "vm.resume", nil); err != nil {
		c.cleanupSandbox(ctx, opts.SandboxId)
		return nil, fmt.Errorf("failed to resume VM: %w", err)
	}

	// Step 7: Clean up temporary snapshot directory
	_ = c.runCommand(ctx, "rm", "-rf", tempSnapshotDir)

	// Step 8: Save config.json for future clone/fork operations
	vmInfo, err := c.GetInfo(ctx, opts.SandboxId)
	if err == nil && vmInfo.Config != nil {
		configPath := c.getConfigPath(opts.SandboxId)
		configJSON, err := json.MarshalIndent(vmInfo.Config, "", "  ")
		if err == nil {
			if err := c.writeFile(ctx, configPath, configJSON); err != nil {
				log.Warnf("Failed to save config.json for %s: %v", opts.SandboxId, err)
			}
		}
	}

	// Wait briefly for daemon to be ready (should be almost instant since memory is restored)
	if err := c.waitForDaemon(ctx, opts.SandboxId, netns.ExternalIP, 10*time.Second); err != nil {
		log.Warnf("Daemon health check failed after warm restore for %s: %v", opts.SandboxId, err)
	}

	log.Infof("Sandbox %s created from warm snapshot in instant restore mode", opts.SandboxId)

	return c.GetSandboxInfo(ctx, opts.SandboxId)
}
