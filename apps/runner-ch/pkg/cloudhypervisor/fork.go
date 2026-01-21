// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package cloudhypervisor

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unsafe"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

// TUN/TAP ioctl constants
const (
	tunSetIff = 0x400454ca // TUNSETIFF ioctl number
	iffTap    = 0x0002     // TAP device (layer 2)
	iffNoPi   = 0x1000     // No packet info
	iffMultiQ = 0x0100     // Multi-queue support
	devNetTun = "/dev/net/tun"
)

// ifreq structure for TUNSETIFF ioctl
type ifreq struct {
	name  [16]byte
	flags uint16
	_     [22]byte // padding to match kernel struct size
}

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

	// Step 3b: Patch the snapshot config to use the new disk path
	// The snapshot's config.json contains the source VM's disk paths which are locked
	// We need to replace them with our CoW overlay disk
	log.Infof("Patching snapshot config to use overlay disk")
	patchConfigCmd := fmt.Sprintf(`
# Read the config.json from snapshot
config_file="%s/config.json"
if [ -f "$config_file" ]; then
    # Use jq to replace the disk path
    # The first disk (_disk0) is the main OS disk, second (_disk1) is cloud-init
    jq '.disks[0].path = "%s"' "$config_file" > "$config_file.new" && mv "$config_file.new" "$config_file"
    echo "Patched disk path in snapshot config"
else
    echo "No config.json found in snapshot"
fi
`, snapshotPath, targetDiskPath)

	if output, err := c.runShellScript(ctx, patchConfigCmd); err != nil {
		log.Warnf("Failed to patch snapshot config: %v (output: %s)", err, output)
		// Continue anyway - the restore might work if disks aren't locked
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
	// In local mode, we can pass the TAP file descriptor for proper live restore
	// In remote mode, we fall back to cold restore (disk only, fresh boot)
	log.Infof("Restoring forked VM from snapshot")

	if !c.IsRemote() {
		// Get the network device ID from source VM config for proper FD mapping
		netId := "_net0" // Default fallback - should match create.go
		if sourceInfo != nil && sourceInfo.Config != nil && len(sourceInfo.Config.Net) > 0 {
			if sourceInfo.Config.Net[0].Id != "" {
				netId = sourceInfo.Config.Net[0].Id
				log.Infof("Using network device ID from source VM: %s", netId)
			}
		}

		// Local mode: Use FD passing for true live fork
		if err := c.restoreWithNetFds(ctx, opts.NewSandboxId, snapshotPath, netns, netId, opts.Prefault); err != nil {
			c.cleanupFork(ctx, opts.NewSandboxId, snapshotPath)
			return nil, fmt.Errorf("failed to restore forked VM with net_fds: %w", err)
		}
	} else {
		// Remote mode: Fall back to standard restore (cold fork - disk only)
		log.Warnf("Remote mode: live memory fork not supported, using cold fork (disk only)")
		restoreConfig := RestoreConfig{
			SourceUrl: fmt.Sprintf("file://%s", snapshotPath),
			Prefault:  opts.Prefault,
		}
		if _, err := c.apiRequest(ctx, opts.NewSandboxId, http.MethodPut, "vm.restore", restoreConfig); err != nil {
			c.cleanupFork(ctx, opts.NewSandboxId, snapshotPath)
			return nil, fmt.Errorf("failed to restore forked VM: %w", err)
		}
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

// restoreWithNetFds performs vm.restore with TAP file descriptor passing
// This enables true live fork with memory state in local mode
func (c *Client) restoreWithNetFds(ctx context.Context, sandboxId, snapshotPath string, netns *NetNamespace, netId string, prefault bool) error {
	socketPath := c.getSocketPath(sandboxId)

	log.Infof("Attempting live restore with net_fds for %s in namespace %s (net_id=%s)", sandboxId, netns.NamespaceName, netId)

	// Step 1: Open the TAP device (tap0) in the namespace
	// We need to enter the namespace, open /dev/net/tun, and attach to tap0
	tapFd, err := c.openTapInNamespace(netns.NamespaceName, netns.TapName)
	if err != nil {
		log.Warnf("Failed to open TAP device for live fork: %v, falling back to cold fork", err)
		return c.restoreWithoutNetFds(ctx, sandboxId, snapshotPath, prefault)
	}

	log.Infof("Opened TAP device %s in namespace %s, fd=%d", netns.TapName, netns.NamespaceName, tapFd)

	// Step 2: Connect to CH's Unix socket
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		unix.Close(tapFd) // Close TAP on error
		return fmt.Errorf("failed to connect to CH socket: %w", err)
	}

	// Step 3: Build the restore request with net_fds
	// The net_fds array tells CH which network device gets which FD
	// Format: [{"id": "<net_id>", "fds": [fd_index]}]
	// The FD is passed via SCM_RIGHTS, fd_index refers to position in the rights array
	restoreReq := RestoreConfig{
		SourceUrl: fmt.Sprintf("file://%s", snapshotPath),
		Prefault:  prefault,
		NetFds: []NetFd{
			{Id: netId, Fds: []int{0}}, // Use the network device ID from source VM
		},
	}

	reqBody, err := json.Marshal(restoreReq)
	if err != nil {
		unix.Close(tapFd)
		conn.Close()
		return fmt.Errorf("failed to marshal restore request: %w", err)
	}

	log.Debugf("Restore request body: %s", string(reqBody))

	// Step 4: Send HTTP request with FD via SCM_RIGHTS
	err = c.sendRestoreWithFd(conn, reqBody, tapFd)
	conn.Close()

	if err != nil {
		unix.Close(tapFd) // Close TAP before falling back
		log.Warnf("Failed to send restore with FD: %v, falling back to cold fork", err)
		return c.restoreWithoutNetFds(ctx, sandboxId, snapshotPath, prefault)
	}

	// Keep the TAP FD open - CH now owns it via SCM_RIGHTS
	// Note: we shouldn't close it here as CH is using it

	log.Infof("Live restore completed for %s with net_fds", sandboxId)
	return nil
}

// openTapInNamespace opens the TAP device in the specified network namespace
func (c *Client) openTapInNamespace(nsName, tapName string) (int, error) {
	// Lock this goroutine to an OS thread since we're changing namespaces
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Get current namespace FD
	currentNsPath := "/proc/self/ns/net"
	currentNsFd, err := unix.Open(currentNsPath, unix.O_RDONLY, 0)
	if err != nil {
		return -1, fmt.Errorf("failed to open current netns: %w", err)
	}
	defer unix.Close(currentNsFd)

	// Open target namespace
	nsPath := fmt.Sprintf("/var/run/netns/%s", nsName)
	nsFd, err := unix.Open(nsPath, unix.O_RDONLY, 0)
	if err != nil {
		return -1, fmt.Errorf("failed to open namespace %s: %w", nsName, err)
	}
	defer unix.Close(nsFd)

	// Enter the namespace
	if err := unix.Setns(nsFd, unix.CLONE_NEWNET); err != nil {
		return -1, fmt.Errorf("failed to enter namespace: %w", err)
	}

	// Open /dev/net/tun
	tunFd, err := unix.Open(devNetTun, unix.O_RDWR, 0)
	if err != nil {
		// Restore original namespace before returning error
		unix.Setns(currentNsFd, unix.CLONE_NEWNET)
		return -1, fmt.Errorf("failed to open %s: %w", devNetTun, err)
	}

	// Configure the interface with TUNSETIFF
	var ifr ifreq
	copy(ifr.name[:], tapName)
	// Use TAP mode without multi-queue (simpler and more compatible)
	ifr.flags = iffTap | iffNoPi

	log.Debugf("TUNSETIFF: opening TAP %s in namespace %s", tapName, nsName)

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(tunFd), tunSetIff, uintptr(unsafe.Pointer(&ifr)))
	if errno != 0 {
		unix.Close(tunFd)
		unix.Setns(currentNsFd, unix.CLONE_NEWNET)
		return -1, fmt.Errorf("TUNSETIFF ioctl failed: %v (tap=%s, flags=0x%x)", errno, tapName, ifr.flags)
	}

	// Return to original namespace
	if err := unix.Setns(currentNsFd, unix.CLONE_NEWNET); err != nil {
		log.Warnf("Failed to restore original namespace: %v", err)
	}

	return tunFd, nil
}

