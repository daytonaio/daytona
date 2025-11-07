package sdisk

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// QCowClient handles QCOW2 operations using qemu-img and qemu-nbd
type QCowClient struct {
	mu             sync.Mutex
	nbdDevices     map[string]string // volume name -> NBD device path
	mountedVolumes map[string]string // volume name -> mount path
}

// NewClient creates a new QCOW2 client
func NewQCowClient() (*QCowClient, error) {
	// Check if qemu-img and qemu-nbd are available
	if err := checkDependencies(); err != nil {
		return nil, err
	}

	return &QCowClient{
		nbdDevices:     make(map[string]string),
		mountedVolumes: make(map[string]string),
	}, nil
}

// checkDependencies verifies that qemu-img and qemu-nbd are installed
func checkDependencies() error {
	// Check qemu-img
	if _, err := exec.LookPath("qemu-img"); err != nil {
		return fmt.Errorf("qemu-img not found: %w", err)
	}

	// Check qemu-nbd
	if _, err := exec.LookPath("qemu-nbd"); err != nil {
		return fmt.Errorf("qemu-nbd not found: %w", err)
	}

	// Check if NBD module is loaded
	if _, err := os.Stat("/sys/module/nbd"); os.IsNotExist(err) {
		return fmt.Errorf("NBD kernel module not loaded (run: modprobe nbd)")
	}

	return nil
}

// CreateConfig contains options for creating a QCOW2 image
type QCowCreateConfig struct {
	Path          string
	SizeGB        int
	Compression   string // "zlib", "zstd", or ""
	ClusterSize   int    // bytes
	LazyRefcounts bool
	Preallocation string // "off", "metadata", "falloc", "full"
}

// Create creates a new QCOW2 image
func (c *QCowClient) Create(ctx context.Context, cfg QCowCreateConfig) error {
	args := []string{"create", "-f", "qcow2"}

	// Build options
	var opts []string

	if cfg.Compression != "" {
		opts = append(opts, fmt.Sprintf("compression_type=%s", cfg.Compression))
	}

	if cfg.ClusterSize > 0 {
		opts = append(opts, fmt.Sprintf("cluster_size=%d", cfg.ClusterSize))
	}

	if cfg.LazyRefcounts {
		opts = append(opts, "lazy_refcounts=on")
	}

	if cfg.Preallocation != "" && cfg.Preallocation != "off" {
		opts = append(opts, fmt.Sprintf("preallocation=%s", cfg.Preallocation))
	}

	if len(opts) > 0 {
		args = append(args, "-o", strings.Join(opts, ","))
	}

	args = append(args, cfg.Path, fmt.Sprintf("%dG", cfg.SizeGB))

	cmd := exec.CommandContext(ctx, "qemu-img", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create QCOW2 image: %w, output: %s", err, string(output))
	}

	return nil
}

// ImageInfo contains information about a QCOW2 image
type ImageInfo struct {
	Filename              string                 `json:"filename"`
	Format                string                 `json:"format"`
	VirtualSize           int64                  `json:"virtual-size"`
	ActualSize            int64                  `json:"actual-size"`
	ClusterSize           int                    `json:"cluster-size"`
	DirtyFlag             bool                   `json:"dirty-flag"`
	CompressionType       string                 `json:"compression-type"`
	BackingFilename       string                 `json:"backing-filename"`
	FullBackingFilename   string                 `json:"full-backing-filename"`
	BackingFilenameFormat string                 `json:"backing-filename-format"`
	FormatSpecific        map[string]interface{} `json:"format-specific"`
}

// Info returns information about a QCOW2 image
func (c *QCowClient) Info(ctx context.Context, imagePath string) (*ImageInfo, error) {
	cmd := exec.CommandContext(ctx, "qemu-img", "info", "--output=json", imagePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get image info: %w, output: %s", err, string(output))
	}

	var info ImageInfo
	if err := json.Unmarshal(output, &info); err != nil {
		return nil, fmt.Errorf("failed to parse image info: %w", err)
	}

	return &info, nil
}

// Resize resizes a QCOW2 image
func (c *QCowClient) Resize(ctx context.Context, imagePath string, newSizeGB int) error {
	cmd := exec.CommandContext(ctx, "qemu-img", "resize", imagePath, fmt.Sprintf("%dG", newSizeGB))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to resize image: %w, output: %s", err, string(output))
	}

	return nil
}

