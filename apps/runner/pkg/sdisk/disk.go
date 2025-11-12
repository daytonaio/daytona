package sdisk

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"golang.org/x/sys/unix"
)

// disk implements the Disk interface
type disk struct {
	name        string
	sizeGB      int64
	imagePath   string
	qcow2Client *QCowClient
	s3Client    *S3Client
	stateDB     *DB
	config      Config
	pool        *DiskPool // Reference to pool manager (if pooling enabled)
	isMounted   bool
	mountPath   string
	nbdDevice   string
	mu          sync.Mutex
}

func (v *disk) Name() string {
	return v.name
}

func (v *disk) Size() int64 {
	return v.sizeGB
}

func (v *disk) Mount(ctx context.Context) (string, error) {
	// If pooling is enabled, use the pool to manage mounting
	if v.pool != nil {
		vol, err := v.pool.Get(ctx, v)
		if err != nil {
			return "", err
		}
		mountPath := vol.MountPath()
		return mountPath, nil
	}

	// Otherwise, use direct mounting
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.isMounted {
		return v.mountPath, nil
	}

	// Connect QCOW2 image to NBD device
	device, err := v.qcow2Client.Connect(ctx, v.name, v.imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to connect NBD device: %w", err)
	}
	v.nbdDevice = device

	// Create mount path
	mountPath := filepath.Join(v.config.DataDir, "mounts", v.name)

	// Mount the NBD device
	if err := v.qcow2Client.Mount(ctx, v.name, device, mountPath); err != nil {
		// Cleanup NBD connection on failure
		v.qcow2Client.Disconnect(ctx, v.name)
		return "", fmt.Errorf("failed to mount disk: %w", err)
	}

	v.isMounted = true
	v.mountPath = mountPath

	// Update state in database
	if err := v.stateDB.UpdateMountState(v.name, true, mountPath); err != nil {
		// Log error but don't fail the mount
		fmt.Fprintf(os.Stderr, "warning: failed to update mount state for disk '%s': %v\n", v.name, err)
	}

	return mountPath, nil
}

// mountInternal performs the actual mounting logic without pool checks
// This is used by the pool to avoid infinite recursion
func (v *disk) mountInternal(ctx context.Context) (string, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	// DETAILED LOGGING
	logFile, _ := os.OpenFile("/tmp/mount-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if logFile != nil {
		fmt.Fprintf(logFile, "\n[DISK-MOUNTINTERNAL] Starting mount for disk=%s, imagePath=%s\n", v.name, v.imagePath)
		fmt.Fprintf(logFile, "[DISK-MOUNTINTERNAL] isMounted=%v, mountPath=%s\n", v.isMounted, v.mountPath)
		defer logFile.Close()
	}

	if v.isMounted {
		if logFile != nil {
			fmt.Fprintf(logFile, "[DISK-MOUNTINTERNAL] Disk already mounted, returning %s\n", v.mountPath)
		}
		return v.mountPath, nil
	}

	// Connect QCOW2 image to NBD device
	if logFile != nil {
		fmt.Fprintf(logFile, "[DISK-MOUNTINTERNAL] Calling qcow2Client.Connect\n")
	}
	device, err := v.qcow2Client.Connect(ctx, v.name, v.imagePath)
	if err != nil {
		if logFile != nil {
			fmt.Fprintf(logFile, "[DISK-MOUNTINTERNAL] ERROR: Connect failed: %v\n", err)
		}
		return "", fmt.Errorf("failed to connect NBD device: %w", err)
	}
	v.nbdDevice = device
	if logFile != nil {
		fmt.Fprintf(logFile, "[DISK-MOUNTINTERNAL] NBD device: %s\n", device)
	}

	// Create mount path
	mountPath := filepath.Join(v.config.DataDir, "mounts", v.name)
	if logFile != nil {
		fmt.Fprintf(logFile, "[DISK-MOUNTINTERNAL] Mount path: %s\n", mountPath)
	}

	// Mount the NBD device
	if logFile != nil {
		fmt.Fprintf(logFile, "[DISK-MOUNTINTERNAL] Calling qcow2Client.Mount\n")
	}
	if err := v.qcow2Client.Mount(ctx, v.name, device, mountPath); err != nil {
		// Cleanup NBD connection on failure
		if logFile != nil {
			fmt.Fprintf(logFile, "[DISK-MOUNTINTERNAL] ERROR: Mount failed: %v, disconnecting NBD\n", err)
		}
		v.qcow2Client.Disconnect(ctx, v.name)
		return "", fmt.Errorf("failed to mount disk: %w", err)
	}

	v.isMounted = true
	v.mountPath = mountPath
	if logFile != nil {
		fmt.Fprintf(logFile, "[DISK-MOUNTINTERNAL] Mount succeeded, updating state\n")
	}

	// Update state in database
	if err := v.stateDB.UpdateMountState(v.name, true, mountPath); err != nil {
		// Log error but don't fail the mount
		fmt.Fprintf(os.Stderr, "warning: failed to update mount state for disk '%s': %v\n", v.name, err)
		if logFile != nil {
			fmt.Fprintf(logFile, "[DISK-MOUNTINTERNAL] WARNING: failed to update database: %v\n", err)
		}
	}

	if logFile != nil {
		fmt.Fprintf(logFile, "[DISK-MOUNTINTERNAL] SUCCESS: Disk %s mounted at %s\n", v.name, mountPath)
	}
	return mountPath, nil
}

