// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package cloudhypervisor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// ClientConfig holds configuration for the Cloud Hypervisor client
type ClientConfig struct {
	// SandboxesPath is the base directory for VM working directories
	SandboxesPath string
	// SnapshotsPath is the directory for snapshot storage
	SnapshotsPath string
	// SocketsPath is the directory for API sockets
	SocketsPath string
	// KernelPath is the path to the vmlinux kernel
	KernelPath string
	// FirmwarePath is the path to hypervisor-fw
	FirmwarePath string
	// BaseImagePath is the path to the base disk image
	BaseImagePath string
	// DefaultCpus is the default number of vCPUs
	DefaultCpus int
	// DefaultMemoryMB is the default memory in MB
	DefaultMemoryMB uint64
	// SSHHost is the remote host for SSH-based operations (empty for local mode)
	SSHHost string
	// SSHKeyPath is the path to the SSH private key
	SSHKeyPath string
	// BridgeName is the network bridge name (e.g., br0)
	BridgeName string
	// TapCreateScript is the path to the TAP creation script
	TapCreateScript string
	// TapDeleteScript is the path to the TAP deletion script
	TapDeleteScript string
}

// Client is a Cloud Hypervisor API client
type Client struct {
	config       ClientConfig
	httpClients  map[string]*http.Client // Socket path -> HTTP client
	httpMutex    sync.RWMutex
	sandboxMutex map[string]*sync.Mutex
	sandboxMuMu  sync.Mutex
}

// NewClient creates a new Cloud Hypervisor client
func NewClient(config ClientConfig) (*Client, error) {
	// Set defaults
	if config.SandboxesPath == "" {
		config.SandboxesPath = "/var/lib/cloud-hypervisor/sandboxes"
	}
	if config.SnapshotsPath == "" {
		config.SnapshotsPath = "/var/lib/cloud-hypervisor/snapshots"
	}
	if config.SocketsPath == "" {
		config.SocketsPath = "/var/run/cloud-hypervisor"
	}
	if config.KernelPath == "" {
		config.KernelPath = "/var/lib/cloud-hypervisor/kernels/vmlinux"
	}
	if config.FirmwarePath == "" {
		config.FirmwarePath = "/var/lib/cloud-hypervisor/firmware/hypervisor-fw"
	}
	if config.BaseImagePath == "" {
		config.BaseImagePath = "/var/lib/cloud-hypervisor/images/ubuntu-24.04-server-cloudimg-amd64.img"
	}
	if config.DefaultCpus == 0 {
		config.DefaultCpus = 2
	}
	if config.DefaultMemoryMB == 0 {
		config.DefaultMemoryMB = 2048
	}
	if config.BridgeName == "" {
		config.BridgeName = "br0"
	}
	if config.TapCreateScript == "" {
		config.TapCreateScript = "/usr/local/bin/ch-create-tap"
	}
	if config.TapDeleteScript == "" {
		config.TapDeleteScript = "/usr/local/bin/ch-delete-tap"
	}

	c := &Client{
		config:       config,
		httpClients:  make(map[string]*http.Client),
		sandboxMutex: make(map[string]*sync.Mutex),
	}

	// Ensure directories exist
	if err := c.ensureDirectories(); err != nil {
		return nil, fmt.Errorf("failed to ensure directories: %w", err)
	}

	return c, nil
}