// findAvailableNBD finds an available NBD device
func (c *QCowClient) findAvailableNBD(ctx context.Context) (string, error) {
	// Get max_part value to determine device count
	maxDevices := 1024 // Default high number

	// Try to read max NBD devices from sysfs
	if data, err := os.ReadFile("/sys/module/nbd/parameters/nbds_max"); err == nil {
		if n, err := fmt.Sscanf(strings.TrimSpace(string(data)), "%d", &maxDevices); err == nil && n == 1 {
			// Successfully read the value
		}
	}

	// Limit search to reasonable number
	if maxDevices > 1024 {
		maxDevices = 1024
	}

	// Try devices from nbd0 up to maxDevices
	for i := 0; i < maxDevices; i++ {
		device := fmt.Sprintf("/dev/nbd%d", i)

		// Check if device exists
		if _, err := os.Stat(device); os.IsNotExist(err) {
			continue
		}

		// Check if device is in use by trying to read the size
		// If the device is not connected, size will be 0
		sizeFile := fmt.Sprintf("/sys/block/nbd%d/size", i)
		if data, err := os.ReadFile(sizeFile); err == nil {
			size := strings.TrimSpace(string(data))
			// Size of "0" means device is not in use
			if size == "0" {
				return device, nil
			}
		}
	}

	return "", fmt.Errorf("no available NBD devices found (tried %d devices)", maxDevices)
}

// findImageNBDDevice checks if an image is already connected to an NBD device
// by checking running qemu-nbd processes and their NBD devices
func (c *QCowClient) findImageNBDDevice(ctx context.Context, imagePath string) (string, error) {
	// Get absolute path for comparison
	absPath, err := filepath.Abs(imagePath)
	if err != nil {
		absPath = imagePath
	}

	// Check all NBD devices to see if any are using this image
	for i := 0; i < 256; i++ {
		device := fmt.Sprintf("/dev/nbd%d", i)

		// Check if device exists
		if _, err := os.Stat(device); os.IsNotExist(err) {
			continue
		}

		// Check if this device is connected (has non-zero size)
		sizeFile := fmt.Sprintf("/sys/block/nbd%d/size", i)
		sizeData, err := os.ReadFile(sizeFile)
		if err != nil {
			continue
		}

		size := strings.TrimSpace(string(sizeData))
		if size == "0" {
			continue // Device not connected
		}

		// Device is connected, check if it's using our image
		// We can check the backing file via qemu-nbd info
		cmd := exec.CommandContext(ctx, "qemu-nbd", "--show", device)
		output, err := cmd.Output()
		if err != nil {
			continue
		}

		deviceImage := strings.TrimSpace(string(output))
		if deviceImage == imagePath || deviceImage == absPath {
			return device, nil
		}
	}

	return "", nil
}

// killStaleQemuNBD finds and kills qemu-nbd processes holding the image file
func (c *QCowClient) killStaleQemuNBD(ctx context.Context, imagePath string) error {
	// Use lsof to find processes holding the file
	cmd := exec.CommandContext(ctx, "lsof", "-t", imagePath)
	output, err := cmd.Output()
	if err != nil {
		// No processes found or lsof failed
		return nil
	}

	// Parse PIDs
	pids := strings.Fields(strings.TrimSpace(string(output)))
	for _, pid := range pids {
		// Check if it's a qemu-nbd process
		cmdlineFile := fmt.Sprintf("/proc/%s/cmdline", pid)
		cmdlineData, err := os.ReadFile(cmdlineFile)
		if err != nil {
			continue
		}

		cmdline := string(cmdlineData)
		if strings.Contains(cmdline, "qemu-nbd") {
			// Kill the process
			killCmd := exec.CommandContext(ctx, "kill", "-9", pid)
			if err := killCmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to kill qemu-nbd process %s: %v\n", pid, err)
			} else {
				fmt.Fprintf(os.Stderr, "info: killed stale qemu-nbd process %s\n", pid)
			}
		}
	}

	return nil
}