func (v *disk) Unmount(ctx context.Context) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	// DETAILED LOGGING - Find who's calling Unmount
	unmountLog, _ := os.OpenFile("/tmp/unmount-trace.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if unmountLog != nil {
		fmt.Fprintf(unmountLog, "\n[UNMOUNT-CALLED] disk=%s, isMounted=%v, mountPath=%s\n", v.name, v.isMounted, v.mountPath)
		// Print stack trace to see who called this
		buf := make([]byte, 4096)
		n := runtime.Stack(buf, false)
		fmt.Fprintf(unmountLog, "[UNMOUNT-STACK]:\n%s\n", buf[:n])
		unmountLog.Close()
	}

	// Check if disk is mounted (using pool-aware check)
	// For pooled disks, mountPath being set indicates the disk is mounted
	isActuallyMounted := v.isMounted || (v.pool != nil && v.mountPath != "")
	if !isActuallyMounted {
		return nil
	}

	// Log to debug file
	logFile, _ := os.OpenFile("/tmp/fork-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if logFile != nil {
		fmt.Fprintf(logFile, "[DISK-UNMOUNT] Starting unmount for disk %s at mountPath=%s\n", v.name, v.mountPath)
		// List files before unmount
		if entries, err := os.ReadDir(v.mountPath); err == nil {
			fmt.Fprintf(logFile, "[DISK-UNMOUNT] Files in mount before unmount:\n")
			for _, entry := range entries {
				info, _ := entry.Info()
				fmt.Fprintf(logFile, "  - %s (size: %d, isDir: %v)\n", entry.Name(), info.Size(), entry.IsDir())
			}
		} else {
			fmt.Fprintf(logFile, "[DISK-UNMOUNT] Failed to list mount: %v\n", err)
		}
		logFile.Close()
	}

	// Unmount the filesystem
	if err := v.qcow2Client.Unmount(ctx, v.name); err != nil {
		return fmt.Errorf("failed to unmount filesystem: %w", err)
	}

	// Disconnect NBD device
	if err := v.qcow2Client.Disconnect(ctx, v.name); err != nil {
		return fmt.Errorf("failed to disconnect NBD device: %w", err)
	}

	v.isMounted = false
	v.mountPath = ""
	v.nbdDevice = ""

	// Update state in database
	if err := v.stateDB.UpdateMountState(v.name, false, ""); err != nil {
		// Log error but don't fail the unmount
		fmt.Fprintf(os.Stderr, "warning: failed to update mount state: %v\n", err)
	}

	return nil
}

func (v *disk) IsMounted() bool {
	v.mu.Lock()
	defer v.mu.Unlock()

	// If pooling is enabled, check if disk is in the pool
	// The pool manages mounting separately from the disk object's isMounted flag
	// Use mountPath as a heuristic: if it's set, the disk is mounted via pool
	if v.pool != nil && v.mountPath != "" {
		return true
	}

	return v.isMounted
}

func (v *disk) MountPath() string {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.mountPath
}

// trimFilesystem runs fstrim on the mounted filesystem to free unused blocks
// This is critical for reducing QCOW2 layer sizes before committing/forking
func (v *disk) trimFilesystem(ctx context.Context, mountPath string) error {
	if mountPath == "" {
		return fmt.Errorf("mount path is empty")
	}

	// Run fstrim on the mounted filesystem
	// -v for verbose output showing how much space was freed
	cmd := exec.CommandContext(ctx, "fstrim", "-v", mountPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("fstrim failed: %w, output: %s", err, string(output))
	}

	fmt.Fprintf(os.Stderr, "[TRIM] Disk %s freed space: %s", v.name, string(output))
	return nil
}

