package sdisk

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// manager implements the DiskManager interface
type manager struct {
	config      Config
	qcow2Client *QCowClient
	s3Client    *S3Client
	stateDB     *DB
	disks       map[string]*disk
	pool        *DiskPool // Disk pool for managing mounted disks
	mu          sync.RWMutex
}

// NewManager creates a new disk manager
func NewManager(config Config) (DiskManager, error) {
	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Create data directories
	if err := os.MkdirAll(filepath.Join(config.DataDir, "disks"), 0755); err != nil {
		return nil, fmt.Errorf("failed to create disks directory: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(config.DataDir, "mounts"), 0755); err != nil {
		return nil, fmt.Errorf("failed to create mounts directory: %w", err)
	}

	// Initialize QCOW2 client
	qcow2Client, err := NewQCowClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create QCOW2 client: %w", err)
	}

	// Initialize S3 client (optional)
	var s3Client *S3Client
	if config.S3.Bucket != "" && config.S3.Region != "" {
		s3Client, err = NewS3Client(context.Background(), S3Config{
			Bucket:          config.S3.Bucket,
			Region:          config.S3.Region,
			AccessKeyID:     config.S3.AccessKeyID,
			SecretAccessKey: config.S3.SecretAccessKey,
			Endpoint:        config.S3.Endpoint,
			UsePathStyle:    config.S3.UsePathStyle,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create S3 client: %w", err)
		}
	}

	// Initialize state database
	dbPath := filepath.Join(config.DataDir, "state.db")
	stateDB, err := NewDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create state database: %w", err)
	}

	mgr := &manager{
		config:      config,
		qcow2Client: qcow2Client,
		s3Client:    s3Client,
		stateDB:     stateDB,
		disks:       make(map[string]*disk),
	}

	// Initialize disk pool if enabled
	if config.Pool.Enabled {
		mgr.pool = NewDiskPool(config.Pool.MaxMounted)
	}

	// Load existing disks
	if err := mgr.loadDisks(); err != nil {
		return nil, fmt.Errorf("failed to load disks: %w", err)
	}

	return mgr, nil
}

func (m *manager) Create(ctx context.Context, name string, sizeGB int) (Disk, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if disk already exists
	if _, exists := m.disks[name]; exists {
		return nil, ErrDiskExists
	}

	// Check state database
	existing, err := m.stateDB.GetDisk(name)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing disk: %w", err)
	}
	if existing != nil {
		return nil, ErrDiskExists
	}

	if sizeGB <= 0 {
		return nil, ErrInvalidSize
	}

	// Create QCOW2 image
	imagePath := filepath.Join(m.config.DataDir, "disks", name+".qcow2")

	createConfig := QCowCreateConfig{
		Path:          imagePath,
		SizeGB:        sizeGB,
		Compression:   m.config.QCOW2.Compression,
		ClusterSize:   m.config.QCOW2.ClusterSize,
		LazyRefcounts: m.config.QCOW2.LazyRefcounts,
		Preallocation: m.config.QCOW2.Preallocation,
	}

	if err := m.qcow2Client.Create(ctx, createConfig); err != nil {
		return nil, fmt.Errorf("failed to create QCOW2 image: %w", err)
	}

	// Save disk state
	now := time.Now()
	state := &DiskState{
		Name:       name,
		SizeGB:     int64(sizeGB),
		CreatedAt:  now,
		ModifiedAt: now,
		IsMounted:  false,
		MountPath:  "",
		InS3:       false,
		Checksum:   "",
	}

	if err := m.stateDB.SaveDisk(state); err != nil {
		// Cleanup created image on failure
		os.Remove(imagePath)
		return nil, fmt.Errorf("failed to save disk state: %w", err)
	}

	// Create disk object
	vol := &disk{
		name:        name,
		sizeGB:      int64(sizeGB),
		imagePath:   imagePath,
		qcow2Client: m.qcow2Client,
		s3Client:    m.s3Client,
		stateDB:     m.stateDB,
		config:      m.config,
		pool:        m.pool,
		isMounted:   false,
		mountPath:   "",
	}

	m.disks[name] = vol

	return vol, nil
}

