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
	layerCache  *LayerCache // Global layer cache
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

	// Initialize layer cache
	layerCacheDir := filepath.Join(config.DataDir, "layer-cache")
	if err := os.MkdirAll(layerCacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create layer cache directory: %w", err)
	}
	layerCache, err := NewLayerCache(layerCacheDir, s3Client, stateDB)
	if err != nil {
		return nil, fmt.Errorf("failed to create layer cache: %w", err)
	}
	mgr.layerCache = layerCache

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

	// Create QCOW2 image using layer stacking approach
	imagePath := filepath.Join(m.config.DataDir, "disks", name+".qcow2")

	// Create a base layer in the global cache
	baseLayerID := fmt.Sprintf("base-%s-%d", name, time.Now().Unix())
	baseLayerPath := m.layerCache.GetLayerPath(baseLayerID)

	// Create the base layer directly in cache
	createConfig := QCowCreateConfig{
		Path:          baseLayerPath,
		SizeGB:        sizeGB,
		Compression:   m.config.QCOW2.Compression,
		ClusterSize:   m.config.QCOW2.ClusterSize,
		LazyRefcounts: m.config.QCOW2.LazyRefcounts,
		Preallocation: m.config.QCOW2.Preallocation,
	}

	if err := m.qcow2Client.Create(ctx, createConfig); err != nil {
		return nil, fmt.Errorf("failed to create base layer: %w", err)
	}

	// Save base layer to database
	baseLayerState := &LayerState{
		ID:       baseLayerID,
		Checksum: "", // Will be calculated on first push
		Size:     0,  // Will be updated on first push
		CachedAt: time.Now(),
		RefCount: 1, // First reference
	}

	if err := m.stateDB.SaveLayer(baseLayerState); err != nil {
		// Cleanup base layer file
		os.Remove(baseLayerPath)
		return nil, fmt.Errorf("failed to save base layer state: %w", err)
	}

	// Track disk-layer mapping
	if err := m.stateDB.AddDiskLayerMapping(name, baseLayerID, 0); err != nil {
		// Cleanup base layer file and database entry
		os.Remove(baseLayerPath)
		m.stateDB.DecrementLayerRefCount(baseLayerID)
		return nil, fmt.Errorf("failed to add disk-layer mapping: %w", err)
	}

	// Create disk-specific working layer that references the base
	diskLayersDir := filepath.Join(m.config.DataDir, "layers", name)
	if err := os.MkdirAll(diskLayersDir, 0755); err != nil {
		// Cleanup
		os.Remove(baseLayerPath)
		m.stateDB.DecrementLayerRefCount(baseLayerID)
		m.stateDB.DeleteDiskLayers(name)
		return nil, fmt.Errorf("failed to create disk layers directory: %w", err)
	}

	// Create working layer with base as backing
	workingLayerPath := filepath.Join(diskLayersDir, "working.qcow2")
	if err := m.qcow2Client.CreateWithBacking(ctx, baseLayerPath, workingLayerPath, sizeGB); err != nil {
		// Cleanup
		os.RemoveAll(diskLayersDir)
		os.Remove(baseLayerPath)
		m.stateDB.DecrementLayerRefCount(baseLayerID)
		m.stateDB.DeleteDiskLayers(name)
		return nil, fmt.Errorf("failed to create working layer: %w", err)
	}

	// Copy working layer to final image path
	if err := m.copyFile(workingLayerPath, imagePath); err != nil {
		// Cleanup
		os.RemoveAll(diskLayersDir)
		os.Remove(baseLayerPath)
		m.stateDB.DecrementLayerRefCount(baseLayerID)
		m.stateDB.DeleteDiskLayers(name)
		return nil, fmt.Errorf("failed to create working image: %w", err)
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

// pullLayeredDisk downloads all layers and creates a stacked backing file chain
// Each layer in S3 is a standalone delta (no backing file references)
// We create a backing file chain: base -> delta1 -> delta2 -> ... -> final
func (m *manager) pullLayeredDisk(ctx context.Context, name string, metadata *S3Metadata, finalPath string) error {
	if len(metadata.Layers) == 0 {
		return fmt.Errorf("no layers found in metadata")
	}

	// Create layers directory for disk-specific working image
	diskLayersDir := filepath.Join(m.config.DataDir, "layers", name)
	if err := os.MkdirAll(diskLayersDir, 0755); err != nil {
		return fmt.Errorf("failed to create disk layers directory: %w", err)
	}

	// Download all layers to global cache
	layerPaths := make([]string, len(metadata.Layers))
	for i, layer := range metadata.Layers {
		layerPath, err := m.layerCache.GetOrDownload(ctx, name, layer.ID, layer)
		if err != nil {
			return fmt.Errorf("failed to get layer %s: %w", layer.ID, err)
		}
		layerPaths[i] = layerPath

		// Track disk-layer mapping
		if err := m.stateDB.AddDiskLayerMapping(name, layer.ID, i); err != nil {
			return fmt.Errorf("failed to add disk-layer mapping: %w", err)
		}
	}

	// Build backing file chain by rebasing layers
	// We need to create local copies that reference the cached layers
	workingLayers := make([]string, len(layerPaths))

	// Base layer: copy to disk-specific directory (read-only reference)
	basePath := filepath.Join(diskLayersDir, "base.qcow2")
	if err := m.createLayerReference(ctx, layerPaths[0], basePath); err != nil {
		return fmt.Errorf("failed to create base layer reference: %w", err)
	}
	workingLayers[0] = basePath

	// Delta layers: create with backing chain
	for i := 1; i < len(layerPaths); i++ {
		deltaPath := filepath.Join(diskLayersDir, fmt.Sprintf("delta-%d.qcow2", i))

		// Copy the cached layer
		if err := m.copyFile(layerPaths[i], deltaPath); err != nil {
			return fmt.Errorf("failed to copy delta layer: %w", err)
		}

		// Rebase to point to previous layer in chain
		if err := m.qcow2Client.RebaseUnsafe(ctx, deltaPath, workingLayers[i-1]); err != nil {
			return fmt.Errorf("failed to rebase delta layer: %w", err)
		}

		workingLayers[i] = deltaPath
	}

	// The top layer in the chain is the working image
	topLayer := workingLayers[len(workingLayers)-1]

	// Copy top layer to final path (this is the disk's working image)
	if err := m.copyFile(topLayer, finalPath); err != nil {
		return fmt.Errorf("failed to create working image: %w", err)
	}

	return nil
}

// createLayerReference creates a read-only reference to a cached layer
func (m *manager) createLayerReference(ctx context.Context, cachedLayerPath, refPath string) error {
	// Simply copy for now; could use hard links for efficiency
	return m.copyFile(cachedLayerPath, refPath)
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

	// Get disk's layer mappings
	diskLayers, err := m.stateDB.GetDiskLayers(name)
	if err != nil {
		return fmt.Errorf("failed to get disk layers: %w", err)
	}

	// Decrement ref count for each layer
	for _, mapping := range diskLayers {
		if err := m.stateDB.DecrementLayerRefCount(mapping.LayerID); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to decrement ref count for layer %s: %v\n", mapping.LayerID, err)
		}
	}

	// Delete disk-layer mappings
	if err := m.stateDB.DeleteDiskLayers(name); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to delete disk-layer mappings: %v\n", err)
	}

	// Clean up unused layers from cache
	if cleanedCount, err := m.layerCache.CleanupUnusedLayers(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to cleanup unused layers: %v\n", err)
	} else if cleanedCount > 0 {
		fmt.Fprintf(os.Stderr, "info: cleaned up %d unused layers from cache\n", cleanedCount)
	}

	// Delete QCOW2 image and disk-specific layer directory
	imagePath := filepath.Join(m.config.DataDir, "disks", name+".qcow2")
	if err := os.Remove(imagePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete disk image: %w", err)
	}

	diskLayersDir := filepath.Join(m.config.DataDir, "layers", name)
	if err := os.RemoveAll(diskLayersDir); err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "warning: failed to delete disk layers directory: %v\n", err)
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

