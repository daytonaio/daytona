// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package cuttlefish

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"
)

// CreateWithOptions creates a new Cuttlefish instance
func (c *Client) CreateWithOptions(ctx context.Context, opts CreateOptions) (*SandboxInfo, error) {
	mutex := c.getSandboxMutex(opts.SandboxId)
	mutex.Lock()
	defer mutex.Unlock()

	log.Infof("Creating sandbox %s: cpus=%d, memoryMB=%d, diskGB=%d, snapshot=%s",
		opts.SandboxId, opts.Cpus, opts.MemoryMB, opts.DiskGB, opts.Snapshot)

	// Check if sandbox already exists
	if _, exists := c.GetInstance(opts.SandboxId); exists {
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
	if opts.DiskGB <= 0 {
		opts.DiskGB = c.config.DefaultDiskGB
	}

	// Validate snapshot exists if specified
	if opts.Snapshot != "" {
		exists, err := c.SnapshotExists(ctx, opts.Snapshot)
		if err != nil {
			return nil, fmt.Errorf("failed to check snapshot: %w", err)
		}
		if !exists {
			helpMsg := c.GetSnapshotNotFoundHelp(opts.Snapshot)
			log.Error(helpMsg)
			return nil, fmt.Errorf("snapshot '%s' not found", opts.Snapshot)
		}
	}

	// Allocate an instance number
	instanceNum, err := c.allocateInstanceNum()
	if err != nil {
		return nil, fmt.Errorf("failed to allocate instance number: %w", err)
	}

	log.Infof("Allocated instance number %d for sandbox %s", instanceNum, opts.SandboxId)

	// Ensure the instance number is actually available in CVD
	// This handles stale CVD state from previous runs
	if err := c.EnsureInstanceAvailable(ctx, instanceNum); err != nil {
		return nil, fmt.Errorf("failed to ensure instance %d is available: %w", instanceNum, err)
	}

	// Create instance directory
	instanceDir := c.getInstanceDir(opts.SandboxId)
	if err := c.runCommand(ctx, "mkdir", "-p", instanceDir); err != nil {
		return nil, fmt.Errorf("failed to create instance directory: %w", err)
	}

	// Calculate ports - CVD uses base_port + (instance_num - 1)
	// Instance 1 uses port 6520, Instance 2 uses 6521, etc.
	adbPort := c.config.ADBBasePort + (instanceNum - 1)
	webrtcPort := c.config.WebRTCBasePort + (instanceNum - 1)
	adbSerial := fmt.Sprintf("0.0.0.0:%d", adbPort)

	// Store snapshot in metadata
	if opts.Metadata == nil {
		opts.Metadata = make(map[string]string)
	}
	if opts.Snapshot != "" {
		opts.Metadata["snapshot"] = opts.Snapshot
	}

	// Create instance info
	info := &InstanceInfo{
		SandboxId:   opts.SandboxId,
		InstanceNum: instanceNum,
		State:       InstanceStateStarting,
		Cpus:        opts.Cpus,
		MemoryMB:    opts.MemoryMB,
		DiskGB:      opts.DiskGB,
		ADBPort:     adbPort,
		ADBSerial:   adbSerial,
		WebRTCPort:  webrtcPort,
		CreatedAt:   time.Now(),
		RuntimeDir:  c.getRuntimeDir(instanceNum),
		Metadata:    opts.Metadata,
	}

	// Save instance info to disk
	infoPath := filepath.Join(instanceDir, "instance.json")
	infoData, _ := json.MarshalIndent(info, "", "  ")
	if err := c.writeFile(ctx, infoPath, infoData); err != nil {
		c.cleanupInstance(ctx, opts.SandboxId, instanceNum)
		return nil, fmt.Errorf("failed to save instance info: %w", err)
	}

	// Register instance in memory
	c.mutex.Lock()
	c.instances[opts.SandboxId] = info
	c.instanceNums[instanceNum] = opts.SandboxId
	c.mutex.Unlock()

	// Save mappings to disk
	if err := c.saveInstanceMappings(); err != nil {
		log.Warnf("Failed to save instance mappings: %v", err)
	}

	// Launch the Cuttlefish instance
	if err := c.launchInstance(ctx, info, opts.Snapshot); err != nil {
		c.cleanupInstance(ctx, opts.SandboxId, instanceNum)
		return nil, fmt.Errorf("failed to launch instance: %w", err)
	}

	// Wait for ADB to be ready
	if err := c.waitForADB(ctx, instanceNum, 120*time.Second); err != nil {
		log.Warnf("ADB not ready after launch: %v (continuing anyway)", err)
	}

	// Update state
	c.mutex.Lock()
	info.State = InstanceStateRunning
	c.mutex.Unlock()

	log.Infof("Sandbox %s created successfully with instance %d (ADB: %s)", opts.SandboxId, instanceNum, adbSerial)

	return &SandboxInfo{
		Id:        info.SandboxId,
		State:     InstanceStateRunning,
		Vcpus:     info.Cpus,
		MemoryMB:  info.MemoryMB,
		ADBSerial: info.ADBSerial,
		ADBPort:   info.ADBPort,
		CreatedAt: info.CreatedAt,
		Metadata:  info.Metadata,
	}, nil
}

// launchInstance launches a Cuttlefish instance using cvd create
func (c *Client) launchInstance(ctx context.Context, info *InstanceInfo, snapshot string) error {
	log.Infof("Launching Cuttlefish instance %d for sandbox %s (snapshot: %s)", info.InstanceNum, info.SandboxId, snapshot)

	// Build cvd create command
	// Modern Cuttlefish uses 'cvd create' instead of launch_cvd
	// Key options:
	// --host_path: path to host tools (cvd_internal_start, etc.)
	// --product_path: path to Android system images
	// --instance_nums: specify instance number
	// --cpus: number of vCPUs
	// --memory_mb: memory in MB

	// Host tools path - cvd appends /bin/ to this path to find cvd_internal_start
	// Default to system-installed cuttlefish, but prefer bundled tools if available
	hostToolsPath := "/usr/lib/cuttlefish-common"

	var launchCmd string

	if snapshot != "" {
		snapshotDir := c.GetSnapshotDir(snapshot)

		if IsCustomSnapshot(snapshot) {
			// Custom snapshot (orgId/name): restore from snapshot
			// Custom snapshots include saved state, so we restore from snapshot_path
			snapshotPath := filepath.Join(snapshotDir, "snapshot")
			log.Infof("Restoring from custom snapshot: %s", snapshotPath)

			// For custom snapshots, use cvd start with snapshot restore
			launchCmd = fmt.Sprintf(
				"cd %s && HOME=%s %s create "+
					"--host_path=%s "+
					"--instance_nums=%d "+
					"--cpus=%d "+
					"--memory_mb=%d "+
					"--report_anonymous_usage_stats=n "+
					"--start_webrtc=true "+
					"--snapshot_path=%s 2>&1",
				snapshotDir,
				c.config.CVDHome,
				c.config.CVDPath,
				hostToolsPath,
				info.InstanceNum,
				info.Cpus,
				info.MemoryMB,
				snapshotPath,
			)
		} else {
			// Base snapshot: use --product_path to specify system images
			log.Infof("Using base system image from: %s", snapshotDir)

			// Check if snapshot has bundled host tools (preferred for version compatibility)
			snapshotHostPath := hostToolsPath
			bundledHostTools := filepath.Join(snapshotDir, "bin", "assemble_cvd")
			if exists, _ := c.fileExists(ctx, bundledHostTools); exists {
				snapshotHostPath = snapshotDir
				log.Infof("Using bundled host tools from snapshot: %s", snapshotDir)
			}

			launchCmd = fmt.Sprintf(
				"cd %s && HOME=%s %s create "+
					"--host_path=%s "+
					"--product_path=%s "+
					"--instance_nums=%d "+
					"--cpus=%d "+
					"--memory_mb=%d "+
					"--report_anonymous_usage_stats=n "+
					"--start_webrtc=true "+
					"--data_policy=create_if_missing 2>&1",
				snapshotDir,
				c.config.CVDHome,
				c.config.CVDPath,
				snapshotHostPath,
				snapshotDir,
				info.InstanceNum,
				info.Cpus,
				info.MemoryMB,
			)
		}
	} else {
		// No snapshot specified - cvd create needs product_path
		// This will fail if no images are available
		launchCmd = fmt.Sprintf(
			"cd %s && HOME=%s %s create "+
				"--host_path=%s "+
				"--instance_nums=%d "+
				"--cpus=%d "+
				"--memory_mb=%d "+
				"--report_anonymous_usage_stats=n "+
				"--start_webrtc=true "+
				"--data_policy=create_if_missing 2>&1",
			c.config.CVDHome,
			c.config.CVDHome,
			c.config.CVDPath,
			hostToolsPath,
			info.InstanceNum,
			info.Cpus,
			info.MemoryMB,
		)
	}

	log.Debugf("Running launch command: %s", launchCmd)

	output, err := c.runShellScript(ctx, launchCmd)
	if err != nil {
		return fmt.Errorf("cvd create failed: %w (output: %s)", err, output)
	}

	log.Infof("cvd create completed for instance %d", info.InstanceNum)
	return nil
}

// cleanupInstance removes instance resources on failure
func (c *Client) cleanupInstance(ctx context.Context, sandboxId string, instanceNum int) {
	log.Warnf("Cleaning up instance %d for sandbox %s after failure", instanceNum, sandboxId)

	// Stop the instance if running
	_ = c.stopInstance(ctx, instanceNum)

	// Force remove any stale CVD state for this instance
	// This handles cases where cvd create fails midway and leaves stale device registrations
	c.forceCleanupInstance(ctx, instanceNum)

	// Remove from tracking
	c.mutex.Lock()
	delete(c.instances, sandboxId)
	delete(c.instanceNums, instanceNum)
	c.mutex.Unlock()

	// Save updated mappings
	_ = c.saveInstanceMappings()

	// Remove instance directory
	instanceDir := c.getInstanceDir(sandboxId)
	_ = c.runCommand(ctx, "rm", "-rf", instanceDir)
}

// forceCleanupInstance forcefully cleans up CVD state for a specific instance
// This is necessary when cvd create fails midway and leaves stale state
func (c *Client) forceCleanupInstance(ctx context.Context, instanceNum int) {
	log.Infof("Force cleaning up CVD state for instance %d", instanceNum)

	// Kill any processes associated with this instance number
	killCmd := fmt.Sprintf(
		"pkill -9 -f 'instance_nums.*%d|CUTTLEFISH_INSTANCE=%d|cvd-%d' 2>/dev/null || true",
		instanceNum, instanceNum, instanceNum,
	)
	_, _ = c.runShellScript(ctx, killCmd)

	// Clean up temp directories for this instance
	cleanupCmd := fmt.Sprintf(
		"rm -rf /tmp/cf_avd_*/%d /tmp/cf_env_*/%d 2>/dev/null || true; "+
			"find /var/tmp/cvd -type d -name '*cvd-%d*' -exec rm -rf {} + 2>/dev/null || true",
		instanceNum, instanceNum, instanceNum,
	)
	_, _ = c.runShellScript(ctx, cleanupCmd)

	// Try to remove from CVD fleet (in case it was partially registered)
	// The group name follows the pattern cvd_N where N is instance_num
	groupName := fmt.Sprintf("cvd_%d", instanceNum)
	rmCmd := fmt.Sprintf(
		"HOME=%s %s -group_name %s stop 2>/dev/null || true; HOME=%s %s rm -group_name %s 2>/dev/null || true",
		c.config.CVDHome, c.config.CVDPath, groupName,
		c.config.CVDHome, c.config.CVDPath, groupName,
	)
	_, _ = c.runShellScript(ctx, rmCmd)

	// Force-clean CVD instance database to remove stale entries
	// This prevents "New instance conflicts with existing instance" errors on next create
	cleanDBCmd := fmt.Sprintf(
		"rm -f /var/tmp/cvd/*/instance_database.binpb 2>/dev/null || true",
	)
	_, _ = c.runShellScript(ctx, cleanDBCmd)

	// Clean stale runtime dirs that have no running processes
	cleanStaleCmd := fmt.Sprintf(
		`for dir in /var/tmp/cvd/1001/*/; do
			if [ -d "$dir" ] && ! pgrep -f "$dir" > /dev/null 2>&1; then
				rm -rf "$dir"
			fi
		done 2>/dev/null || true`,
	)
	_, _ = c.runShellScript(ctx, cleanStaleCmd)

	// Also try to clean up stale operator registrations
	if err := c.EnsureOperatorDeviceClean(ctx, instanceNum); err != nil {
		log.Debugf("Could not clean operator registration: %v", err)
	}
}

// stopInstance stops a Cuttlefish instance
func (c *Client) stopInstance(ctx context.Context, instanceNum int) error {
	log.Infof("Stopping Cuttlefish instance %d", instanceNum)

	// Modern CVD uses group_name to select which group to stop
	// Group names follow the pattern cvd_N where N is the instance number
	groupName := fmt.Sprintf("cvd_%d", instanceNum)

	// Use 'cvd -group_name <name> stop' to stop a specific group
	stopCmd := fmt.Sprintf(
		"HOME=%s %s -group_name %s stop 2>&1",
		c.config.CVDHome,
		c.config.CVDPath,
		groupName,
	)

	output, err := c.runShellScript(ctx, stopCmd)
	if err != nil {
		log.Warnf("cvd stop returned error (may be OK if instance wasn't running): %v (output: %s)", err, output)
	} else {
		log.Infof("CVD group %s stopped successfully", groupName)
	}

	// Verify the instance actually stopped
	state := c.getInstanceState(ctx, instanceNum)
	if state == InstanceStateRunning {
		log.Warnf("Instance %d still running after stop command, attempting force kill", instanceNum)
		// Force kill as fallback
		killCmd := fmt.Sprintf("pkill -9 -f 'instance_nums.*%d|cvd-%d' 2>/dev/null || true", instanceNum, instanceNum)
		_, _ = c.runShellScript(ctx, killCmd)
	}

	return nil
}
