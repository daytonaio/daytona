// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package cuttlefish

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// Client is a Cuttlefish management client
type Client struct {
	config       ClientConfig
	instances    map[string]*InstanceInfo // sandboxId -> instance info
	instanceNums map[int]string           // instanceNum -> sandboxId (for reverse lookup)
	mutex        sync.RWMutex
	sandboxMutex map[string]*sync.Mutex
	sandboxMuMu  sync.Mutex

	// Exported for proxy access
	SSHHost    string
	SSHKeyPath string
}

// NewClient creates a new Cuttlefish client
func NewClient(config ClientConfig) (*Client, error) {
	// Set defaults
	if config.InstancesPath == "" {
		config.InstancesPath = "/var/lib/cuttlefish/instances"
	}
	if config.ArtifactsPath == "" {
		config.ArtifactsPath = "/var/lib/cuttlefish/artifacts"
	}
	if config.CVDHome == "" {
		config.CVDHome = "/home/vsoc-01"
	}
	if config.LaunchCVDPath == "" {
		config.LaunchCVDPath = "/home/vsoc-01/bin/launch_cvd"
	}
	if config.StopCVDPath == "" {
		config.StopCVDPath = "/home/vsoc-01/bin/stop_cvd"
	}
	if config.CVDPath == "" {
		config.CVDPath = "/home/vsoc-01/bin/cvd"
	}
	if config.ADBPath == "" {
		config.ADBPath = "adb"
	}
	if config.DefaultCpus == 0 {
		config.DefaultCpus = 2
	}
	if config.DefaultMemoryMB == 0 {
		config.DefaultMemoryMB = 4096
	}
	if config.DefaultDiskGB == 0 {
		config.DefaultDiskGB = 20
	}
	if config.BaseInstanceNum == 0 {
		config.BaseInstanceNum = 1
	}
	if config.MaxInstances == 0 {
		config.MaxInstances = 100
	}
	if config.ADBBasePort == 0 {
		config.ADBBasePort = 6520
	}
	if config.WebRTCBasePort == 0 {
		config.WebRTCBasePort = 8443
	}

	c := &Client{
		config:       config,
		instances:    make(map[string]*InstanceInfo),
		instanceNums: make(map[int]string),
		sandboxMutex: make(map[string]*sync.Mutex),
		SSHHost:      config.SSHHost,
		SSHKeyPath:   config.SSHKeyPath,
	}

	// Ensure directories exist
	if err := c.ensureDirectories(); err != nil {
		return nil, fmt.Errorf("failed to ensure directories: %w", err)
	}

	// Load existing instance mappings
	if err := c.loadInstanceMappings(); err != nil {
		log.Warnf("Failed to load instance mappings: %v", err)
	}

	return c, nil
}