func (v *disk) Sync(ctx context.Context) error {
	v.mu.Lock()
	mountPath := v.mountPath
	nbdDevice := v.nbdDevice
	v.mu.Unlock()

	if mountPath == "" {
		return fmt.Errorf("disk %s is not mounted", v.name)
	}

	// Step 0: Trim freed filesystem blocks (optimization for QCOW2 layer size)
	// This sends DISCARD commands to qemu-nbd, which deallocates blocks in the QCOW2 file
	// Critical for reducing layer sizes before fork/commit operations
	if err := v.trimFilesystem(ctx, mountPath); err != nil {
		// Don't fail sync on trim error - it's an optimization, not critical
		fmt.Fprintf(os.Stderr, "[TRIM] Warning: trim failed for disk %s: %v\n", v.name, err)
	}

	// Step 1: Sync the filesystem using syncfs (flushes filesystem buffers to NBD device)
	if f, err := os.Open(mountPath); err == nil {
		defer f.Close()
		if err := unix.Syncfs(int(f.Fd())); err != nil {
			return fmt.Errorf("failed to sync filesystem at %s: %w", mountPath, err)
		}
	} else {
		return fmt.Errorf("failed to open mount path %s: %w", mountPath, err)
	}

	// Step 2: CRITICAL - Flush NBD device buffers to qemu-nbd
	// This ensures data is written from the NBD layer to the QCOW2 file
	syncLog, _ := os.OpenFile("/tmp/sync-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if syncLog != nil {
		fmt.Fprintf(syncLog, "\n[SYNC] Disk=%s, NBD device=%s\n", v.name, nbdDevice)
		defer syncLog.Close()
	}

	if nbdDevice != "" {
		if f, err := os.OpenFile(nbdDevice, os.O_RDWR, 0); err == nil {
			if syncLog != nil {
				fmt.Fprintf(syncLog, "[SYNC] Opened NBD device %s for flushing\n", nbdDevice)
			}
			// Issue BLKFLSBUF ioctl to flush NBD buffers
			// This flushes qemu-nbd's internal cache to the QCOW2 file
			_, _, errno := unix.Syscall(unix.SYS_IOCTL, f.Fd(), unix.BLKFLSBUF, 0)
			if syncLog != nil {
				fmt.Fprintf(syncLog, "[SYNC] BLKFLSBUF ioctl result: errno=%v\n", errno)
			}
			f.Sync() // Also call fsync on the device
			f.Close()
			if errno != 0 {
				if syncLog != nil {
					fmt.Fprintf(syncLog, "[SYNC] ERROR: BLKFLSBUF failed with errno %v\n", errno)
				}
				return fmt.Errorf("failed to flush NBD device %s: %v", nbdDevice, errno)
			}
			if syncLog != nil {
				fmt.Fprintf(syncLog, "[SYNC] Successfully flushed NBD device %s\n", nbdDevice)
			}
		} else {
			if syncLog != nil {
				fmt.Fprintf(syncLog, "[SYNC] ERROR: Failed to open NBD device: %v\n", err)
			}
			return fmt.Errorf("failed to open NBD device %s: %w", nbdDevice, err)
		}
	} else {
		if syncLog != nil {
			fmt.Fprintf(syncLog, "[SYNC] WARNING: No NBD device set for disk %s\n", v.name)
		}
	}

	// Step 3: Run system-wide sync as final safety measure
	syncCmd := exec.CommandContext(ctx, "sync")
	if err := syncCmd.Run(); err != nil {
		// Don't fail on sync error, but log it
		fmt.Fprintf(os.Stderr, "warning: system sync failed: %v\n", err)
	}

	return nil
}

func (v *disk) Push(ctx context.Context) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.isMounted {
		return fmt.Errorf("cannot push mounted disk: unmount it first")
	}

	if v.s3Client == nil {
		return ErrS3NotConfigured
	}

	// Get current state
	state, err := v.stateDB.GetDisk(v.name)
	if err != nil {
		return fmt.Errorf("failed to get disk state: %w", err)
	}

	// Try to download existing metadata to check if disk already exists in S3
	existingMetadata, err := v.s3Client.DownloadMetadata(ctx, v.name)
	if err != nil {
		// Disk doesn't exist in S3 - this is the first push (base layer)
		return v.pushBaseLayer(ctx, state)
	}

	// Disk exists in S3 - create an incremental snapshot
	return v.pushIncrementalLayer(ctx, state, existingMetadata)
}

// pushBaseLayer uploads the base layer (first push of a disk)
func (v *disk) pushBaseLayer(ctx context.Context, state *DiskState) error {
	// CRITICAL: We need to commit the current working image as a layer first
	// This ensures the data is preserved in the layer cache and can be uploaded

	// Check if this disk has layer mappings (e.g., from a fork operation)
	diskLayers, err := v.stateDB.GetDiskLayers(v.name)
	if err != nil {
		return fmt.Errorf("failed to get disk layers: %w", err)
	}

	// If we DON'T have layer mappings yet, create one for the current working image
	if len(diskLayers) == 0 {
		// This is a newly created disk without layer structure yet
		// We need to commit the working image as a layer first
		if err := v.commitWorkingImageAsLayer(ctx); err != nil {
			return fmt.Errorf("failed to commit working image: %w", err)
		}

		// Now get the updated layer mappings
		diskLayers, err = v.stateDB.GetDiskLayers(v.name)
		if err != nil {
			return fmt.Errorf("failed to get disk layers after commit: %w", err)
		}
	}

	// Use layer-aware push for ALL disks (not just forked ones)
	return v.pushLayeredDisk(ctx, state, diskLayers)
}