// disconnectStaleNBD attempts to disconnect any stale NBD connection for an image
func (c *QCowClient) disconnectStaleNBD(ctx context.Context, imagePath string) error {
	device, err := c.findImageNBDDevice(ctx, imagePath)
	if err != nil || device == "" {
		// No NBD device found, but check for stale qemu-nbd processes
		return c.killStaleQemuNBD(ctx, imagePath)
	}

	// Disconnect the stale device
	cmd := exec.CommandContext(ctx, "qemu-nbd", "--disconnect", device)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to disconnect stale NBD device %s: %v, output: %s\n",
			device, err, string(output))
	}

	// Give the kernel a moment to clean up
	time.Sleep(100 * time.Millisecond)

	// Also kill any stale qemu-nbd processes that might still be holding the file
	if err := c.killStaleQemuNBD(ctx, imagePath); err != nil {
		fmt.Fprintf(os.Stderr, "warning: error killing stale qemu-nbd processes: %v\n", err)
	}

	// Wait a bit more for process cleanup
	time.Sleep(100 * time.Millisecond)

	return nil
}

// Connect connects a QCOW2 image to an NBD device
func (c *QCowClient) Connect(ctx context.Context, volumeName, imagePath string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if already connected
	if device, exists := c.nbdDevices[volumeName]; exists {
		return device, nil
	}

	// Check for and disconnect any stale NBD connections to this image
	// This handles cases where a previous process crashed or didn't clean up properly
	if err := c.disconnectStaleNBD(ctx, imagePath); err != nil {
		fmt.Fprintf(os.Stderr, "warning: error checking for stale NBD connections: %v\n", err)
	}

	// Check and repair the image if needed (after ensuring no processes are using it)
	checkCmd := exec.CommandContext(ctx, "qemu-img", "check", "-r", "all", imagePath)
	checkOutput, checkErr := checkCmd.CombinedOutput()
	if checkErr != nil {
		// Log but continue - qemu-img check returns non-zero even after successful repair
		fmt.Fprintf(os.Stderr, "info: qemu-img check output: %s\n", string(checkOutput))
	}

	// Find available NBD device using our improved detection
	device, err := c.findAvailableNBD(ctx)
	if err != nil {
		return "", err
	}

	// Connect the image to the NBD device
	cmd := exec.CommandContext(ctx, "qemu-nbd", "--connect="+device, imagePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to connect NBD device %s: %w, output: %s", device, err, string(output))
	}

	c.nbdDevices[volumeName] = device
	return device, nil
}

// Disconnect disconnects a QCOW2 image from an NBD device
func (c *QCowClient) Disconnect(ctx context.Context, volumeName string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	device, exists := c.nbdDevices[volumeName]
	if !exists {
		return nil // Already disconnected
	}

	cmd := exec.CommandContext(ctx, "qemu-nbd", "--disconnect", device)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to disconnect NBD device: %w, output: %s", err, string(output))
	}

	delete(c.nbdDevices, volumeName)
	return nil
}

// Mount mounts an NBD device to a directory
func (c *QCowClient) Mount(ctx context.Context, volumeName, device, mountPath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if already mounted
	if _, exists := c.mountedVolumes[volumeName]; exists {
		return fmt.Errorf("volume already mounted")
	}

	// Create mount directory if it doesn't exist
	if err := os.MkdirAll(mountPath, 0755); err != nil {
		return fmt.Errorf("failed to create mount directory: %w", err)
	}

	// Check if device has a filesystem, if not create ext4
	cmd := exec.CommandContext(ctx, "blkid", device)
	if err := cmd.Run(); err != nil {
		// No filesystem found, create ext4
		mkfsCmd := exec.CommandContext(ctx, "mkfs.ext4", "-F", device)
		output, err := mkfsCmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to create filesystem: %w, output: %s", err, string(output))
		}
	}

	// Mount the device
	mountCmd := exec.CommandContext(ctx, "mount", device, mountPath)
	output, err := mountCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to mount device: %w, output: %s", err, string(output))
	}

	c.mountedVolumes[volumeName] = mountPath
	return nil
}