func (m *manager) Open(ctx context.Context, name string) (Disk, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if already loaded
	if vol, exists := m.disks[name]; exists {
		return vol, nil
	}

	// Get from state database
	state, err := m.stateDB.GetDisk(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get disk state: %w", err)
	}
	if state == nil {
		return nil, ErrDiskNotFound
	}

	// Check if image file exists
	imagePath := filepath.Join(m.config.DataDir, "disks", name+".qcow2")
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return nil, ErrDiskNotFound
	}

	// Create disk object
	vol := &disk{
		name:        name,
		sizeGB:      state.SizeGB,
		imagePath:   imagePath,
		qcow2Client: m.qcow2Client,
		s3Client:    m.s3Client,
		stateDB:     m.stateDB,
		config:      m.config,
		pool:        m.pool,
		isMounted:   state.IsMounted,
		mountPath:   state.MountPath,
	}

	m.disks[name] = vol

	return vol, nil
}

func (m *manager) Pull(ctx context.Context, name string) (Disk, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.s3Client == nil {
		return nil, ErrS3NotConfigured
	}

	// Check if disk already exists locally
	if _, exists := m.disks[name]; exists {
		return nil, ErrDiskExists
	}

	// Check state database
	existing, err := m.stateDB.GetDisk(name)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing disk: %w", err)
	}
	if existing != nil {
		return nil, ErrDiskExists
	}

	// Download metadata first to check if it's a layered disk
	metadata, err := m.s3Client.DownloadMetadata(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to download metadata: %w", err)
	}

	// Final image path
	imagePath := filepath.Join(m.config.DataDir, "disks", name+".qcow2")

	// Check if this is a layered disk
	if len(metadata.Layers) > 0 {
		// Download and reconstruct layered disk
		if err := m.pullLayeredDisk(ctx, name, metadata, imagePath); err != nil {
			return nil, err
		}

		// CRITICAL: After restoring, set up backing file structure for future deltas
		// The restored image is standalone, but we need it as a delta layer for future pushes
		if err := m.setupRestoredDiskForDeltas(ctx, name, metadata, imagePath, metadata.SizeGB); err != nil {
			return nil, fmt.Errorf("failed to setup delta structure: %w", err)
		}
	} else {
		// Legacy non-layered disk - download directly
		if err := m.s3Client.DownloadDisk(ctx, name, imagePath); err != nil {
			return nil, fmt.Errorf("failed to download disk: %w", err)
		}
	}

	// Save disk state
	state := &DiskState{
		Name:       name,
		SizeGB:     metadata.SizeGB,
		CreatedAt:  metadata.Created,
		ModifiedAt: metadata.Modified,
		IsMounted:  false,
		MountPath:  "",
		InS3:       true,
		Checksum:   metadata.Checksum,
	}

	if err := m.stateDB.SaveDisk(state); err != nil {
		// Cleanup
		os.Remove(imagePath)
		return nil, fmt.Errorf("failed to save disk state: %w", err)
	}

	// Create disk object
	vol := &disk{
		name:        name,
		sizeGB:      metadata.SizeGB,
		imagePath:   imagePath,
		qcow2Client: m.qcow2Client,
		s3Client:    m.s3Client,
		stateDB:     m.stateDB,
		config:      m.config,
		pool:        m.pool,
		isMounted:   false,
		mountPath:   "",
	}

	m.disks[name] = vol

	return vol, nil
}