// walkBackingChain walks through the backing file chain and returns layer IDs
// of all cached layers in the chain (from base to top, excluding the working layer)
func (v *disk) walkBackingChain(ctx context.Context, imagePath, layerCacheDir string) ([]string, error) {
	var layerIDs []string
	currentPath := imagePath

	// Get the first backing file (the working layer's backing)
	backingFile, err := v.qcow2Client.GetBackingFile(ctx, currentPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get backing file: %w", err)
	}

	// If no backing file, this is a standalone disk (no layers)
	if backingFile == "" {
		return layerIDs, nil
	}

	// Walk the backing chain
	currentPath = backingFile
	visited := make(map[string]bool)
	maxDepth := 100

	for depth := 0; depth < maxDepth; depth++ {
		// Prevent circular references
		absPath, _ := filepath.Abs(currentPath)
		if visited[absPath] {
			return nil, fmt.Errorf("circular reference detected in backing chain")
		}
		visited[absPath] = true

		// Check if this is a cached layer
		if strings.HasPrefix(currentPath, layerCacheDir) {
			// Extract layer ID from the path (e.g., /path/layer-cache/layer-12345.qcow2 -> layer-12345)
			basename := filepath.Base(currentPath)
			layerID := strings.TrimSuffix(basename, ".qcow2")

			// Verify the layer exists in the database
			layerState, err := v.stateDB.GetLayer(layerID)
			if err != nil || layerState == nil {
				return nil, fmt.Errorf("cached layer %s not found in database", layerID)
			}

			// Add to the beginning of the list (we're walking from top to base)
			layerIDs = append([]string{layerID}, layerIDs...)

			fmt.Fprintf(os.Stderr, "[WALK-CHAIN] Found cached layer: %s (size: %d)\n", layerID, layerState.Size)
		}

		// Get the next backing file
		nextBackingFile, err := v.qcow2Client.GetBackingFile(ctx, currentPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get backing file for %s: %w", currentPath, err)
		}

		// If no more backing files, we've reached the end
		if nextBackingFile == "" {
			break
		}

		// Resolve relative paths
		if !filepath.IsAbs(nextBackingFile) {
			nextBackingFile = filepath.Join(filepath.Dir(currentPath), nextBackingFile)
		}

		currentPath = nextBackingFile
	}

	return layerIDs, nil
}