// ensureDirectories creates required directories
func (c *Client) ensureDirectories() error {
	dirs := []string{
		c.config.SandboxesPath,
		c.config.SnapshotsPath,
		c.config.SocketsPath,
	}

	for _, dir := range dirs {
		if err := c.runCommand(context.Background(), "mkdir", "-p", dir); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
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

// getSocketPath returns the API socket path for a sandbox
func (c *Client) getSocketPath(sandboxId string) string {
	return filepath.Join(c.config.SocketsPath, fmt.Sprintf("%s.sock", sandboxId))
}

// getSandboxDir returns the working directory for a sandbox
func (c *Client) getSandboxDir(sandboxId string) string {
	return filepath.Join(c.config.SandboxesPath, sandboxId)
}

// getDiskPath returns the disk image path for a sandbox
func (c *Client) getDiskPath(sandboxId string) string {
	return filepath.Join(c.getSandboxDir(sandboxId), "disk.raw")
}

// getConfigPath returns the config JSON path for a sandbox
func (c *Client) getConfigPath(sandboxId string) string {
	return filepath.Join(c.getSandboxDir(sandboxId), "config.json")
}

// getSnapshotPath returns the snapshot path for a sandbox
func (c *Client) getSnapshotPath(sandboxId string) string {
	return filepath.Join(c.config.SnapshotsPath, sandboxId)
}

// getTapName returns the TAP device name for a sandbox
func (c *Client) getTapName(sandboxId string) string {
	// Use first 12 chars of sandbox ID for tap name (Linux limit is 15 chars)
	name := sandboxId
	if len(name) > 12 {
		name = name[:12]
	}
	return fmt.Sprintf("tap-%s", name)
}

// IsRemote returns true if the client is configured for remote SSH operation
func (c *Client) IsRemote() bool {
	return c.config.SSHHost != ""
}

// getHTTPClient returns an HTTP client for the given socket path
func (c *Client) getHTTPClient(socketPath string) *http.Client {
	c.httpMutex.RLock()
	client, exists := c.httpClients[socketPath]
	c.httpMutex.RUnlock()

	if exists {
		return client
	}

	c.httpMutex.Lock()
	defer c.httpMutex.Unlock()

	// Double-check after acquiring write lock
	if client, exists = c.httpClients[socketPath]; exists {
		return client
	}

	// Create HTTP client that connects via Unix socket
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			if c.IsRemote() {
				// For remote connections, use SSH tunnel to access the socket
				return c.dialRemoteSocket(ctx, socketPath)
			}
			// Local connection via Unix socket
			return net.DialTimeout("unix", socketPath, 10*time.Second)
		},
	}

	client = &http.Client{
		Transport: transport,
		Timeout:   60 * time.Second,
	}

	c.httpClients[socketPath] = client
	return client
}

// dialRemoteSocket creates a connection to a remote Unix socket via SSH
func (c *Client) dialRemoteSocket(ctx context.Context, socketPath string) (net.Conn, error) {
	// Use SSH to create a tunnel to the remote socket
	// This creates a local TCP connection that forwards to the remote Unix socket

	// For simplicity, we use socat on the remote host
	// In production, you might want to use a proper SSH library
	cmd := exec.CommandContext(ctx, "ssh",
		"-i", c.config.SSHKeyPath,
		"-o", "StrictHostKeyChecking=accept-new",
		c.config.SSHHost,
		fmt.Sprintf("socat - UNIX-CONNECT:%s", socketPath),
	)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start SSH tunnel: %w", err)
	}

	return &sshConn{
		cmd:    cmd,
		stdin:  stdin,
		stdout: stdout,
	}, nil
}

// sshConn wraps an SSH tunnel as a net.Conn
type sshConn struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
}

func (s *sshConn) Read(b []byte) (int, error) {
	return s.stdout.Read(b)
}

func (s *sshConn) Write(b []byte) (int, error) {
	return s.stdin.Write(b)
}

func (s *sshConn) Close() error {
	s.stdin.Close()
	s.stdout.Close()
	return s.cmd.Wait()
}

func (s *sshConn) LocalAddr() net.Addr {
	return &net.UnixAddr{Name: "ssh-tunnel", Net: "unix"}
}

func (s *sshConn) RemoteAddr() net.Addr {
	return &net.UnixAddr{Name: "ssh-tunnel-remote", Net: "unix"}
}

func (s *sshConn) SetDeadline(t time.Time) error {
	return nil
}