// sendRestoreWithFd sends the restore HTTP request with the TAP FD via SCM_RIGHTS
func (c *Client) sendRestoreWithFd(conn net.Conn, reqBody []byte, tapFd int) error {
	// Get the underlying file descriptor for the Unix socket
	unixConn, ok := conn.(*net.UnixConn)
	if !ok {
		return fmt.Errorf("connection is not a Unix socket")
	}

	file, err := unixConn.File()
	if err != nil {
		return fmt.Errorf("failed to get socket file: %w", err)
	}
	defer file.Close()
	sockFd := int(file.Fd())

	// Build HTTP request
	httpReq := fmt.Sprintf("PUT /api/v1/vm.restore HTTP/1.1\r\n"+
		"Host: localhost\r\n"+
		"Content-Type: application/json\r\n"+
		"Content-Length: %d\r\n"+
		"\r\n"+
		"%s", len(reqBody), string(reqBody))

	// Send the request with the FD attached via SCM_RIGHTS
	rights := syscall.UnixRights(tapFd)
	err = syscall.Sendmsg(sockFd, []byte(httpReq), rights, nil, 0)
	if err != nil {
		return fmt.Errorf("sendmsg failed: %w", err)
	}

	// Read response
	buf := make([]byte, 4096)
	n, err := syscall.Read(sockFd, buf)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	response := string(buf[:n])
	log.Debugf("CH restore response: %s", response)

	// Check for HTTP success status
	if !strings.Contains(response, "200") && !strings.Contains(response, "204") {
		return fmt.Errorf("restore failed: %s", response)
	}

	return nil
}

// restoreWithoutNetFds performs a cold restore (disk only, fresh boot)
func (c *Client) restoreWithoutNetFds(ctx context.Context, sandboxId, snapshotPath string, prefault bool) error {
	socketPath := c.getSocketPath(sandboxId)

	log.Infof("Performing cold restore (without net_fds) for %s", sandboxId)

	// Build restore request without net_fds
	restoreReq := fmt.Sprintf(`{"source_url":"file://%s","prefault":%v}`, snapshotPath, prefault)

	// Use curl to call the CH API
	curlCmd := fmt.Sprintf(`curl -s -X PUT -H "Content-Type: application/json" --unix-socket "%s" -d '%s' "http://localhost/api/v1/vm.restore"`,
		socketPath, restoreReq)

	output, err := c.runShellScript(ctx, curlCmd)
	if err != nil {
		return fmt.Errorf("vm.restore failed: %w (output: %s)", err, output)
	}

	output = strings.TrimSpace(output)
	if strings.Contains(strings.ToLower(output), "error") {
		return fmt.Errorf("vm.restore returned error: %s", output)
	}

	return nil
}