// commitWorkingImageAsLayer preserves the layer structure and commits the working layer as a delta
// This allows proper layer reuse and deduplication when pushing to S3
func (v *disk) commitWorkingImageAsLayer(ctx context.Context) error {
	// Get existing layer mappings (from fork or previous commits)
	oldLayers, err := v.stateDB.GetDiskLayers(v.name)
	if err != nil {
		return fmt.Errorf("failed to get old disk layers: %w", err)
	}

	// Get the backing chain to preserve layer structure
	layerCacheDir := filepath.Join(v.config.DataDir, "layer-cache")
	backingChain, err := v.walkBackingChain(ctx, v.imagePath, layerCacheDir)
	if err != nil {
		return fmt.Errorf("failed to walk backing chain: %w", err)
	}

	fmt.Fprintf(os.Stderr, "[COMMIT] Found %d existing layer mappings, %d layers in backing chain\n",
		len(oldLayers), len(backingChain))

	// Check if we already have correct mappings (e.g., from fork)
	if len(oldLayers) > 0 && len(oldLayers) == len(backingChain) {
		// Verify mappings match the backing chain
		mappingsMatch := true
		for i, oldLayer := range oldLayers {
			if i >= len(backingChain) || oldLayer.LayerID != backingChain[i] {
				mappingsMatch = false
				break
			}
		}

		if mappingsMatch {
			fmt.Fprintf(os.Stderr, "[COMMIT] Layer mappings already correct, only adding new working layer\n")
			// Don't delete mappings, we'll just add the new working layer
		} else {
			// Mappings don't match, clean up and recreate
			fmt.Fprintf(os.Stderr, "[COMMIT] Layer mappings don't match backing chain, recreating...\n")
			for _, oldLayer := range oldLayers {
				if err := v.stateDB.DecrementLayerRefCount(oldLayer.LayerID); err != nil {
					fmt.Fprintf(os.Stderr, "warning: failed to decrement ref count for layer %s: %v\n", oldLayer.LayerID, err)
				}
			}
			if err := v.stateDB.DeleteDiskLayers(v.name); err != nil {
				return fmt.Errorf("failed to delete old disk-layer mappings: %w", err)
			}
			oldLayers = nil // Mark as needing recreation
		}
	} else if len(oldLayers) > 0 {
		// Have some mappings but count doesn't match backing chain
		fmt.Fprintf(os.Stderr, "[COMMIT] Layer mapping count mismatch, cleaning up...\n")
		for _, oldLayer := range oldLayers {
			if err := v.stateDB.DecrementLayerRefCount(oldLayer.LayerID); err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to decrement ref count for layer %s: %v\n", oldLayer.LayerID, err)
			}
		}
		if err := v.stateDB.DeleteDiskLayers(v.name); err != nil {
			return fmt.Errorf("failed to delete old disk-layer mappings: %w", err)
		}
		oldLayers = nil // Mark as needing recreation
	}

	// Create layer ID for the working layer
	layerID := fmt.Sprintf("layer-%d", time.Now().Unix())

	// Save the working layer (delta only) to cache
	// This preserves only the changes made in this disk, not the entire flattened data
	cachedLayerPath := filepath.Join(layerCacheDir, layerID+".qcow2")

	if err := v.copyFile(v.imagePath, cachedLayerPath); err != nil {
		return fmt.Errorf("failed to cache working layer: %w", err)
	}

	// CRITICAL: Remove backing file reference from the cached layer
	// This makes it a standalone delta-only layer that won't merge backing data when "flattened"
	// for S3 upload. We use RebaseUnsafe with empty backing file to strip the reference.
	if err := v.qcow2Client.RebaseUnsafe(ctx, cachedLayerPath, ""); err != nil {
		os.Remove(cachedLayerPath)
		return fmt.Errorf("failed to remove backing reference from cached layer: %w", err)
	}
	fmt.Fprintf(os.Stderr, "[COMMIT] Removed backing file reference from cached layer %s (now standalone delta)\n", layerID)

	// Get layer file info (delta layer only)
	fileInfo, err := os.Stat(cachedLayerPath)
	if err != nil {
		os.Remove(cachedLayerPath)
		return fmt.Errorf("failed to get layer file info: %w", err)
	}

	// Calculate checksum of the delta layer
	checksum, err := v.qcow2Client.Checksum(ctx, cachedLayerPath)
	if err != nil {
		os.Remove(cachedLayerPath)
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}

	// Create layer state for the working layer
	layerState := &LayerState{
		ID:       layerID,
		Checksum: checksum,
		Size:     fileInfo.Size(),
		CachedAt: time.Now(),
		RefCount: 1, // Referenced by this disk
	}

	if err := v.stateDB.SaveLayer(layerState); err != nil {
		os.Remove(cachedLayerPath)
		return fmt.Errorf("failed to save layer state: %w", err)
	}

	// Create disk-layer mappings if they don't already exist
	// If oldLayers is nil, we need to recreate all mappings
	// If oldLayers exists and matches, we only add the new working layer
	if oldLayers == nil {
		// Create mappings for ALL layers in the chain (preserving order)
		fmt.Fprintf(os.Stderr, "[COMMIT] Creating layer mappings for %d backing layers\n", len(backingChain))
		for i, backingLayerID := range backingChain {
			if err := v.stateDB.AddDiskLayerMapping(v.name, backingLayerID, i); err != nil {
				os.Remove(cachedLayerPath)
				v.stateDB.DecrementLayerRefCount(layerID)
				return fmt.Errorf("failed to add backing layer mapping for %s: %w", backingLayerID, err)
			}
			// Increment ref count for backing layers we're now referencing
			if err := v.stateDB.IncrementLayerRefCount(backingLayerID); err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to increment ref count for layer %s: %v\n", backingLayerID, err)
			}
			fmt.Fprintf(os.Stderr, "[COMMIT] Added backing layer mapping: position %d -> %s\n", i, backingLayerID)
		}
	} else {
		fmt.Fprintf(os.Stderr, "[COMMIT] Keeping existing %d layer mappings\n", len(oldLayers))
	}

	// Add the new working layer as the top layer
	topPosition := len(backingChain)
	if err := v.stateDB.AddDiskLayerMapping(v.name, layerID, topPosition); err != nil {
		os.Remove(cachedLayerPath)
		v.stateDB.DecrementLayerRefCount(layerID)
		return fmt.Errorf("failed to add working layer mapping: %w", err)
	}
	fmt.Fprintf(os.Stderr, "[COMMIT] Added working layer mapping: position %d -> %s\n", topPosition, layerID)

	// Create new empty working layer BEFORE removing old files
	diskLayersDir := filepath.Join(v.config.DataDir, "layers", v.name)
	newWorkingPath := v.imagePath + ".new"
	if err := v.qcow2Client.CreateWithBacking(ctx, cachedLayerPath, newWorkingPath, int(v.sizeGB)); err != nil {
		os.Remove(cachedLayerPath)
		v.stateDB.DecrementLayerRefCount(layerID)
		v.stateDB.DeleteDiskLayers(v.name)
		return fmt.Errorf("failed to create new working layer: %w", err)
	}

	// Replace old working image with new one
	if err := os.Rename(newWorkingPath, v.imagePath); err != nil {
		os.Remove(newWorkingPath)
		os.Remove(cachedLayerPath)
		v.stateDB.DecrementLayerRefCount(layerID)
		v.stateDB.DeleteDiskLayers(v.name)
		return fmt.Errorf("failed to replace working image: %w", err)
	}

	// Verify the backing file is set correctly
	verifiedBackingFile, err := v.qcow2Client.GetBackingFile(ctx, v.imagePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[COMMIT] WARNING: Failed to verify backing file: %v\n", err)
	} else if verifiedBackingFile != cachedLayerPath {
		fmt.Fprintf(os.Stderr, "[COMMIT] ERROR: Backing file mismatch! Expected: %s, Got: %s\n", cachedLayerPath, verifiedBackingFile)
		return fmt.Errorf("backing file verification failed: expected %s, got %s", cachedLayerPath, verifiedBackingFile)
	} else {
		fmt.Fprintf(os.Stderr, "[COMMIT] Verified backing file: %s\n", verifiedBackingFile)
	}

	// NOW remove old disk-specific layer directory (after we've replaced the working image)
	// The new working image doesn't depend on these files
	os.RemoveAll(diskLayersDir)

	// Recreate the layers directory for future use
	if err := os.MkdirAll(diskLayersDir, 0755); err != nil {
		// Non-fatal, just log it
		fmt.Fprintf(os.Stderr, "warning: failed to recreate layers directory: %v\n", err)
	}

	fmt.Fprintf(os.Stderr, "[COMMIT] Committed working image as layer %s (size: %d, checksum: %s)\n",
		layerID, fileInfo.Size(), checksum[:16])
	fmt.Fprintf(os.Stderr, "[COMMIT] New working image: %s -> backing: %s\n", v.imagePath, cachedLayerPath)

	return nil
}

