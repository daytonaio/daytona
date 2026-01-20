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

	// Create sandbox directory
	sandboxDir := c.getSandboxDir(opts.SandboxId)
	if err := c.runCommand(ctx, "mkdir", "-p", sandboxDir); err != nil {
		return nil, fmt.Errorf("failed to create sandbox directory: %w", err)
	}

	// Create disk image
	diskPath, err := c.createDisk(ctx, opts)
	if err != nil {
		c.cleanupSandbox(ctx, opts.SandboxId)
		return nil, fmt.Errorf("failed to create disk: %w", err)
	}

	// Create TAP interface
	tapName := c.getTapName(opts.SandboxId)
	if err := c.createTapInterface(ctx, tapName); err != nil {
		c.cleanupSandbox(ctx, opts.SandboxId)
		return nil, fmt.Errorf("failed to create TAP interface: %w", err)
	}

	// Generate MAC address
	mac := c.generateMAC(opts.SandboxId)

	// Build VM configuration
	vmConfig := c.buildVmConfig(opts, diskPath, tapName, mac)

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

	// Start Cloud Hypervisor process
	if err := c.startVMProcess(ctx, opts.SandboxId); err != nil {
		c.cleanupSandbox(ctx, opts.SandboxId)
		return nil, fmt.Errorf("failed to start VM process: %w", err)
	}

	// Wait for API socket to be ready
	if err := c.waitForSocket(ctx, opts.SandboxId, 30*time.Second); err != nil {
		c.cleanupSandbox(ctx, opts.SandboxId)
		return nil, fmt.Errorf("failed to wait for API socket: %w", err)
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

	log.Infof("Sandbox %s created and booted successfully", opts.SandboxId)

	return c.GetSandboxInfo(ctx, opts.SandboxId)
}

// createDisk creates the disk image for a sandbox
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

	// For Cloud Hypervisor, we create a qcow2 overlay with backing file
	// This is copy-on-write, so very fast and space-efficient
	qcowPath := filepath.Join(c.getSandboxDir(opts.SandboxId), "disk.qcow2")

	// Create qcow2 overlay with backing file
	// The backing file format is auto-detected by qemu-img
	createCmd := fmt.Sprintf("qemu-img create -f qcow2 -F qcow2 -b '%s' '%s' %dG",
		baseImage, qcowPath, opts.StorageGB)

	log.Debugf("Running: %s", createCmd)
	if err := c.runCommand(ctx, "sh", "-c", createCmd); err != nil {
		return "", fmt.Errorf("failed to create overlay disk: %w", err)
	}

	// For better CH performance, convert to raw format
	// This takes longer but gives better runtime performance
	log.Infof("Converting disk to raw format for better performance")
	convertCmd := fmt.Sprintf("qemu-img convert -p -O raw '%s' '%s'", qcowPath, diskPath)

	if err := c.runCommand(ctx, "sh", "-c", convertCmd); err != nil {
		// If conversion fails, keep qcow2 (CH supports both)
		log.Warnf("Failed to convert to raw (will use qcow2): %v", err)
		// Rename qcow2 to be the disk path
		if err := c.runCommand(ctx, "mv", qcowPath, diskPath); err != nil {
			return "", fmt.Errorf("failed to move qcow2 disk: %w", err)
		}
		return diskPath, nil
	}

	// Remove qcow2 overlay after successful conversion
	_ = c.runCommand(ctx, "rm", "-f", qcowPath)

	log.Infof("Disk created: %s", diskPath)
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
func (c *Client) buildVmConfig(opts CreateOptions, diskPath, tapName, mac string) *VmConfig {
	// Build payload config
	// Cloud Hypervisor supports two boot methods:
	// 1. Firmware boot (hypervisor-fw): boots like a BIOS, works with cloud images
	// 2. Direct kernel boot: requires PVH-enabled kernel
	// We use firmware boot by default since it works with standard cloud images
	payload := &PayloadConfig{
		Firmware: c.config.FirmwarePath,
	}

	// If kernel path is explicitly provided and KernelArgs are set, use kernel boot
	if opts.KernelArgs != "" && c.config.KernelPath != "" {
		cmdline := "console=hvc0 root=/dev/vda1 rw"
		cmdline += " " + opts.KernelArgs
		payload = &PayloadConfig{
			Kernel:  c.config.KernelPath,
			Cmdline: cmdline,
		}
	}

	config := &VmConfig{
		Payload: payload,
		Cpus: &CpusConfig{
			BootVcpus: opts.Cpus,
			MaxVcpus:  opts.Cpus * 2, // Allow hotplug up to 2x
		},
		Memory: &MemoryConfig{
			Size:          opts.MemoryMB * 1024 * 1024, // Convert MB to bytes
			HotplugMethod: "VirtioMem",
			HotplugSize:   ptrUint64(opts.MemoryMB * 1024 * 1024 * 2), // Allow 2x hotplug
			Thp:           true,
		},
		Disks: []DiskConfig{
			{
				Path:   diskPath,
				Direct: true, // Use O_DIRECT for better performance
			},
		},
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
			c.config.SSHHost,
			fmt.Sprintf("nohup %s", cmdStr),
		)
		return cmd.Run()
	}

	// Local execution
	cmd := exec.CommandContext(ctx, "sh", "-c", cmdStr)
	return cmd.Run()
}

// waitForSocket waits for the API socket to be ready
func (c *Client) waitForSocket(ctx context.Context, sandboxId string, timeout time.Duration) error {
	socketPath := c.getSocketPath(sandboxId)
	deadline := time.Now().Add(timeout)

	log.Infof("Waiting for API socket %s", socketPath)

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		exists, _ := c.fileExists(ctx, socketPath)
		if exists {
			// Give it a moment to be fully ready
			time.Sleep(500 * time.Millisecond)
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

	// Delete TAP interface
	tapName := c.getTapName(sandboxId)
	_ = c.deleteTapInterface(ctx, tapName)

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