// ensureDirectories creates required directories
func (c *Client) ensureDirectories() error {
	dirs := []string{
		c.config.InstancesPath,
	}

	for _, dir := range dirs {
		if err := c.runCommand(context.Background(), "mkdir", "-p", dir); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// loadInstanceMappings loads existing instance mappings from disk
func (c *Client) loadInstanceMappings() error {
	mappingsFile := filepath.Join(c.config.InstancesPath, "mappings.json")

	data, err := c.readFile(context.Background(), mappingsFile)
	if err != nil {
		// File doesn't exist yet, that's OK
		return nil
	}

	var mappings []InstanceMapping
	if err := json.Unmarshal(data, &mappings); err != nil {
		return fmt.Errorf("failed to unmarshal mappings: %w", err)
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	for _, m := range mappings {
		c.instances[m.SandboxId] = &InstanceInfo{
			SandboxId:   m.SandboxId,
			InstanceNum: m.InstanceNum,
			CreatedAt:   m.CreatedAt,
			ADBPort:     c.config.ADBBasePort + (m.InstanceNum - 1),
			ADBSerial:   fmt.Sprintf("0.0.0.0:%d", c.config.ADBBasePort+(m.InstanceNum-1)),
		}
		c.instanceNums[m.InstanceNum] = m.SandboxId
	}

	log.Infof("Loaded %d instance mappings", len(mappings))
	return nil
}

// saveInstanceMappings persists instance mappings to disk
func (c *Client) saveInstanceMappings() error {
	c.mutex.RLock()
	mappings := make([]InstanceMapping, 0, len(c.instances))
	for _, info := range c.instances {
		mappings = append(mappings, InstanceMapping{
			SandboxId:   info.SandboxId,
			InstanceNum: info.InstanceNum,
			CreatedAt:   info.CreatedAt,
		})
	}
	c.mutex.RUnlock()

	data, err := json.MarshalIndent(mappings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal mappings: %w", err)
	}

	mappingsFile := filepath.Join(c.config.InstancesPath, "mappings.json")
	return c.writeFile(context.Background(), mappingsFile, data)
}

// allocateInstanceNum finds an available instance number
func (c *Client) allocateInstanceNum() (int, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for i := c.config.BaseInstanceNum; i < c.config.BaseInstanceNum+c.config.MaxInstances; i++ {
		if _, exists := c.instanceNums[i]; !exists {
			return i, nil
		}
	}

	return 0, fmt.Errorf("no available instance numbers (max: %d)", c.config.MaxInstances)
}

// getSandboxMutex returns a mutex for a specific sandbox
func (c *Client) getSandboxMutex(sandboxId string) *sync.Mutex {
	c.sandboxMuMu.Lock()
	defer c.sandboxMuMu.Unlock()

	if _, ok := c.sandboxMutex[sandboxId]; !ok {
		c.sandboxMutex[sandboxId] = &sync.Mutex{}
	}
	return c.sandboxMutex[sandboxId]
}

// getInstanceDir returns the data directory for a sandbox
func (c *Client) getInstanceDir(sandboxId string) string {
	return filepath.Join(c.config.InstancesPath, sandboxId)
}

// getRuntimeDir returns the Cuttlefish runtime directory for an instance
func (c *Client) getRuntimeDir(instanceNum int) string {
	return filepath.Join(c.config.CVDHome, "cuttlefish", "instances", fmt.Sprintf("cvd-%d", instanceNum))
}

// IsRemote returns true if the client is configured for remote SSH operation
func (c *Client) IsRemote() bool {
	return c.config.SSHHost != ""
}

// runCommand executes a command locally or remotely via SSH
func (c *Client) runCommand(ctx context.Context, name string, args ...string) error {
	output, err := c.runCommandOutput(ctx, name, args...)
	if err != nil {
		return fmt.Errorf("%w: %s", err, output)
	}
	return nil
}

// runCommandOutput executes a command and returns stdout
func (c *Client) runCommandOutput(ctx context.Context, name string, args ...string) (string, error) {
	var cmd *exec.Cmd

	if c.IsRemote() {
		// Build command string for SSH with proper quoting
		cmdStr := name
		for _, arg := range args {
			if needsQuoting(arg) {
				cmdStr += " '" + arg + "'"
			} else {
				cmdStr += " " + arg
			}
		}
		cmd = exec.CommandContext(ctx, "ssh",
			"-i", c.config.SSHKeyPath,
			"-o", "StrictHostKeyChecking=accept-new",
			"-o", "BatchMode=yes",
			c.config.SSHHost,
			cmdStr,
		)
	} else {
		cmd = exec.CommandContext(ctx, name, args...)
	}

	output, err := cmd.CombinedOutput()
	return string(output), err
}

// runShellScript executes a shell script either locally or remotely via SSH
func (c *Client) runShellScript(ctx context.Context, script string) (string, error) {
	if c.IsRemote() {
		return c.runSSHCommand(ctx, script)
	}

	// Local mode: execute directly via /bin/sh
	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", script)
	log.Debugf("Running local shell script: %s", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Warnf("Local shell script failed: err=%v, output=%s", err, string(output))
		return string(output), fmt.Errorf("shell script failed: %w (output: %s)", err, string(output))
	}
	return string(output), nil
}

// runSSHCommand executes a command on the remote host
func (c *Client) runSSHCommand(ctx context.Context, command string) (string, error) {
	if !c.IsRemote() {
		return "", fmt.Errorf("runSSHCommand is only available in remote mode")
	}

	// Check if SSH key file exists
	if _, err := os.Stat(c.config.SSHKeyPath); os.IsNotExist(err) {
		return "", fmt.Errorf("SSH key file not found: %s", c.config.SSHKeyPath)
	}

	cmd := exec.CommandContext(ctx, "ssh",
		"-i", c.config.SSHKeyPath,
		"-o", "StrictHostKeyChecking=accept-new",
		"-o", "ConnectTimeout=10",
		"-o", "BatchMode=yes",
		c.config.SSHHost,
		command,
	)

	log.Debugf("Running SSH command: ssh -i %s %s '%s'", c.config.SSHKeyPath, c.config.SSHHost, command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Warnf("SSH command failed: cmd=%s, err=%v, output=%s", command, err, string(output))
		return string(output), fmt.Errorf("ssh command failed: %w (output: %s)", err, string(output))
	}
	return string(output), nil
}

// fileExists checks if a file exists (locally or remotely)
func (c *Client) fileExists(ctx context.Context, path string) (bool, error) {
	if c.IsRemote() {
		cmd := exec.CommandContext(ctx, "ssh",
			"-i", c.config.SSHKeyPath,
			"-o", "StrictHostKeyChecking=accept-new",
			c.config.SSHHost,
			fmt.Sprintf("test -e %s && echo exists", path),
		)
		output, err := cmd.Output()
		if err != nil {
			return false, nil
		}
		return strings.TrimSpace(string(output)) == "exists", nil
	}

	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// writeFile writes content to a file (locally or remotely)
func (c *Client) writeFile(ctx context.Context, path string, content []byte) error {
	if c.IsRemote() {
		cmd := exec.CommandContext(ctx, "ssh",
			"-i", c.config.SSHKeyPath,
			"-o", "StrictHostKeyChecking=accept-new",
			c.config.SSHHost,
			fmt.Sprintf("cat > %s", path),
		)
		cmd.Stdin = strings.NewReader(string(content))
		return cmd.Run()
	}

	return os.WriteFile(path, content, 0644)
}

// readFile reads content from a file (locally or remotely)
func (c *Client) readFile(ctx context.Context, path string) ([]byte, error) {
	if c.IsRemote() {
		cmd := exec.CommandContext(ctx, "ssh",
			"-i", c.config.SSHKeyPath,
			"-o", "StrictHostKeyChecking=accept-new",
			c.config.SSHHost,
			fmt.Sprintf("cat %s", path),
		)
		return cmd.Output()
	}

	return os.ReadFile(path)
}

// GetInstance returns information about a specific instance
func (c *Client) GetInstance(sandboxId string) (*InstanceInfo, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	info, exists := c.instances[sandboxId]
	return info, exists
}

// GetInstanceByNum returns the sandbox ID for an instance number
func (c *Client) GetInstanceByNum(instanceNum int) (string, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	sandboxId, exists := c.instanceNums[instanceNum]
	return sandboxId, exists
}

// List returns all sandbox IDs
func (c *Client) List(ctx context.Context) ([]string, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	ids := make([]string, 0, len(c.instances))
	for id := range c.instances {
		ids = append(ids, id)
	}
	return ids, nil
}

// ListWithInfo returns all sandboxes with their info
func (c *Client) ListWithInfo(ctx context.Context) ([]*SandboxInfo, error) {
	c.mutex.RLock()
	instances := make([]*InstanceInfo, 0, len(c.instances))
	for _, info := range c.instances {
		instances = append(instances, info)
	}
	c.mutex.RUnlock()

	sandboxes := make([]*SandboxInfo, 0, len(instances))
	for _, info := range instances {
		// Get current state
		state := c.getInstanceState(ctx, info.InstanceNum)

		sandboxes = append(sandboxes, &SandboxInfo{
			Id:        info.SandboxId,
			State:     state,
			Vcpus:     info.Cpus,
			MemoryMB:  info.MemoryMB,
			ADBSerial: info.ADBSerial,
			ADBPort:   info.ADBPort,
			CreatedAt: info.CreatedAt,
			Metadata:  info.Metadata,
		})
	}

	return sandboxes, nil
}

// getInstanceState checks the current state of an instance
func (c *Client) getInstanceState(ctx context.Context, instanceNum int) InstanceState {
	// Check if the instance process is running
	checkCmd := fmt.Sprintf("pgrep -f 'cuttlefish.*instance_nums.*%d' > /dev/null 2>&1 && echo running || echo stopped", instanceNum)
	output, err := c.runShellScript(ctx, checkCmd)
	if err != nil {
		return InstanceStateUnknown
	}

	if strings.TrimSpace(output) == "running" {
		return InstanceStateRunning
	}
	return InstanceStateStopped
}

// GetSandboxInfo returns information about a sandbox
func (c *Client) GetSandboxInfo(ctx context.Context, sandboxId string) (*SandboxInfo, error) {
	info, exists := c.GetInstance(sandboxId)
	if !exists {
		return nil, fmt.Errorf("sandbox %s not found", sandboxId)
	}

	state := c.getInstanceState(ctx, info.InstanceNum)

	return &SandboxInfo{
		Id:        info.SandboxId,
		State:     state,
		Vcpus:     info.Cpus,
		MemoryMB:  info.MemoryMB,
		ADBSerial: info.ADBSerial,
		ADBPort:   info.ADBPort,
		CreatedAt: info.CreatedAt,
		Metadata:  info.Metadata,
	}, nil
}

// GetRemoteMetrics collects system metrics from the remote host via SSH
func (c *Client) GetRemoteMetrics(ctx context.Context) (*RemoteMetrics, error) {
	if !c.IsRemote() {
		return nil, fmt.Errorf("GetRemoteMetrics is only available in remote mode")
	}

	metrics := &RemoteMetrics{}
	var wg sync.WaitGroup
	var mu sync.Mutex
	errChan := make(chan error, 4)

	// Get CPU count
	wg.Add(1)
	go func() {
		defer wg.Done()
		output, err := c.runSSHCommand(ctx, "nproc")
		if err != nil {
			errChan <- fmt.Errorf("cpu count: %w", err)
			return
		}
		count, _ := strconv.Atoi(strings.TrimSpace(output))
		mu.Lock()
		metrics.TotalCPUs = count
		mu.Unlock()
	}()

	// Get memory info
	wg.Add(1)
	go func() {
		defer wg.Done()
		output, err := c.runSSHCommand(ctx, "free -b | grep Mem")
		if err != nil {
			errChan <- fmt.Errorf("memory: %w", err)
			return
		}
		fields := strings.Fields(output)
		if len(fields) >= 7 {
			var total, available float64
			fmt.Sscanf(fields[1], "%f", &total)
			fmt.Sscanf(fields[6], "%f", &available)
			mu.Lock()
			if total > 0 {
				metrics.MemoryUsagePercent = ((total - available) / total) * 100
				metrics.TotalMemoryGiB = total / (1024 * 1024 * 1024)
			}
			mu.Unlock()
		}
	}()

	// Get disk info
	wg.Add(1)
	go func() {
		defer wg.Done()
		output, err := c.runSSHCommand(ctx, "df -B1 / | tail -1")
		if err != nil {
			errChan <- fmt.Errorf("disk: %w", err)
			return
		}
		fields := strings.Fields(output)
		if len(fields) >= 5 {
			var total, used float64
			fmt.Sscanf(fields[1], "%f", &total)
			fmt.Sscanf(fields[2], "%f", &used)
			mu.Lock()
			if total > 0 {
				metrics.DiskUsagePercent = (used / total) * 100
				metrics.TotalDiskGiB = total / (1024 * 1024 * 1024)
			}
			mu.Unlock()
		}
	}()

	wg.Wait()
	close(errChan)

	// Collect errors
	var errs []string
	for err := range errChan {
		errs = append(errs, err.Error())
	}
	if len(errs) > 0 {
		log.Warnf("Some remote metrics failed: %v", errs)
	}

	return metrics, nil
}

// Close cleans up client resources
func (c *Client) Close() error {
	return c.saveInstanceMappings()
}

// needsQuoting checks if a string needs to be quoted for shell
func needsQuoting(s string) bool {
	for _, char := range s {
		if char == ' ' || char == '\t' || char == '\n' || char == '$' || char == '`' ||
			char == '"' || char == '\\' || char == '!' || char == '*' || char == '?' ||
			char == '[' || char == ']' || char == '{' || char == '}' || char == '(' || char == ')' {
			return true
		}
	}
	return false
}

// waitForADB waits for ADB to be ready for an instance
func (c *Client) waitForADB(ctx context.Context, instanceNum int, timeout time.Duration) error {
	adbSerial := fmt.Sprintf("0.0.0.0:%d", c.config.ADBBasePort+(instanceNum-1))
	log.Infof("Waiting for ADB to be ready on %s", adbSerial)

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Try to connect via ADB
		checkCmd := fmt.Sprintf("%s -s %s shell echo ready 2>/dev/null", c.config.ADBPath, adbSerial)
		output, err := c.runShellScript(ctx, checkCmd)
		if err == nil && strings.Contains(output, "ready") {
			log.Infof("ADB ready on %s", adbSerial)
			return nil
		}

		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("timeout waiting for ADB on %s", adbSerial)
}