// pushLayeredDisk uploads a disk with layer structure, sharing layers in S3
func (v *disk) pushLayeredDisk(ctx context.Context, state *DiskState, diskLayers []*DiskLayerMapping) error {
	fmt.Fprintf(os.Stderr, "[PUSH-LAYERED] Pushing disk '%s' with %d layers\n", v.name, len(diskLayers))

	s3Layers := make([]S3LayerInfo, 0, len(diskLayers))
	var baseLayerID, topLayerID string
	var finalChecksum string

	// Process each layer, skipping empty base layers
	s3LayerIndex := 0
	for _, diskLayer := range diskLayers {
		// Get the cached layer path
		layerCacheDir := filepath.Join(v.config.DataDir, "layer-cache")
		layerPath := filepath.Join(layerCacheDir, diskLayer.LayerID+".qcow2")

		// Check if layer file exists
		if _, err := os.Stat(layerPath); os.IsNotExist(err) {
			return fmt.Errorf("layer %s not found in cache", diskLayer.LayerID)
		}

		// Get layer state to get checksum and size
		layerState, err := v.stateDB.GetLayer(diskLayer.LayerID)
		if err != nil || layerState == nil {
			return fmt.Errorf("failed to get layer state for %s: %w", diskLayer.LayerID, err)
		}

		// Skip empty base layers (size < 1MB means it's likely an empty base with no user data)
		// These are created during disk creation but contain no user content
		const minLayerSize = 1024 * 1024 // 1 MB
		if layerState.Size < minLayerSize {
			fmt.Fprintf(os.Stderr, "[PUSH-LAYERED] Skipping empty base layer %s (size: %d bytes)\n",
				diskLayer.LayerID, layerState.Size)
			continue
		}

		// Ensure the layer has a checksum
		if layerState.Checksum == "" {
			fmt.Fprintf(os.Stderr, "[PUSH-LAYERED] Layer %s missing checksum, computing...\n", diskLayer.LayerID)
			checksum, err := v.qcow2Client.Checksum(ctx, layerPath)
			if err != nil {
				return fmt.Errorf("failed to compute checksum for layer %s: %w", diskLayer.LayerID, err)
			}
			layerState.Checksum = checksum
			if err := v.stateDB.SaveLayer(layerState); err != nil {
				return fmt.Errorf("failed to update layer checksum for %s: %w", diskLayer.LayerID, err)
			}
		}

		// Use checksum-based naming for S3 layers (shared across disks)
		s3LayerID := fmt.Sprintf("layer-%s", layerState.Checksum[:16])

		// Check if this layer already exists in S3 (by checksum)
		layerExists := false
		if err := v.s3Client.CheckLayerExists(ctx, s3LayerID); err == nil {
			fmt.Fprintf(os.Stderr, "[PUSH-LAYERED] Layer %s already in S3 (checksum: %s), skipping upload\n",
				diskLayer.LayerID, layerState.Checksum[:16])
			layerExists = true
		}

		// Upload the layer if it doesn't exist in S3
		if !layerExists {
			fmt.Fprintf(os.Stderr, "[PUSH-LAYERED] Uploading layer %s to S3 as %s...\n", diskLayer.LayerID, s3LayerID)

			// Upload the cached layer directly (no flattening needed)
			// The layer already has its backing file reference removed in commitWorkingImageAsLayer,
			// making it a standalone sparse delta layer that tar will compress efficiently
			if err := v.s3Client.UploadSharedLayer(ctx, s3LayerID, layerPath); err != nil {
				return fmt.Errorf("failed to upload shared layer %s: %w", diskLayer.LayerID, err)
			}

			fmt.Fprintf(os.Stderr, "[PUSH-LAYERED] Successfully uploaded layer %s (size: %d bytes)\n",
				diskLayer.LayerID, layerState.Size)
		}

		// Create S3 layer info referencing the shared layer
		parentID := ""
		if s3LayerIndex > 0 {
			parentID = s3Layers[s3LayerIndex-1].ID
		}

		layerInfo := S3LayerInfo{
			ID:          s3LayerID,
			ParentID:    parentID,
			Created:     layerState.CachedAt,
			Size:        layerState.Size,
			Checksum:    layerState.Checksum,
			Description: fmt.Sprintf("Shared layer (position %d)", s3LayerIndex),
		}
		s3Layers = append(s3Layers, layerInfo)

		if s3LayerIndex == 0 {
			baseLayerID = s3LayerID
		}
		topLayerID = s3LayerID
		finalChecksum = layerState.Checksum
		s3LayerIndex++
	}

	// Create metadata with shared layer references
	metadata := S3Metadata{
		Name:        v.name,
		SizeGB:      v.sizeGB,
		Created:     state.CreatedAt,
		Modified:    time.Now(),
		Checksum:    finalChecksum,
		Layers:      s3Layers,
		BaseLayerID: baseLayerID,
		TopLayerID:  topLayerID,
	}

	// Upload metadata
	if err := v.s3Client.UploadMetadata(ctx, v.name, metadata); err != nil {
		return fmt.Errorf("failed to upload metadata: %w", err)
	}

	// Update local state
	if err := v.stateDB.UpdateS3State(v.name, true, finalChecksum); err != nil {
		return fmt.Errorf("failed to update S3 state: %w", err)
	}

	fmt.Fprintf(os.Stderr, "[PUSH-LAYERED] Successfully pushed disk '%s' with %d shared layers\n", v.name, len(s3Layers))
	return nil
}