// pullLayeredDisk downloads all layers and reconstructs the disk
// Each layer in S3 is a standalone delta (no backing file references)
// We need to apply them sequentially: base + delta1 + delta2 + ... = final
func (m *manager) pullLayeredDisk(ctx context.Context, name string, metadata *S3Metadata, finalPath string) error {
	if len(metadata.Layers) == 0 {
		return fmt.Errorf("no layers found in metadata")
	}

	// Create layers directory for this disk
	layersDir := filepath.Join(m.config.DataDir, "layers", name)
	if err := os.MkdirAll(layersDir, 0755); err != nil {
		return fmt.Errorf("failed to create layers directory: %w", err)
	}
	defer m.cleanupLayers(layersDir)

	// Download the base layer first
	baseLayer := metadata.Layers[0]
	baseLayerPath := filepath.Join(layersDir, baseLayer.ID+".qcow2")
	if err := m.s3Client.DownloadLayer(ctx, name, baseLayer.ID, baseLayerPath); err != nil {
		return fmt.Errorf("failed to download base layer %s: %w", baseLayer.ID, err)
	}

	// Start with the base layer as our current image
	currentImage := baseLayerPath

	// Apply each delta layer on top
	for i := 1; i < len(metadata.Layers); i++ {
		layer := metadata.Layers[i]
		deltaLayerPath := filepath.Join(layersDir, layer.ID+".qcow2")

		// Download the delta layer
		if err := m.s3Client.DownloadLayer(ctx, name, layer.ID, deltaLayerPath); err != nil {
			return fmt.Errorf("failed to download layer %s: %w", layer.ID, err)
		}

		// Create a new merged image by applying the delta on top of current
		mergedPath := filepath.Join(layersDir, fmt.Sprintf("merged-%d.qcow2", i))

		// Apply the delta: currentImage is the backing, deltaLayerPath contains the changes
		// We use qemu-img rebase in safe mode to properly merge them
		if err := m.applyDeltaLayer(ctx, currentImage, deltaLayerPath, mergedPath); err != nil {
			return fmt.Errorf("failed to apply delta layer %s: %w", layer.ID, err)
		}

		// The merged image becomes our new current image for the next iteration
		currentImage = mergedPath
	}

	// Convert the final merged image to the destination path
	// This ensures it's a clean standalone QCOW2
	if err := m.qcow2Client.Convert(ctx, currentImage, finalPath); err != nil {
		return fmt.Errorf("failed to convert final image: %w", err)
	}

	return nil
}

// applyDeltaLayer applies a delta layer on top of a base image
// This creates a new merged image that combines base + delta
func (m *manager) applyDeltaLayer(ctx context.Context, baseImage, deltaLayer, outputPath string) error {
	// First, rebase the delta layer to use the base image as backing
	// This is done on a temporary copy to avoid modifying the original delta
	tempDelta := deltaLayer + ".rebased"
	defer os.Remove(tempDelta)

	// Copy the delta layer
	if err := m.copyFile(deltaLayer, tempDelta); err != nil {
		return fmt.Errorf("failed to copy delta layer: %w", err)
	}

	// Rebase the delta to use baseImage as its backing file
	if err := m.qcow2Client.RebaseUnsafe(ctx, tempDelta, baseImage); err != nil {
		return fmt.Errorf("failed to rebase delta layer: %w", err)
	}

	// Now convert the rebased delta (which has baseImage as backing) to a standalone image
	// This effectively merges base + delta into a single image
	if err := m.qcow2Client.Convert(ctx, tempDelta, outputPath); err != nil {
		return fmt.Errorf("failed to merge layers: %w", err)
	}

	return nil
}

// cleanupLayers removes downloaded layer files
func (m *manager) cleanupLayers(layersDir string) {
	os.RemoveAll(layersDir)
}

// copyFile copies a file from src to dst
func (m *manager) copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := dstFile.ReadFrom(srcFile); err != nil {
		return err
	}

	return dstFile.Sync()
}

func (m *manager) List(ctx context.Context) ([]DiskInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	states, err := m.stateDB.ListDisks()
	if err != nil {
		return nil, fmt.Errorf("failed to list disks: %w", err)
	}

	var infos []DiskInfo
	for _, state := range states {
		// Get actual size from QCOW2 image
		imagePath := filepath.Join(m.config.DataDir, "disks", state.Name+".qcow2")
		actualSizeBytes, err := m.qcow2Client.GetActualSize(ctx, imagePath)
		actualSizeGB := actualSizeBytes / (1024 * 1024 * 1024)
		if err != nil {
			actualSizeGB = 0
		}

		infos = append(infos, DiskInfo{
			Name:         state.Name,
			SizeGB:       state.SizeGB,
			ActualSizeGB: actualSizeGB,
			Created:      state.CreatedAt,
			Modified:     state.ModifiedAt,
			IsMounted:    state.IsMounted,
			InS3:         state.InS3,
			Checksum:     state.Checksum,
		})
	}

	return infos, nil
}