func (s *sshConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (s *sshConn) SetWriteDeadline(t time.Time) error {
	return nil
}

// apiRequest makes an HTTP request to the Cloud Hypervisor API
// For remote hosts, it uses ch-remote via SSH instead of direct socket access
func (c *Client) apiRequest(ctx context.Context, sandboxId, method, endpoint string, body interface{}) ([]byte, error) {
	socketPath := c.getSocketPath(sandboxId)

	// For remote hosts, use ch-remote via SSH
	if c.IsRemote() {
		return c.apiRequestRemote(ctx, sandboxId, socketPath, method, endpoint, body)
	}

	// Local mode: direct HTTP to socket
	client := c.getHTTPClient(socketPath)

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
		log.Debugf("CH API request body: %s", string(jsonBody))
	}

	// Cloud Hypervisor API uses http://localhost as the base URL for Unix socket connections
	url := fmt.Sprintf("http://localhost/api/v1/%s", endpoint)

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// apiRequestRemote executes a CH API request via ch-remote over SSH
func (c *Client) apiRequestRemote(ctx context.Context, sandboxId, socketPath, method, endpoint string, body interface{}) ([]byte, error) {
	// Handle vm.create specially since it needs JSON config via stdin
	if endpoint == "vm.create" && body != nil {
		return c.createVmRemote(ctx, socketPath, body)
	}

	// Map HTTP method + endpoint to ch-remote command
	chRemoteCmd, err := c.mapEndpointToChRemote(method, endpoint, body)
	if err != nil {
		return nil, err
	}

	// Build the full ch-remote command
	cmdStr := fmt.Sprintf("ch-remote --api-socket %s %s", socketPath, chRemoteCmd)
	log.Debugf("Executing ch-remote: %s", cmdStr)

	output, err := c.runCommandOutput(ctx, "sh", "-c", cmdStr)
	if err != nil {
		return nil, fmt.Errorf("ch-remote failed: %w (output: %s)", err, output)
	}

	return []byte(output), nil
}

// createVmRemote creates a VM on a remote host by writing config to a temp file
func (c *Client) createVmRemote(ctx context.Context, socketPath string, body interface{}) ([]byte, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal VM config: %w", err)
	}

	// Write config to temp file on remote host, then use ch-remote to create
	tmpConfig := fmt.Sprintf("/tmp/ch-config-%d.json", time.Now().UnixNano())

	// Write the config file
	if err := c.writeFile(ctx, tmpConfig, jsonBody); err != nil {
		return nil, fmt.Errorf("failed to write config file: %w", err)
	}

	// Create VM using the config file
	cmdStr := fmt.Sprintf("ch-remote --api-socket %s create %s && rm -f %s", socketPath, tmpConfig, tmpConfig)
	log.Debugf("Executing ch-remote create: %s", cmdStr)

	output, err := c.runCommandOutput(ctx, "sh", "-c", cmdStr)
	if err != nil {
		// Clean up temp file on failure
		_ = c.runCommand(ctx, "rm", "-f", tmpConfig)
		return nil, fmt.Errorf("ch-remote create failed: %w (output: %s)", err, output)
	}

	return []byte(output), nil
}

// mapEndpointToChRemote maps CH API endpoints to ch-remote CLI commands
func (c *Client) mapEndpointToChRemote(method, endpoint string, body interface{}) (string, error) {
	switch endpoint {
	case "vmm.ping":
		return "ping", nil
	case "vm.info":
		return "info", nil
	case "vm.create":
		// Handled specially in apiRequestRemote
		return "create", nil
	case "vm.boot":
		return "boot", nil
	case "vm.pause":
		return "pause", nil
	case "vm.resume":
		return "resume", nil
	case "vm.shutdown":
		return "shutdown", nil
	case "vm.reboot":
		return "reboot", nil
	case "vm.power-button":
		return "power-button", nil
	case "vm.delete":
		return "delete", nil
	case "vmm.shutdown":
		return "shutdown-vmm", nil
	case "vm.snapshot":
		if body != nil {
			if m, ok := body.(map[string]string); ok {
				if destUrl, exists := m["destination_url"]; exists {
					return fmt.Sprintf("snapshot %s", destUrl), nil
				}
			}
		}
		return "", fmt.Errorf("snapshot requires destination_url")
	case "vm.restore":
		if cfg, ok := body.(RestoreConfig); ok {
			cmd := fmt.Sprintf("restore source_url=%s", cfg.SourceUrl)
			if cfg.Prefault {
				cmd += ",prefault=on"
			}
			return cmd, nil
		}
		return "", fmt.Errorf("restore requires RestoreConfig")
	case "vm.add-device":
		if cfg, ok := body.(string); ok {
			return fmt.Sprintf("add-device %s", cfg), nil
		}
		return "", fmt.Errorf("add-device requires config string")
	case "vm.remove-device":
		if m, ok := body.(map[string]string); ok {
			if id, exists := m["id"]; exists {
				return fmt.Sprintf("remove-device %s", id), nil
			}
		}
		return "", fmt.Errorf("remove-device requires device id")
	default:
		return "", fmt.Errorf("unsupported endpoint: %s", endpoint)
	}
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
			// Quote arguments that might contain special characters
			if needsQuoting(arg) {
				cmdStr += " '" + arg + "'"
			} else {
				cmdStr += " " + arg
			}
		}
		cmd = exec.CommandContext(ctx, "ssh",
			"-i", c.config.SSHKeyPath,
			"-o", "StrictHostKeyChecking=accept-new",
			c.config.SSHHost,
			cmdStr,
		)
	} else {
		cmd = exec.CommandContext(ctx, name, args...)
	}

	output, err := cmd.CombinedOutput()
	return string(output), err
}