// pushIncrementalLayer creates and uploads an incremental snapshot
//
// STACKED BACKING FILE APPROACH:
// The working image is always a thin QCOW2 layer with a backing file (previous layer).
// When pushing:
// 1. Upload the current working layer (already contains only deltas)
// 2. Create a new empty working layer with the just-uploaded layer as backing
// This ensures true delta layers - each layer only contains what changed during that session.
func (v *disk) pushIncrementalLayer(ctx context.Context, state *DiskState, existingMetadata *S3Metadata) error {
	// Get file size of the current working layer
	fileInfo, err := os.Stat(v.imagePath)
	if err != nil {
		return fmt.Errorf("failed to get file size: %w", err)
	}
	actualSize := fileInfo.Size()

	// Check if working layer has meaningful changes
	// Empty working layers (just created by createNewWorkingLayer) are typically 200-500KB
	// If actualSize is very small AND matches the last recorded size, there were no writes
	const emptyLayerMaxSize = 600 * 1024 // 600 KB
	if actualSize < emptyLayerMaxSize {
		// Check if size matches the expected size of an empty layer
		// If the layer is this small, it's likely unchanged
		// We can safely skip pushing it
		return nil
	}

	// Calculate checksum of working image (the delta layer)
	checksum, err := v.qcow2Client.Checksum(ctx, v.imagePath)
	if err != nil {
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}

	// Determine if we should reuse the last layer or create a new one
	// Check the configured threshold
	thresholdBytes := v.config.S3.LayerSizeThresholdMB * 1024 * 1024
	layerID := fmt.Sprintf("layer-%d", time.Now().Unix())
	var parentID string
	reusingLayer := false

	if len(existingMetadata.Layers) > 0 {
		lastLayer := existingMetadata.Layers[len(existingMetadata.Layers)-1]

		// Check if the last layer is below the threshold and should be reused
		if lastLayer.Size < thresholdBytes {
			// Reuse the existing layer ID
			layerID = lastLayer.ID
			parentID = lastLayer.ParentID
			reusingLayer = true
			fmt.Fprintf(os.Stderr, "info: reusing layer %s (size: %d bytes < threshold: %d bytes)\n",
				layerID, lastLayer.Size, thresholdBytes)
		} else {
			// Create a new layer
			parentID = existingMetadata.TopLayerID
			fmt.Fprintf(os.Stderr, "info: creating new layer (last layer size: %d bytes >= threshold: %d bytes)\n",
				lastLayer.Size, thresholdBytes)
		}
	} else {
		parentID = ""
	}

	// CRITICAL FIX: Upload the layer WITHOUT backing file references
	// The current working layer has local backing file paths that won't exist after download
	// We need to create a copy of the layer with backing references removed

	// Create temp copy without backing file
	tempLayerPath := v.imagePath + ".upload-temp"
	defer os.Remove(tempLayerPath)

	// Copy the QCOW2 file
	if err := v.copyFile(v.imagePath, tempLayerPath); err != nil {
		return fmt.Errorf("failed to copy layer for upload: %w", err)
	}

	// Remove backing file reference from the copy (rebase -u -b "")
	// This makes it a standalone QCOW2 but keeps the layer data intact
	if err := v.qcow2Client.RebaseUnsafe(ctx, tempLayerPath, ""); err != nil {
		return fmt.Errorf("failed to remove backing reference: %w", err)
	}

	// Upload the layer without backing file
	if err := v.s3Client.UploadLayer(ctx, v.name, layerID, tempLayerPath); err != nil {
		return fmt.Errorf("failed to upload layer: %w", err)
	}

	// Create or update layer info
	newLayer := S3LayerInfo{
		ID:       layerID,
		ParentID: parentID,
		Created:  time.Now(),
		Size:     actualSize, // Size of the delta layer
		Checksum: checksum,
	}

	if reusingLayer {
		newLayer.Description = fmt.Sprintf("Incremental snapshot (reused layer, delta-only) at %s", time.Now().Format(time.RFC3339))
		// Update the existing layer in the metadata
		for i := range existingMetadata.Layers {
			if existingMetadata.Layers[i].ID == layerID {
				existingMetadata.Layers[i] = newLayer
				break
			}
		}
	} else {
		newLayer.Description = fmt.Sprintf("Incremental snapshot (new layer, delta-only) at %s", time.Now().Format(time.RFC3339))
		// Append new layer to metadata
		existingMetadata.Layers = append(existingMetadata.Layers, newLayer)
	}

	existingMetadata.TopLayerID = layerID
	existingMetadata.Modified = time.Now()
	existingMetadata.Checksum = checksum // Checksum of the full working image

	// Upload updated metadata
	if err := v.s3Client.UploadMetadata(ctx, v.name, *existingMetadata); err != nil {
		return fmt.Errorf("failed to upload metadata: %w", err)
	}

	// Update local state
	if err := v.stateDB.UpdateS3State(v.name, true, checksum); err != nil {
		return fmt.Errorf("failed to update S3 state: %w", err)
	}

	// Decide whether to create a new working layer based on current layer size
	// If we're reusing and the layer is still below threshold, keep the current working layer
	// If the layer has grown to/above the threshold, create a new layer for next push
	if reusingLayer && actualSize < thresholdBytes {
		// Keep the current working layer for continued reuse
		fmt.Fprintf(os.Stderr, "info: keeping current working layer for reuse (size: %d bytes < threshold: %d bytes)\n",
			actualSize, thresholdBytes)
	} else {
		// Create a new empty working layer with the just-uploaded layer as backing
		// This is critical for the stacked approach - future writes go to the new layer
		if err := v.createNewWorkingLayer(ctx, layerID); err != nil {
			return fmt.Errorf("failed to create new working layer: %w", err)
		}
	}

	return nil
}