func (m *manager) Delete(ctx context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get disk
	vol, exists := m.disks[name]
	if exists {
		// Check if mounted
		if vol.isMounted {
			return ErrDiskInUse
		}

		// Close the disk
		if err := vol.Close(); err != nil {
			return fmt.Errorf("failed to close disk: %w", err)
		}
	}

	// Delete QCOW2 image
	imagePath := filepath.Join(m.config.DataDir, "disks", name+".qcow2")
	if err := os.Remove(imagePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete disk image: %w", err)
	}

	// Delete from state database
	if err := m.stateDB.DeleteDisk(name); err != nil {
		return fmt.Errorf("failed to delete disk state: %w", err)
	}

	// Remove from manager
	delete(m.disks, name)

	return nil
}

func (m *manager) PoolStats() *PoolStats {
	if m.pool == nil {
		return nil
	}
	stats := m.pool.Stats()
	return &stats
}

func (m *manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Close the pool first (will unmount all pooled disks)
	if m.pool != nil {
		if err := m.pool.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to close disk pool: %v\n", err)
		}
	}

	// Close all disks
	for name, vol := range m.disks {
		if err := vol.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to close disk %s: %v\n", name, err)
		}
	}

	// Close QCOW2 client
	ctx := context.Background()
	if err := m.qcow2Client.Close(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to close QCOW2 client: %v\n", err)
	}

	// Close state database
	if err := m.stateDB.Close(); err != nil {
		return fmt.Errorf("failed to close state database: %w", err)
	}

	return nil
}

func (m *manager) loadDisks() error {
	states, err := m.stateDB.ListDisks()
	if err != nil {
		return err
	}

	for _, state := range states {
		imagePath := filepath.Join(m.config.DataDir, "disks", state.Name+".qcow2")

		// Check if image exists
		if _, err := os.Stat(imagePath); os.IsNotExist(err) {
			// Image missing, skip
			continue
		}

		vol := &disk{
			name:        state.Name,
			sizeGB:      state.SizeGB,
			imagePath:   imagePath,
			qcow2Client: m.qcow2Client,
			s3Client:    m.s3Client,
			stateDB:     m.stateDB,
			config:      m.config,
			pool:        m.pool,
			isMounted:   state.IsMounted,
			mountPath:   state.MountPath,
		}

		m.disks[state.Name] = vol
	}

	return nil
}

// pullLayersToConsolidated downloads and merges a subset of layers into a consolidated image
func (m *manager) pullLayersToConsolidated(ctx context.Context, name string, layers []S3LayerInfo, outputPath string) error {
	if len(layers) == 0 {
		return fmt.Errorf("no layers to consolidate")
	}

	// Create temporary directory for layer downloads
	tempDir := filepath.Join(m.config.DataDir, "layers", name, "temp")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Download the base layer first
	baseLayer := layers[0]
	baseLayerPath := filepath.Join(tempDir, baseLayer.ID+".qcow2")
	if err := m.s3Client.DownloadLayer(ctx, name, baseLayer.ID, baseLayerPath); err != nil {
		return fmt.Errorf("failed to download base layer %s: %w", baseLayer.ID, err)
	}

	// Start with the base layer as our current image
	currentImage := baseLayerPath

	// Apply each delta layer on top
	for i := 1; i < len(layers); i++ {
		layer := layers[i]
		deltaLayerPath := filepath.Join(tempDir, layer.ID+".qcow2")

		// Download the delta layer
		if err := m.s3Client.DownloadLayer(ctx, name, layer.ID, deltaLayerPath); err != nil {
			return fmt.Errorf("failed to download layer %s: %w", layer.ID, err)
		}

		// Create a new merged image by applying the delta on top of current
		mergedPath := filepath.Join(tempDir, fmt.Sprintf("merged-%d.qcow2", i))

		// Apply the delta
		if err := m.applyDeltaLayer(ctx, currentImage, deltaLayerPath, mergedPath); err != nil {
			return fmt.Errorf("failed to apply delta layer %s: %w", layer.ID, err)
		}

		// The merged image becomes our new current image for the next iteration
		currentImage = mergedPath
	}

	// Convert the final merged image to the output path
	if err := m.qcow2Client.Convert(ctx, currentImage, outputPath); err != nil {
		return fmt.Errorf("failed to convert consolidated image: %w", err)
	}

	return nil
}