// Unmount unmounts a volume with retry logic for busy filesystems
func (c *QCowClient) Unmount(ctx context.Context, volumeName string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	mountPath, exists := c.mountedVolumes[volumeName]
	if !exists {
		return nil // Already unmounted
	}

	// Retry parameters
	maxRetries := 5
	retryDelay := 500 * time.Millisecond

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			time.Sleep(retryDelay)
		}

		cmd := exec.CommandContext(ctx, "umount", mountPath)
		output, err := cmd.CombinedOutput()
		if err == nil {
			// Success!
			delete(c.mountedVolumes, volumeName)
			return nil
		}

		lastErr = err

		// Check if the error is "target is busy"
		outputStr := string(output)
		if strings.Contains(outputStr, "target is busy") || strings.Contains(outputStr, "device is busy") {
			// Filesystem is busy, will retry
			continue
		}

		// Other error - don't retry
		return fmt.Errorf("failed to unmount volume: %w, output: %s", err, outputStr)
	}

	// All retries exhausted
	return fmt.Errorf("failed to unmount volume after %d attempts (filesystem busy): %w", maxRetries, lastErr)
}

// IsMounted checks if a volume is currently mounted
func (c *QCowClient) IsMounted(volumeName string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, exists := c.mountedVolumes[volumeName]
	return exists
}

// GetMountPath returns the mount path for a volume (empty if not mounted)
func (c *QCowClient) GetMountPath(volumeName string) string {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.mountedVolumes[volumeName]
}