// needsQuoting checks if a string needs to be quoted for shell
func needsQuoting(s string) bool {
	for _, c := range s {
		if c == ' ' || c == '\t' || c == '\n' || c == '$' || c == '`' ||
			c == '"' || c == '\\' || c == '!' || c == '*' || c == '?' ||
			c == '[' || c == ']' || c == '{' || c == '}' || c == '(' || c == ')' {
			return true
		}
	}
	return false
}

// fileExists checks if a file exists (locally or remotely)
func (c *Client) fileExists(ctx context.Context, path string) (bool, error) {
	var cmd *exec.Cmd

	if c.IsRemote() {
		cmd = exec.CommandContext(ctx, "ssh",
			"-i", c.config.SSHKeyPath,
			"-o", "StrictHostKeyChecking=accept-new",
			c.config.SSHHost,
			fmt.Sprintf("test -e %s && echo exists", path),
		)
	} else {
		_, err := os.Stat(path)
		if err == nil {
			return true, nil
		}
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	output, err := cmd.Output()
	if err != nil {
		return false, nil // File doesn't exist
	}
	return string(output) == "exists\n", nil
}

// writeFile writes content to a file (locally or remotely)
func (c *Client) writeFile(ctx context.Context, path string, content []byte) error {
	if c.IsRemote() {
		// Use SSH + cat to write file
		cmd := exec.CommandContext(ctx, "ssh",
			"-i", c.config.SSHKeyPath,
			"-o", "StrictHostKeyChecking=accept-new",
			c.config.SSHHost,
			fmt.Sprintf("cat > %s", path),
		)
		cmd.Stdin = bytes.NewReader(content)
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

// Ping checks if the API is accessible for a sandbox
func (c *Client) Ping(ctx context.Context, sandboxId string) error {
	socketPath := c.getSocketPath(sandboxId)

	// Check if socket exists
	exists, err := c.fileExists(ctx, socketPath)
	if err != nil {
		return fmt.Errorf("failed to check socket existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("VM socket does not exist: %s", socketPath)
	}

	// Try to get VM info
	_, err = c.apiRequest(ctx, sandboxId, http.MethodGet, "vmm.ping", nil)
	if err != nil {
		return fmt.Errorf("failed to ping VM: %w", err)
	}

	return nil
}

// GetInfo returns information about a VM
func (c *Client) GetInfo(ctx context.Context, sandboxId string) (*VmInfo, error) {
	resp, err := c.apiRequest(ctx, sandboxId, http.MethodGet, "vm.info", nil)
	if err != nil {
		return nil, err
	}

	var info VmInfo
	if err := json.Unmarshal(resp, &info); err != nil {
		return nil, fmt.Errorf("failed to unmarshal VM info: %w", err)
	}

	return &info, nil
}

// Close cleans up client resources
func (c *Client) Close() error {
	c.httpMutex.Lock()
	defer c.httpMutex.Unlock()

	for _, client := range c.httpClients {
		client.CloseIdleConnections()
	}
	c.httpClients = make(map[string]*http.Client)

	return nil
}

// GetRemoteMetrics collects system metrics from the remote host via SSH
func (c *Client) GetRemoteMetrics(ctx context.Context) (*RemoteMetrics, error) {
	if !c.IsRemote() {
		return nil, fmt.Errorf("GetRemoteMetrics is only available in remote mode")
	}

	log.Debugf("Getting remote metrics from %s (key: %s)", c.config.SSHHost, c.config.SSHKeyPath)

	metrics := &RemoteMetrics{}
	var wg sync.WaitGroup
	var mu sync.Mutex
	errChan := make(chan error, 4)

	// Get CPU count
	wg.Add(1)
	go func() {
		defer wg.Done()
		count, err := c.getRemoteCPUCount(ctx)
		if err != nil {
			errChan <- fmt.Errorf("cpu count: %w", err)
			return
		}
		mu.Lock()
		metrics.TotalCPUs = count
		mu.Unlock()
	}()

	// Get CPU usage
	wg.Add(1)
	go func() {
		defer wg.Done()
		usage, err := c.getRemoteCPUUsage(ctx)
		if err != nil {
			errChan <- fmt.Errorf("cpu usage: %w", err)
			return
		}
		mu.Lock()
		metrics.CPUUsagePercent = usage
		mu.Unlock()
	}()

	// Get memory info
	wg.Add(1)
	go func() {
		defer wg.Done()
		usage, total, err := c.getRemoteMemoryUsage(ctx)
		if err != nil {
			errChan <- fmt.Errorf("memory: %w", err)
			return
		}
		mu.Lock()
		metrics.MemoryUsagePercent = usage
		metrics.TotalMemoryGiB = total
		mu.Unlock()
	}()

	// Get disk info
	wg.Add(1)
	go func() {
		defer wg.Done()
		usage, total, err := c.getRemoteDiskUsage(ctx)
		if err != nil {
			errChan <- fmt.Errorf("disk: %w", err)
			return
		}
		mu.Lock()
		metrics.DiskUsagePercent = usage
		metrics.TotalDiskGiB = total
		mu.Unlock()
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

// getRemoteCPUCount gets the number of CPUs from the remote host
func (c *Client) getRemoteCPUCount(ctx context.Context) (int, error) {
	output, err := c.runSSHCommand(ctx, "nproc")
	if err != nil {
		return 0, err
	}
	var count int
	_, err = fmt.Sscanf(strings.TrimSpace(output), "%d", &count)
	return count, err
}

// getRemoteCPUUsage gets CPU usage percentage from the remote host
func (c *Client) getRemoteCPUUsage(ctx context.Context) (float64, error) {
	// Use vmstat to get CPU idle percentage
	output, err := c.runSSHCommand(ctx, "vmstat 1 2 | tail -1")
	if err != nil {
		// Fallback: use /proc/loadavg
		return c.getRemoteCPUUsageFromLoad(ctx)
	}

	fields := strings.Fields(strings.TrimSpace(output))
	if len(fields) < 15 {
		return c.getRemoteCPUUsageFromLoad(ctx)
	}

	var idle float64
	_, err = fmt.Sscanf(fields[14], "%f", &idle)
	if err != nil {
		return c.getRemoteCPUUsageFromLoad(ctx)
	}

	return 100.0 - idle, nil
}

// getRemoteCPUUsageFromLoad gets CPU usage from load average (fallback)
func (c *Client) getRemoteCPUUsageFromLoad(ctx context.Context) (float64, error) {
	output, err := c.runSSHCommand(ctx, "cat /proc/loadavg")
	if err != nil {
		return 0, err
	}

	fields := strings.Fields(strings.TrimSpace(output))
	if len(fields) < 1 {
		return 0, fmt.Errorf("invalid loadavg format")
	}

	var load1 float64
	_, err = fmt.Sscanf(fields[0], "%f", &load1)
	if err != nil {
		return 0, err
	}

	// Get CPU count for normalization
	cpuCount, err := c.getRemoteCPUCount(ctx)
	if err != nil || cpuCount == 0 {
		cpuCount = 1
	}

	// Normalize load to percentage (capped at 100)
	usage := (load1 / float64(cpuCount)) * 100
	if usage > 100 {
		usage = 100
	}
	return usage, nil
}

// getRemoteMemoryUsage gets memory usage from the remote host
func (c *Client) getRemoteMemoryUsage(ctx context.Context) (usagePercent float64, totalGiB float64, err error) {
	output, err := c.runSSHCommand(ctx, "free -b")
	if err != nil {
		return 0, 0, err
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Mem:") {
			fields := strings.Fields(line)
			if len(fields) < 7 {
				return 0, 0, fmt.Errorf("invalid free output format")
			}

			var total, available float64
			fmt.Sscanf(fields[1], "%f", &total)
			fmt.Sscanf(fields[6], "%f", &available)

			if total > 0 {
				usagePercent = ((total - available) / total) * 100
				totalGiB = total / (1024 * 1024 * 1024)
			}
			return usagePercent, totalGiB, nil
		}
	}
	return 0, 0, fmt.Errorf("Mem line not found in free output")
}

// getRemoteDiskUsage gets disk usage from the remote host
func (c *Client) getRemoteDiskUsage(ctx context.Context) (usagePercent float64, totalGiB float64, err error) {
	output, err := c.runSSHCommand(ctx, "df -B1 /")
	if err != nil {
		return 0, 0, err
	}

	lines := strings.Split(output, "\n")
	if len(lines) < 2 {
		return 0, 0, fmt.Errorf("invalid df output")
	}

	fields := strings.Fields(lines[1])
	if len(fields) < 5 {
		return 0, 0, fmt.Errorf("invalid df output format")
	}

	var total, used float64
	fmt.Sscanf(fields[1], "%f", &total)
	fmt.Sscanf(fields[2], "%f", &used)

	if total > 0 {
		usagePercent = (used / total) * 100
		totalGiB = total / (1024 * 1024 * 1024)
	}

	return usagePercent, totalGiB, nil
}

// runSSHCommand executes a simple command on the remote host
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
