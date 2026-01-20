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

	// Check if sandbox already exists
	socketPath := c.getSocketPath(opts.SandboxId)
	exists, _ := c.fileExists(ctx, socketPath)
	if exists {
		log.Infof("Sandbox %s already exists, returning existing info", opts.SandboxId)
		return c.GetSandboxInfo(ctx, opts.SandboxId)
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

	// Get TAP interface (from pool if enabled, otherwise create)
	var tapName string
	if c.tapPool.IsEnabled() {
		var err error
		tapName, err = c.tapPool.Acquire(ctx, opts.SandboxId)
		if err != nil {
			c.cleanupSandbox(ctx, opts.SandboxId)
			return nil, fmt.Errorf("failed to acquire TAP from pool: %w", err)
		}
		log.Infof("Acquired TAP %s from pool for %s", tapName, opts.SandboxId)
	} else {
		tapName = c.getTapName(opts.SandboxId)
		if err := c.createTapInterface(ctx, tapName); err != nil {
			c.cleanupSandbox(ctx, opts.SandboxId)
			return nil, fmt.Errorf("failed to create TAP interface: %w", err)
		}
	}

	// Allocate static IP from pool (instant, no DHCP wait)
	ip, err := c.ipPool.Allocate(opts.SandboxId)
	if err != nil {
		c.cleanupSandbox(ctx, opts.SandboxId)
		return nil, fmt.Errorf("failed to allocate IP: %w", err)
	}
	log.Infof("Allocated IP %s for sandbox %s", ip, opts.SandboxId)

	// Store IP immediately (no need to wait for DHCP)
	ipFilePath := filepath.Join(c.getSandboxDir(opts.SandboxId), "ip")
	c.runCommand(ctx, fmt.Sprintf("echo '%s' > %s", ip, ipFilePath))
	GetIPCache().Set(opts.SandboxId, ip)

	// Create cloud-init ISO with static IP configuration
	cloudInitISO, err := c.createCloudInitISO(ctx, opts.SandboxId, CloudInitConfig{
		IP:       ip,
		Gateway:  IPPoolGateway,
		Netmask:  IPPoolNetmask,
		Hostname: opts.SandboxId[:12], // Use first 12 chars of ID as hostname
	})
	if err != nil {
		log.Warnf("Failed to create cloud-init ISO: %v (VM will use DHCP fallback)", err)
		cloudInitISO = "" // Continue without cloud-init
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

	// Start Cloud Hypervisor process and wait for socket (combined for efficiency)
	if err := c.startVMProcessAndWait(ctx, opts.SandboxId, 30*time.Second); err != nil {
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
		// Use snapshot as base
		baseImage = filepath.Join(c.config.SnapshotsPath, opts.Snapshot)
		if !hasImageExtension(baseImage) {
			baseImage += ".qcow2"
		}
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
		// Use snapshot as base
		baseImage = filepath.Join(c.config.SnapshotsPath, opts.Snapshot)
		if !hasImageExtension(baseImage) {
			baseImage += ".qcow2"
		}
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

	// Release TAP interface (back to pool or delete)
	if c.tapPool.IsEnabled() {
		_ = c.tapPool.Release(ctx, sandboxId)
	} else {
		tapName := c.getTapName(sandboxId)
		_ = c.deleteTapInterface(ctx, tapName)
	}

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