// Close cleans up all NBD connections and unmounts
func (c *QCowClient) Close(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var errs []error

	// Unmount all volumes
	for volumeName := range c.mountedVolumes {
		mountPath := c.mountedVolumes[volumeName]
		cmd := exec.CommandContext(ctx, "umount", mountPath)
		if err := cmd.Run(); err != nil {
			errs = append(errs, fmt.Errorf("failed to unmount %s: %w", volumeName, err))
		}
	}

	// Disconnect all NBD devices
	for volumeName, device := range c.nbdDevices {
		cmd := exec.CommandContext(ctx, "qemu-nbd", "--disconnect", device)
		if err := cmd.Run(); err != nil {
			errs = append(errs, fmt.Errorf("failed to disconnect %s: %w", volumeName, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during close: %v", errs)
	}

	return nil
}

// Checksum calculates SHA256 checksum of a QCOW2 image
func (c *QCowClient) Checksum(ctx context.Context, imagePath string) (string, error) {
	cmd := exec.CommandContext(ctx, "sha256sum", imagePath)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to calculate checksum: %w", err)
	}

	// Parse output: "checksum  filename"
	parts := strings.Fields(string(output))
	if len(parts) < 1 {
		return "", fmt.Errorf("invalid checksum output")
	}

	return parts[0], nil
}

// GetActualSize returns the actual disk usage of a QCOW2 image
func (c *QCowClient) GetActualSize(ctx context.Context, imagePath string) (int64, error) {
	info, err := c.Info(ctx, imagePath)
	if err != nil {
		return 0, err
	}

	return info.ActualSize, nil
}

// GetVirtualSize returns the virtual size of a QCOW2 image
func (c *QCowClient) GetVirtualSize(ctx context.Context, imagePath string) (int64, error) {
	info, err := c.Info(ctx, imagePath)
	if err != nil {
		return 0, err
	}

	return info.VirtualSize, nil
}

// CreateSnapshot creates a new QCOW2 image with the specified backing file
// The new image will be a CoW (Copy-on-Write) layer on top of the backing file
func (c *QCowClient) CreateSnapshot(ctx context.Context, backingFile, snapshotPath string) error {
	args := []string{
		"create",
		"-f", "qcow2",
		"-F", "qcow2", // Backing file format
		"-b", backingFile, // Backing file path
		snapshotPath,
	}

	cmd := exec.CommandContext(ctx, "qemu-img", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create snapshot: %w, output: %s", err, string(output))
	}

	return nil
}

// CreateWithBacking creates a new QCOW2 image with a backing file and specific size
// This creates an empty delta layer that references the backing file
func (c *QCowClient) CreateWithBacking(ctx context.Context, backingFile, newImagePath string, sizeGB int) error {
	args := []string{
		"create",
		"-f", "qcow2",
		"-F", "qcow2", // Backing file format
		"-b", backingFile, // Backing file path
		newImagePath,
		fmt.Sprintf("%dG", sizeGB),
	}

	cmd := exec.CommandContext(ctx, "qemu-img", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create QCOW2 with backing: %w, output: %s", err, string(output))
	}

	return nil
}

// Rebase changes the backing file of a QCOW2 image
// This is useful when you need to update the backing file path or commit changes
func (c *QCowClient) Rebase(ctx context.Context, imagePath, newBackingFile string) error {
	args := []string{
		"rebase",
		"-f", "qcow2",
		"-F", "qcow2",
		"-b", newBackingFile,
		imagePath,
	}

	cmd := exec.CommandContext(ctx, "qemu-img", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to rebase image: %w, output: %s", err, string(output))
	}

	return nil
}

// RebaseUnsafe updates the backing file path without checking block content
// This is much faster but assumes the backing file is correct
func (c *QCowClient) RebaseUnsafe(ctx context.Context, imagePath, newBackingFile string) error {
	args := []string{
		"rebase",
		"-u", // Unsafe mode - just update path, don't check content
		"-f", "qcow2",
	}

	// Only add backing format and file if we're setting a backing file
	// If newBackingFile is empty, we're removing the backing file
	if newBackingFile != "" {
		args = append(args, "-F", "qcow2", "-b", newBackingFile)
	} else {
		// To remove backing file, use -b ""
		args = append(args, "-b", "")
	}

	args = append(args, imagePath)

	cmd := exec.CommandContext(ctx, "qemu-img", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to rebase image (unsafe): %w, output: %s", err, string(output))
	}

	return nil
}

// Commit merges a QCOW2 image into its backing file
func (c *QCowClient) Commit(ctx context.Context, imagePath string) error {
	cmd := exec.CommandContext(ctx, "qemu-img", "commit", imagePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to commit image: %w, output: %s", err, string(output))
	}

	return nil
}

// GetBackingFile returns the backing file path of a QCOW2 image (if any)
func (c *QCowClient) GetBackingFile(ctx context.Context, imagePath string) (string, error) {
	info, err := c.Info(ctx, imagePath)
	if err != nil {
		return "", err
	}

	// The backing filename is now directly in the ImageInfo struct
	return info.BackingFilename, nil
}

// ValidateBackingChain checks for circular references in the backing file chain
// Returns an error if a circular reference is detected or chain is too deep
func (c *QCowClient) ValidateBackingChain(ctx context.Context, imagePath string) error {
	visited := make(map[string]bool)
	currentPath := imagePath
	const maxChainDepth = 1000 // Safety limit to prevent extremely deep chains

	// Get absolute path for comparison
	absImagePath, err := filepath.Abs(imagePath)
	if err != nil {
		absImagePath = imagePath
	}

	depth := 0
	for {
		// Safety check: prevent infinite loops from extremely deep chains
		if depth >= maxChainDepth {
			return fmt.Errorf("backing file chain exceeds maximum depth of %d (possible circular reference)", maxChainDepth)
		}
		depth++

		// Check if we've seen this file before (circular reference)
		absCurrentPath, err := filepath.Abs(currentPath)
		if err != nil {
			absCurrentPath = currentPath
		}

		if visited[absCurrentPath] {
			return fmt.Errorf("circular backing file reference detected: %s references itself in the chain", currentPath)
		}
		visited[absCurrentPath] = true

		// Get backing file
		backingFile, err := c.GetBackingFile(ctx, currentPath)
		if err != nil {
			return fmt.Errorf("failed to get backing file for %s: %w", currentPath, err)
		}

		// If no backing file, we've reached the end of the chain
		if backingFile == "" {
			break
		}

		// Resolve relative paths
		if !filepath.IsAbs(backingFile) {
			backingFile = filepath.Join(filepath.Dir(currentPath), backingFile)
		}

		// Check if backing file points back to the original image
		absBackingFile, err := filepath.Abs(backingFile)
		if err != nil {
			absBackingFile = backingFile
		}

		if absBackingFile == absImagePath {
			return fmt.Errorf("circular backing file reference: backing file %s points back to source image %s", backingFile, imagePath)
		}

		currentPath = backingFile
	}

	return nil
}

// Convert converts a QCOW2 image to standalone (no backing file)
// This collapses all layers into a single image
func (c *QCowClient) Convert(ctx context.Context, sourcePath, destPath string) error {
	// Validate backing chain before conversion to detect circular references
	if err := c.ValidateBackingChain(ctx, sourcePath); err != nil {
		return fmt.Errorf("invalid backing file chain: %w", err)
	}

	args := []string{
		"convert",
		"-f", "qcow2",
		"-O", "qcow2",
		sourcePath,
		destPath,
	}

	cmd := exec.CommandContext(ctx, "qemu-img", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to convert image: %w, output: %s", err, string(output))
	}

	return nil
}
