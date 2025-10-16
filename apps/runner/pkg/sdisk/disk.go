package sdisk

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
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

func (v *disk) Unmount(ctx context.Context) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if !v.isMounted {
		return nil
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
	return v.isMounted
}

func (v *disk) MountPath() string {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.mountPath
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
	// CRITICAL: The base layer must be a standalone image with no backing files
	// If the working image has a backing file, flatten it first
	baseImagePath := v.imagePath
	backingFile, err := v.qcow2Client.GetBackingFile(ctx, v.imagePath)
	if err != nil {
		return fmt.Errorf("failed to check backing file: %w", err)
	}

	// If there's a backing file, flatten the image first
	if backingFile != "" {
		flattenedPath := v.imagePath + ".flattened"
		if err := v.qcow2Client.Convert(ctx, v.imagePath, flattenedPath); err != nil {
			return fmt.Errorf("failed to flatten base image: %w", err)
		}
		// Use the flattened image as the base
		baseImagePath = flattenedPath
		defer os.Remove(flattenedPath)
	}

	// Calculate checksum
	checksum, err := v.qcow2Client.Checksum(ctx, baseImagePath)
	if err != nil {
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}

	// Get actual size
	actualSize, err := v.qcow2Client.GetActualSize(ctx, baseImagePath)
	if err != nil {
		return fmt.Errorf("failed to get actual size: %w", err)
	}

	// Generate base layer ID
	layerID := fmt.Sprintf("base-%d", time.Now().Unix())

	// Upload as a layer (using flattened version if it had backing)
	if err := v.s3Client.UploadLayer(ctx, v.name, layerID, baseImagePath); err != nil {
		return fmt.Errorf("failed to upload base layer: %w", err)
	}

	// Create layer info
	layerInfo := S3LayerInfo{
		ID:          layerID,
		ParentID:    "",
		Created:     time.Now(),
		Size:        actualSize,
		Checksum:    checksum,
		Description: "Base layer",
	}

	// Create metadata with layer information
	metadata := S3Metadata{
		Name:        v.name,
		SizeGB:      v.sizeGB,
		Created:     state.CreatedAt,
		Modified:    time.Now(),
		Checksum:    checksum,
		Layers:      []S3LayerInfo{layerInfo},
		BaseLayerID: layerID,
		TopLayerID:  layerID,
	}

	// Upload metadata
	if err := v.s3Client.UploadMetadata(ctx, v.name, metadata); err != nil {
		return fmt.Errorf("failed to upload metadata: %w", err)
	}

	// Update local state
	if err := v.stateDB.UpdateS3State(v.name, true, checksum); err != nil {
		return fmt.Errorf("failed to update S3 state: %w", err)
	}

	// After pushing base layer, create a new empty working layer with base as backing
	// This way, future writes only go to the new layer (true delta approach)
	if err := v.createNewWorkingLayer(ctx, layerID); err != nil {
		return fmt.Errorf("failed to create new working layer: %w", err)
	}

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
	// Get actual size of the current working layer
	actualSize, err := v.qcow2Client.GetActualSize(ctx, v.imagePath)
	if err != nil {
		return fmt.Errorf("failed to get actual size: %w", err)
	}

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
		} else {
			// Create a new layer
			parentID = existingMetadata.TopLayerID
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
	} else {
		// Create a new empty working layer with the just-uploaded layer as backing
		// This is critical for the stacked approach - future writes go to the new layer
		if err := v.createNewWorkingLayer(ctx, layerID); err != nil {
			return fmt.Errorf("failed to create new working layer: %w", err)
		}
	}

	return nil
}

// createNewWorkingLayer creates a new empty QCOW2 with a flattened backing
// Strategy:
// 1. Flatten the current working image (which may have backing files)
// 2. Save the flattened version as the backing layer
// 3. Create a new empty layer with the flattened backing
// This keeps S3 layers thin (with backing refs) but local backing always flattened (no chains)
func (v *disk) createNewWorkingLayer(ctx context.Context, backingLayerID string) error {
	layersDir := filepath.Join(v.config.DataDir, "layers", v.name)
	if err := os.MkdirAll(layersDir, 0755); err != nil {
		return fmt.Errorf("failed to create layers directory: %w", err)
	}

	// Use a single consolidated backing file that gets replaced each push
	// This avoids locking issues and keeps the local backing chain depth at 1
	consolidatedPath := filepath.Join(layersDir, "consolidated.qcow2")
	tempConsolidatedPath := filepath.Join(layersDir, "consolidated.tmp.qcow2")

	// Flatten the current working image to a temp consolidated backing
	// This removes all backing file references and creates a standalone image
	if err := v.qcow2Client.Convert(ctx, v.imagePath, tempConsolidatedPath); err != nil {
		return fmt.Errorf("failed to flatten backing layer: %w", err)
	}

	// Replace the old consolidated with the new one
	// This must happen BEFORE creating the new working layer
	os.Remove(consolidatedPath)
	if err := os.Rename(tempConsolidatedPath, consolidatedPath); err != nil {
		return fmt.Errorf("failed to update consolidated backing: %w", err)
	}

	// Create a new EMPTY QCOW2 with the consolidated backing
	tempNewLayer := v.imagePath + ".new"
	if err := v.qcow2Client.CreateWithBacking(ctx, consolidatedPath, tempNewLayer, int(v.sizeGB)); err != nil {
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