// CleanupUnusedLayers removes cached layers that are no longer referenced by any disk
func (m *manager) CleanupUnusedLayers(ctx context.Context) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.layerCache.CleanupUnusedLayers(ctx)
}

// Fork creates a new disk that shares all existing layers of the source disk
// Both disks will have independent write layers for independent operation
func (m *manager) Fork(ctx context.Context, sourceDiskName, newDiskName string) (Disk, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validation: Check source disk exists
	sourceDisk, exists := m.disks[sourceDiskName]
	if !exists {
		// Try to open it
		sourceState, err := m.stateDB.GetDisk(sourceDiskName)
		if err != nil {
			return nil, fmt.Errorf("failed to get source disk state: %w", err)
		}
		if sourceState == nil {
			return nil, ErrDiskNotFound
		}

		// Check if image file exists
		imagePath := filepath.Join(m.config.DataDir, "disks", sourceDiskName+".qcow2")
		if _, err := os.Stat(imagePath); os.IsNotExist(err) {
			return nil, ErrDiskNotFound
		}

		// Create temporary disk object for validation
		sourceDisk = &disk{
			name:        sourceDiskName,
			sizeGB:      sourceState.SizeGB,
			imagePath:   imagePath,
			qcow2Client: m.qcow2Client,
			s3Client:    m.s3Client,
			stateDB:     m.stateDB,
			config:      m.config,
			pool:        m.pool,
			isMounted:   sourceState.IsMounted,
			mountPath:   sourceState.MountPath,
		}
	}

	// Validation: Check source disk is not mounted
	if sourceDisk.IsMounted() {
		return nil, fmt.Errorf("cannot fork mounted disk: unmount it first")
	}

	// Validation: Check new disk name doesn't exist
	if _, exists := m.disks[newDiskName]; exists {
		return nil, ErrDiskExists
	}

	existing, err := m.stateDB.GetDisk(newDiskName)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing disk: %w", err)
	}
	if existing != nil {
		return nil, ErrDiskExists
	}

	// Get source disk's layer mappings
	sourceLayers, err := m.stateDB.GetDiskLayers(sourceDiskName)
	if err != nil {
		return nil, fmt.Errorf("failed to get source disk layers: %w", err)
	}

	if len(sourceLayers) == 0 {
		return nil, fmt.Errorf("source disk has no layers to share")
	}

	// Get source disk state
	sourceState, err := m.stateDB.GetDisk(sourceDiskName)
	if err != nil {
		return nil, fmt.Errorf("failed to get source disk state: %w", err)
	}
	if sourceState == nil {
		return nil, ErrDiskNotFound
	}

	// Validate source disk doesn't have circular backing file references before forking
	if err := m.qcow2Client.ValidateBackingChain(ctx, sourceDisk.imagePath); err != nil {
		return nil, fmt.Errorf("source disk has invalid backing file chain: %w", err)
	}

	// Always commit the source disk's working layer as a new shared layer during fork
	// This ensures all uncommitted changes (even small ones) are preserved in both disks
	// TODO: Optimize by adding a size check (e.g., using GetActualSize) to skip committing
	//       if the working layer is empty or unchanged, similar to pushIncrementalLayer logic
	// Get file size for metadata
	fileInfo, err := os.Stat(sourceDisk.imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get source disk file info: %w", err)
	}
	fileSize := fileInfo.Size()

	var newSharedLayerID string
	var newSharedLayerPath string

	// Calculate checksum of working image
	checksum, err := m.qcow2Client.Checksum(ctx, sourceDisk.imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate checksum: %w", err)
	}

	// Create new layer ID
	newSharedLayerID = fmt.Sprintf("layer-%d", time.Now().Unix())

	// Create temp copy of working layer
	// Keep the backing file reference - we'll rebase it when building the chain
	// This preserves the layer structure and data access
	tempLayerPath := sourceDisk.imagePath + ".fork-temp"
	defer os.Remove(tempLayerPath)

	if err := m.copyFile(sourceDisk.imagePath, tempLayerPath); err != nil {
		return nil, fmt.Errorf("failed to copy working layer: %w", err)
	}

	// Copy layer to cache directory WITH backing file reference
	// The backing file path will be updated when we rebase it during chain building
	newSharedLayerPath = m.layerCache.GetLayerPath(newSharedLayerID)
	if err := m.copyFile(tempLayerPath, newSharedLayerPath); err != nil {
		return nil, fmt.Errorf("failed to copy layer to cache: %w", err)
	}

	// Create layer state and save it
	layerState := &LayerState{
		ID:       newSharedLayerID,
		Checksum: checksum,
		Size:     fileSize,
		CachedAt: time.Now(),
		RefCount: 2, // Will be referenced by both source and new disk
	}
	if err := m.stateDB.SaveLayer(layerState); err != nil {
		// Cleanup: remove layer file
		os.Remove(newSharedLayerPath)
		return nil, fmt.Errorf("failed to save layer state: %w", err)
	}

	// Add the new layer to source disk's layer mappings
	newLayerPosition := len(sourceLayers)
	if err := m.stateDB.AddDiskLayerMapping(sourceDiskName, newSharedLayerID, newLayerPosition); err != nil {
		// Cleanup: remove layer file and decrement ref count (will delete if ref count goes to 0)
		os.Remove(newSharedLayerPath)
		m.stateDB.DecrementLayerRefCount(newSharedLayerID)
		return nil, fmt.Errorf("failed to add new layer to source disk mappings: %w", err)
	}

	// Update source layers list to include the new layer
	sourceLayers = append(sourceLayers, &DiskLayerMapping{
		DiskName: sourceDiskName,
		LayerID:  newSharedLayerID,
		Position: newLayerPosition,
	})

	// For new disk: Create disk state entry
	now := time.Now()
	newDiskState := &DiskState{
		Name:       newDiskName,
		SizeGB:     sourceState.SizeGB,
		CreatedAt:  now,
		ModifiedAt: now,
		IsMounted:  false,
		MountPath:  "",
		InS3:       sourceState.InS3,
		Checksum:   "", // Will be set on first push
	}

	// Copy all layer mappings from source disk and increment ref counts
	for _, sourceLayer := range sourceLayers {
		// Copy the layer mapping
		if err := m.stateDB.AddDiskLayerMapping(newDiskName, sourceLayer.LayerID, sourceLayer.Position); err != nil {
			// Cleanup: rollback layer mappings
			m.stateDB.DeleteDiskLayers(newDiskName)
			return nil, fmt.Errorf("failed to add disk-layer mapping: %w", err)
		}

		// Increment ref count for the shared layer
		if err := m.stateDB.IncrementLayerRefCount(sourceLayer.LayerID); err != nil {
			// Cleanup: rollback layer mappings and decrement ref counts
			m.stateDB.DeleteDiskLayers(newDiskName)
			for _, sl := range sourceLayers {
				m.stateDB.DecrementLayerRefCount(sl.LayerID)
			}
			return nil, fmt.Errorf("failed to increment ref count for layer %s: %w", sourceLayer.LayerID, err)
		}
	}

	// Save new disk state
	if err := m.stateDB.SaveDisk(newDiskState); err != nil {
		// Cleanup: rollback layer mappings and decrement ref counts
		m.stateDB.DeleteDiskLayers(newDiskName)
		for _, sl := range sourceLayers {
			m.stateDB.DecrementLayerRefCount(sl.LayerID)
		}
		return nil, fmt.Errorf("failed to save new disk state: %w", err)
	}

	// Create new disk-specific layers directory
	newDiskLayersDir := filepath.Join(m.config.DataDir, "layers", newDiskName)
	if err := os.MkdirAll(newDiskLayersDir, 0755); err != nil {
		// Cleanup
		m.stateDB.DeleteDisk(newDiskName)
		m.stateDB.DeleteDiskLayers(newDiskName)
		for _, sl := range sourceLayers {
			m.stateDB.DecrementLayerRefCount(sl.LayerID)
		}
		return nil, fmt.Errorf("failed to create new disk layers directory: %w", err)
	}

	// Build the same backing file chain as source disk
	// Get cached layer paths (these are the shared physical files)
	layerPaths := make([]string, len(sourceLayers))
	for i, sourceLayer := range sourceLayers {
		layerPath := m.layerCache.GetLayerPath(sourceLayer.LayerID)
		// Verify layer exists in cache
		if _, err := os.Stat(layerPath); os.IsNotExist(err) {
			// Cleanup
			os.RemoveAll(newDiskLayersDir)
			m.stateDB.DeleteDisk(newDiskName)
			m.stateDB.DeleteDiskLayers(newDiskName)
			for _, sl := range sourceLayers {
				m.stateDB.DecrementLayerRefCount(sl.LayerID)
			}
			return nil, fmt.Errorf("layer %s not found in cache", sourceLayer.LayerID)
		}
		layerPaths[i] = layerPath
	}

	// Build backing file chain in new disk's directory (similar to pullLayeredDisk)
	workingLayers := make([]string, len(layerPaths))

	// Base layer: create reference to cached layer
	basePath := filepath.Join(newDiskLayersDir, "base.qcow2")
	if err := m.createLayerReference(ctx, layerPaths[0], basePath); err != nil {
		// Cleanup
		os.RemoveAll(newDiskLayersDir)
		m.stateDB.DeleteDisk(newDiskName)
		m.stateDB.DeleteDiskLayers(newDiskName)
		for _, sl := range sourceLayers {
			m.stateDB.DecrementLayerRefCount(sl.LayerID)
		}
		return nil, fmt.Errorf("failed to create base layer reference: %w", err)
	}
	workingLayers[0] = basePath

	// Delta layers: create with backing chain pointing to cached layers
	for i := 1; i < len(layerPaths); i++ {
		deltaPath := filepath.Join(newDiskLayersDir, fmt.Sprintf("delta-%d.qcow2", i))

		// Copy the cached layer
		if err := m.copyFile(layerPaths[i], deltaPath); err != nil {
			// Cleanup
			os.RemoveAll(newDiskLayersDir)
			m.stateDB.DeleteDisk(newDiskName)
			m.stateDB.DeleteDiskLayers(newDiskName)
			for _, sl := range sourceLayers {
				m.stateDB.DecrementLayerRefCount(sl.LayerID)
			}
			return nil, fmt.Errorf("failed to copy delta layer: %w", err)
		}

		// Rebase to point to previous layer in chain
		if err := m.qcow2Client.RebaseUnsafe(ctx, deltaPath, workingLayers[i-1]); err != nil {
			// Cleanup
			os.RemoveAll(newDiskLayersDir)
			m.stateDB.DeleteDisk(newDiskName)
			m.stateDB.DeleteDiskLayers(newDiskName)
			for _, sl := range sourceLayers {
				m.stateDB.DecrementLayerRefCount(sl.LayerID)
			}
			return nil, fmt.Errorf("failed to rebase delta layer: %w", err)
		}

		workingLayers[i] = deltaPath
	}

	// The top layer in the chain becomes the base for the working image
	topLayer := workingLayers[len(workingLayers)-1]

	// Create new disk's working image path
	newDiskImagePath := filepath.Join(m.config.DataDir, "disks", newDiskName+".qcow2")

	// Create new empty working layer directly with topLayer as backing
	// This preserves all data from topLayer through the backing chain
	// The new empty layer on top allows independent writes for the new disk
	if err := m.qcow2Client.CreateWithBacking(ctx, topLayer, newDiskImagePath, int(newDiskState.SizeGB)); err != nil {
		// Cleanup
		os.RemoveAll(newDiskLayersDir)
		m.stateDB.DeleteDisk(newDiskName)
		m.stateDB.DeleteDiskLayers(newDiskName)
		for _, sl := range sourceLayers {
			m.stateDB.DecrementLayerRefCount(sl.LayerID)
		}
		return nil, fmt.Errorf("failed to create new working layer for new disk: %w", err)
	}

	// Final validation: ensure the forked disk has a valid backing chain
	if err := m.qcow2Client.ValidateBackingChain(ctx, newDiskImagePath); err != nil {
		// Cleanup
		os.Remove(newDiskImagePath)
		os.RemoveAll(newDiskLayersDir)
		m.stateDB.DeleteDisk(newDiskName)
		m.stateDB.DeleteDiskLayers(newDiskName)
		for _, sl := range sourceLayers {
			m.stateDB.DecrementLayerRefCount(sl.LayerID)
		}
		return nil, fmt.Errorf("invalid backing file chain in forked disk: %w", err)
	}

	// For source disk: Create new empty working layer preserving current state
	// Use the newly committed layer as backing (we always commit during fork)
	tempSourceLayer := sourceDisk.imagePath + ".new"
	if err := m.qcow2Client.CreateWithBacking(ctx, newSharedLayerPath, tempSourceLayer, int(sourceState.SizeGB)); err != nil {
		// Cleanup
		os.Remove(newDiskImagePath)
		os.RemoveAll(newDiskLayersDir)
		m.stateDB.DeleteDisk(newDiskName)
		m.stateDB.DeleteDiskLayers(newDiskName)
		for _, sl := range sourceLayers {
			m.stateDB.DecrementLayerRefCount(sl.LayerID)
		}
		return nil, fmt.Errorf("failed to create new working layer for source disk: %w", err)
	}

	// Replace the working image with the new empty layer
	if err := os.Rename(tempSourceLayer, sourceDisk.imagePath); err != nil {
		// Cleanup
		os.Remove(newDiskImagePath)
		os.RemoveAll(newDiskLayersDir)
		m.stateDB.DeleteDisk(newDiskName)
		m.stateDB.DeleteDiskLayers(newDiskName)
		for _, sl := range sourceLayers {
			m.stateDB.DecrementLayerRefCount(sl.LayerID)
		}
		return nil, fmt.Errorf("failed to replace source disk working image: %w", err)
	}

	// Create disk object
	newDisk := &disk{
		name:        newDiskName,
		sizeGB:      newDiskState.SizeGB,
		imagePath:   newDiskImagePath,
		qcow2Client: m.qcow2Client,
		s3Client:    m.s3Client,
		stateDB:     m.stateDB,
		config:      m.config,
		pool:        m.pool,
		isMounted:   false,
		mountPath:   "",
	}

	m.disks[newDiskName] = newDisk

	return newDisk, nil
}