// setupRestoredDiskForDeltas sets up the backing file structure for a restored disk
// After pulling from S3, the disk is a standalone image. This method:
// 1. Checks the size of the last layer in S3
// 2. If last layer < threshold: Reuse it for writes (don't create new layer)
// 3. If last layer >= threshold: Create new empty layer for writes
// This ensures future pushes create delta layers instead of uploading the full image
func (m *manager) setupRestoredDiskForDeltas(ctx context.Context, name string, metadata *S3Metadata, imagePath string, sizeGB int64) error {
	layersDir := filepath.Join(m.config.DataDir, "layers", name)
	if err := os.MkdirAll(layersDir, 0755); err != nil {
		return fmt.Errorf("failed to create layers directory: %w", err)
	}

	// Check if we should reuse the last layer based on its size
	thresholdBytes := m.config.S3.LayerSizeThresholdMB * 1024 * 1024
	shouldReuseLastLayer := false

	if len(metadata.Layers) > 0 {
		lastLayer := metadata.Layers[len(metadata.Layers)-1]
		if lastLayer.Size < thresholdBytes {
			shouldReuseLastLayer = true
			fmt.Fprintf(os.Stderr, "info: last layer size (%d bytes) is below threshold (%d bytes), will reuse for writes\n",
				lastLayer.Size, thresholdBytes)
		} else {
			fmt.Fprintf(os.Stderr, "info: last layer size (%d bytes) is >= threshold (%d bytes), will create new layer for writes\n",
				lastLayer.Size, thresholdBytes)
		}
	}

	if shouldReuseLastLayer {
		// Reuse the last layer: Set up backing structure to allow the last layer to grow
		// Strategy:
		// 1. Download all layers except the last one and create a consolidated backing
		// 2. Download the last layer separately
		// 3. Rebase the last layer to use the consolidated backing
		// This way, new writes will go to the existing last layer, allowing it to grow

		if len(metadata.Layers) == 1 {
			// Only one layer (base) - just keep it as standalone for writing
			// Save it as consolidated backing for consistency
			consolidatedPath := filepath.Join(layersDir, "consolidated.qcow2")
			if err := m.copyFile(imagePath, consolidatedPath); err != nil {
				return fmt.Errorf("failed to save consolidated backing: %w", err)
			}
			return nil
		}

		// Multiple layers - need to reconstruct with last layer separate
		// Download and merge all layers except the last one
		layersToMerge := metadata.Layers[:len(metadata.Layers)-1]
		lastLayer := metadata.Layers[len(metadata.Layers)-1]

		// Create consolidated backing from all layers except the last
		consolidatedPath := filepath.Join(layersDir, "consolidated.qcow2")
		if err := m.pullLayersToConsolidated(ctx, name, layersToMerge, consolidatedPath); err != nil {
			return fmt.Errorf("failed to create consolidated backing: %w", err)
		}

		// Download the last layer
		lastLayerPath := filepath.Join(layersDir, lastLayer.ID+".qcow2")
		if err := m.s3Client.DownloadLayer(ctx, name, lastLayer.ID, lastLayerPath); err != nil {
			return fmt.Errorf("failed to download last layer: %w", err)
		}

		// Rebase the last layer to use consolidated as backing
		if err := m.qcow2Client.RebaseUnsafe(ctx, lastLayerPath, consolidatedPath); err != nil {
			return fmt.Errorf("failed to rebase last layer: %w", err)
		}

		// Copy the rebased last layer as the working image
		if err := m.copyFile(lastLayerPath, imagePath); err != nil {
			return fmt.Errorf("failed to set up working image: %w", err)
		}

		// Clean up temp last layer file
		os.Remove(lastLayerPath)

		return nil
	}

	// Last layer is >= threshold: create new empty layer for writes (existing behavior)

	// Save the restored standalone image as consolidated backing
	consolidatedPath := filepath.Join(layersDir, "consolidated.qcow2")
	if err := m.copyFile(imagePath, consolidatedPath); err != nil {
		return fmt.Errorf("failed to save consolidated backing: %w", err)
	}

	// Create new empty working layer with consolidated backing
	tempNewLayer := imagePath + ".new"
	if err := m.qcow2Client.CreateWithBacking(ctx, consolidatedPath, tempNewLayer, int(sizeGB)); err != nil {
		return fmt.Errorf("failed to create working layer: %w", err)
	}

	// Replace the standalone image with the new delta layer
	if err := os.Rename(tempNewLayer, imagePath); err != nil {
		return fmt.Errorf("failed to replace image: %w", err)
	}

	return nil
}