// createNewWorkingLayer creates a new empty QCOW2 with the original backing chain preserved
// Strategy:
// 1. Get the backing file of the current working image (preserves the layer chain)
// 2. Create a new empty layer with that backing file
// This preserves the original layer structure instead of flattening
func (v *disk) createNewWorkingLayer(ctx context.Context, backingLayerID string) error {
	layersDir := filepath.Join(v.config.DataDir, "layers", v.name)
	if err := os.MkdirAll(layersDir, 0755); err != nil {
		return fmt.Errorf("failed to create layers directory: %w", err)
	}

	// Get the current backing file of the working image
	// This preserves the entire layer chain (base -> delta-1 -> delta-2 -> ... -> top)
	backingFile, err := v.qcow2Client.GetBackingFile(ctx, v.imagePath)
	if err != nil {
		return fmt.Errorf("failed to get backing file: %w", err)
	}

	// If there's no backing file, something is wrong (should always have one after pull)
	if backingFile == "" {
		return fmt.Errorf("working image has no backing file - cannot create new working layer")
	}

	// Create a new EMPTY QCOW2 with the preserved backing chain
	tempNewLayer := v.imagePath + ".new"
	if err := v.qcow2Client.CreateWithBacking(ctx, backingFile, tempNewLayer, int(v.sizeGB)); err != nil {
		return fmt.Errorf("failed to create new working layer: %w", err)
	}

	// Replace the working image with the new empty layer
	if err := os.Rename(tempNewLayer, v.imagePath); err != nil {
		return fmt.Errorf("failed to replace working image: %w", err)
	}

	return nil
}

// copyFile copies a file from src to dst
func (v *disk) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

func (v *disk) Info() DiskInfo {
	v.mu.Lock()
	defer v.mu.Unlock()

	// Get state from database
	state, err := v.stateDB.GetDisk(v.name)
	if err != nil {
		// Return minimal info if database query fails
		return DiskInfo{
			Name:      v.name,
			SizeGB:    v.sizeGB,
			IsMounted: v.isMounted,
		}
	}

	// Get actual size from QCOW2 image
	actualSizeBytes, err := v.qcow2Client.GetActualSize(context.Background(), v.imagePath)
	actualSizeGB := actualSizeBytes / (1024 * 1024 * 1024)
	if err != nil {
		actualSizeGB = 0
	}

	return DiskInfo{
		Name:         v.name,
		SizeGB:       v.sizeGB,
		ActualSizeGB: actualSizeGB,
		Created:      state.CreatedAt,
		Modified:     state.ModifiedAt,
		IsMounted:    v.isMounted,
		InS3:         state.InS3,
		Checksum:     state.Checksum,
	}
}

func (v *disk) Resize(ctx context.Context, newSizeGB int) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.isMounted {
		return fmt.Errorf("cannot resize mounted disk: unmount it first")
	}

	if newSizeGB <= int(v.sizeGB) {
		return fmt.Errorf("new size must be larger than current size")
	}

	// Resize the QCOW2 image
	if err := v.qcow2Client.Resize(ctx, v.imagePath, newSizeGB); err != nil {
		return fmt.Errorf("failed to resize QCOW2 image: %w", err)
	}

	v.sizeGB = int64(newSizeGB)

	// Update state
	state, err := v.stateDB.GetDisk(v.name)
	if err != nil {
		return fmt.Errorf("failed to get disk state: %w", err)
	}

	state.SizeGB = int64(newSizeGB)
	state.ModifiedAt = time.Now()

	if err := v.stateDB.SaveDisk(state); err != nil {
		return fmt.Errorf("failed to update disk state: %w", err)
	}

	return nil
}

func (v *disk) Close() error {
	v.mu.Lock()
	defer v.mu.Unlock()

	// Unmount if mounted
	if v.isMounted {
		ctx := context.Background()
		if err := v.qcow2Client.Unmount(ctx, v.name); err != nil {
			return err
		}
		if err := v.qcow2Client.Disconnect(ctx, v.name); err != nil {
			return err
		}
		v.isMounted = false
	}

	return nil
}
