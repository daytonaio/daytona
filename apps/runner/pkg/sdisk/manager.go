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

	// Disk image created successfully - no initialization mount needed

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
	// CRITICAL LOGGING - log immediately before lock
	earlyLog, _ := os.OpenFile("/tmp/fork-early.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if earlyLog != nil {
		fmt.Fprintf(earlyLog, "\n=== FORK CALLED: source=%s, target=%s, time=%v ===\n", sourceDiskName, newDiskName, time.Now())
		earlyLog.Close()
	}

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

	// Step 1: Ensure disk is properly unmounted from pool if needed
	logFile, _ := os.OpenFile("/tmp/fork-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if logFile != nil {
		defer logFile.Close()
		fmt.Fprintf(logFile, "\n=== FORK START for source=%s, target=%s ===\n", sourceDiskName, newDiskName)
		fmt.Fprintf(logFile, "[FORK-DEBUG] Source disk imagePath: %s\n", sourceDisk.imagePath)
		fmt.Fprintf(logFile, "[FORK-DEBUG] Source disk '%s' IsMounted: %v, MountPath: %s\n", sourceDiskName, sourceDisk.IsMounted(), sourceDisk.MountPath())
	}
	fmt.Fprintf(os.Stderr, "[FORK-DEBUG] Source disk imagePath: %s\n", sourceDisk.imagePath)
	fmt.Fprintf(os.Stderr, "[FORK-DEBUG] Source disk '%s' IsMounted: %v, MountPath: %s\n", sourceDiskName, sourceDisk.IsMounted(), sourceDisk.MountPath())

	// CRITICAL: Wait for any ongoing unmount operations to complete
	// The sandbox Stop operation runs docker cp and unmount, which might still be in progress
	// We need to wait for the disk to be fully unmounted before forking
	fmt.Fprintf(os.Stderr, "[FORK-DEBUG] Waiting for disk to be fully unmounted (max 10 seconds)...\n")
	if logFile != nil {
		fmt.Fprintf(logFile, "[FORK-DEBUG] Waiting for disk to be fully unmounted (max 10 seconds)...\n")
	}
	for i := 0; i < 20; i++ {
		if !sourceDisk.IsMounted() {
			fmt.Fprintf(os.Stderr, "[FORK-DEBUG] Disk is unmounted after %d checks (%.1f seconds)\n", i, float64(i)*0.5)
			if logFile != nil {
				fmt.Fprintf(logFile, "[FORK-DEBUG] Disk is unmounted after %d checks (%.1f seconds)\n", i, float64(i)*0.5)
			}
			break
		}
		if i == 19 {
			fmt.Fprintf(os.Stderr, "[FORK-DEBUG] WARNING: Disk still mounted after 10 seconds, proceeding anyway\n")
			if logFile != nil {
				fmt.Fprintf(logFile, "[FORK-DEBUG] WARNING: Disk still mounted after 10 seconds, proceeding anyway\n")
			}
		}
		time.Sleep(500 * time.Millisecond)
	}

	// Check if disk is in pool (it might be mounted via pool even if disk.IsMounted() is false)
	if m.pool != nil {
		fmt.Fprintf(os.Stderr, "[FORK-DEBUG] Checking if disk is in pool\n")
		if logFile != nil {
			fmt.Fprintf(logFile, "[FORK-DEBUG] Checking if disk is in pool\n")
		}
		// Try to evict from pool first to ensure it's properly unmounted and synced
		if err := m.pool.Evict(ctx, sourceDiskName); err != nil {
			// Evict might fail if not in pool, that's okay
			fmt.Fprintf(os.Stderr, "[FORK-DEBUG] Pool evict result: %v\n", err)
			if logFile != nil {
				fmt.Fprintf(logFile, "[FORK-DEBUG] Pool evict result: %v\n", err)
			}
		} else {
			fmt.Fprintf(os.Stderr, "[FORK-DEBUG] Successfully evicted disk from pool\n")
			if logFile != nil {
				fmt.Fprintf(logFile, "[FORK-DEBUG] Successfully evicted disk from pool\n")
			}
		}
	}

	// Now check mount state again and unmount if needed
	if sourceDisk.IsMounted() {
		fmt.Fprintf(os.Stderr, "[FORK-DEBUG] Disk still mounted after pool eviction, unmounting directly\n")
		if logFile != nil {
			fmt.Fprintf(logFile, "[FORK-DEBUG] Disk still mounted after pool eviction, unmounting directly\n")
		}
		if err := sourceDisk.Unmount(ctx); err != nil {
			return nil, fmt.Errorf("failed to unmount source disk: %w", err)
		}
		fmt.Fprintf(os.Stderr, "[FORK-DEBUG] Source disk '%s' unmounted successfully\n", sourceDiskName)
		if logFile != nil {
			fmt.Fprintf(logFile, "[FORK-DEBUG] Source disk '%s' unmounted successfully\n", sourceDiskName)
		}
	} else {
		fmt.Fprintf(os.Stderr, "[FORK-DEBUG] Source disk '%s' is not mounted\n", sourceDiskName)
		if logFile != nil {
			fmt.Fprintf(logFile, "[FORK-DEBUG] Source disk '%s' is not mounted\n", sourceDiskName)
		}
	}

	// DEBUG: Mount the imagePath directly to see if the file is actually in the QCOW2
	if logFile != nil {
		fmt.Fprintf(logFile, "[FORK] Verifying contents of imagePath QCOW2 file after unmount: %s\n", sourceDisk.imagePath)
		device, mountErr := m.qcow2Client.Connect(ctx, sourceDiskName+"-verify", sourceDisk.imagePath)
		if mountErr == nil {
			tempMount := filepath.Join(m.config.DataDir, "mounts", ".verify-"+sourceDiskName)
			if mountErr2 := m.qcow2Client.Mount(ctx, sourceDiskName+"-verify", device, tempMount); mountErr2 == nil {
				if entries, readErr := os.ReadDir(tempMount); readErr == nil {
					fmt.Fprintf(logFile, "[FORK] Files in imagePath QCOW2 after unmount:\n")
					for _, entry := range entries {
						info, _ := entry.Info()
						fmt.Fprintf(logFile, "  - %s (size: %d, isDir: %v)\n", entry.Name(), info.Size(), entry.IsDir())
					}
				} else {
					fmt.Fprintf(logFile, "[FORK] Failed to read directory from imagePath: %v\n", readErr)
				}
				m.qcow2Client.Unmount(ctx, sourceDiskName+"-verify")
			} else {
				fmt.Fprintf(logFile, "[FORK] Failed to mount imagePath for verification: %v\n", mountErr2)
			}
			m.qcow2Client.Disconnect(ctx, sourceDiskName+"-verify")
			os.RemoveAll(tempMount)
		} else {
			fmt.Fprintf(logFile, "[FORK] Failed to connect imagePath for verification: %v\n", mountErr)
		}
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

	// Log source disk's existing layers
	fmt.Fprintf(os.Stderr, "[FORK] Source disk '%s' has %d layers before commit:\n", sourceDiskName, len(sourceLayers))
	for i, layer := range sourceLayers {
		fmt.Fprintf(os.Stderr, "  [%d] LayerID: %s, Position: %d\n", i, layer.LayerID, layer.Position)
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

	// Step 2: Commit the source disk's working layer to a new shared layer
	// This is similar to how it's stored on S3 - commit the working layer
	// Create new layer ID for the committed working layer
	committedLayerID := fmt.Sprintf("layer-%d", time.Now().Unix())

	// DEBUG: Mount the source disk temporarily to see what files exist before commit
	if logFile != nil {
		fmt.Fprintf(logFile, "[FORK] About to commit working layer from: %s\n", sourceDisk.imagePath)
		// Try to mount and list files
		device, mountErr := m.qcow2Client.Connect(ctx, sourceDiskName+"-fork-check", sourceDisk.imagePath)
		if mountErr == nil {
			tempMount := filepath.Join(m.config.DataDir, "mounts", ".fork-check-"+sourceDiskName)
			if mountErr2 := m.qcow2Client.Mount(ctx, sourceDiskName+"-fork-check", device, tempMount); mountErr2 == nil {
				if entries, readErr := os.ReadDir(tempMount); readErr == nil {
					fmt.Fprintf(logFile, "[FORK] Files in source disk before commit:\n")
					for _, entry := range entries {
						info, _ := entry.Info()
						fmt.Fprintf(logFile, "  - %s (size: %d, isDir: %v)\n", entry.Name(), info.Size(), entry.IsDir())
					}
				} else {
					fmt.Fprintf(logFile, "[FORK] Failed to read directory: %v\n", readErr)
				}
				m.qcow2Client.Unmount(ctx, sourceDiskName+"-fork-check")
			}
			m.qcow2Client.Disconnect(ctx, sourceDiskName+"-fork-check")
			os.RemoveAll(tempMount)
		}
	}

	// Create temp copy of working layer and flatten it to include all data
	// CRITICAL: We need to flatten the working layer (not just remove backing reference)
	// because it may be a delta layer that depends on its backing file
	// Flattening merges all data from the backing chain into a standalone image
	tempLayerPath := sourceDisk.imagePath + ".fork-temp"
	defer os.Remove(tempLayerPath)

	// Check if the working image has a backing file
	backingFile, err := m.qcow2Client.GetBackingFile(ctx, sourceDisk.imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to check backing file: %w", err)
	}

	if backingFile != "" {
		// Flatten the image to include all data from the backing chain
		// This is similar to how pushBaseLayer works
		if err := m.qcow2Client.Convert(ctx, sourceDisk.imagePath, tempLayerPath); err != nil {
			return nil, fmt.Errorf("failed to flatten working layer: %w", err)
		}
		fmt.Fprintf(os.Stderr, "[FORK] Flattened working layer (had backing file: %s)\n", backingFile)
	} else {
		// No backing file, just copy it
		if err := m.copyFile(sourceDisk.imagePath, tempLayerPath); err != nil {
			return nil, fmt.Errorf("failed to copy working layer: %w", err)
		}
		fmt.Fprintf(os.Stderr, "[FORK] Working layer has no backing file, copied directly\n")
	}

	// Copy committed layer to cache directory
	committedLayerPath := m.layerCache.GetLayerPath(committedLayerID)
	if err := m.copyFile(tempLayerPath, committedLayerPath); err != nil {
		return nil, fmt.Errorf("failed to copy committed layer to cache: %w", err)
	}

	// Get file size and checksum of the flattened/committed layer
	fileInfo, err := os.Stat(committedLayerPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get committed layer file info: %w", err)
	}
	fileSize := fileInfo.Size()

	// Calculate checksum of the committed layer (after flattening)
	checksum, err := m.qcow2Client.Checksum(ctx, committedLayerPath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate checksum: %w", err)
	}

	// Create layer state and save it
	layerState := &LayerState{
		ID:       committedLayerID,
		Checksum: checksum,
		Size:     fileSize,
		CachedAt: time.Now(),
		RefCount: 2, // Will be referenced by both source and new disk
	}
	if err := m.stateDB.SaveLayer(layerState); err != nil {
		// Cleanup: remove layer file
		os.Remove(committedLayerPath)
		return nil, fmt.Errorf("failed to save layer state: %w", err)
	}

	// Add the committed layer to source disk's layer mappings
	newLayerPosition := len(sourceLayers)
	if err := m.stateDB.AddDiskLayerMapping(sourceDiskName, committedLayerID, newLayerPosition); err != nil {
		// Cleanup: remove layer file and decrement ref count
		os.Remove(committedLayerPath)
		m.stateDB.DecrementLayerRefCount(committedLayerID)
		return nil, fmt.Errorf("failed to add committed layer to source disk mappings: %w", err)
	}

	// Update source layers list to include the committed layer
	sourceLayers = append(sourceLayers, &DiskLayerMapping{
		DiskName: sourceDiskName,
		LayerID:  committedLayerID,
		Position: newLayerPosition,
	})

	// Log the committed layer
	fmt.Fprintf(os.Stderr, "[FORK] Committed working layer as new shared layer:\n")
	fmt.Fprintf(os.Stderr, "  LayerID: %s, Position: %d, Size: %d bytes, Checksum: %s\n", committedLayerID, newLayerPosition, fileSize, checksum)
	fmt.Fprintf(os.Stderr, "[FORK] Source disk '%s' now has %d layers (including committed layer)\n", sourceDiskName, len(sourceLayers))

	// Step 3: Create new disk state entry
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

	// Copy all layer mappings from source disk (including the newly committed layer) and increment ref counts
	fmt.Fprintf(os.Stderr, "[FORK] Copying %d layer mappings to new disk '%s':\n", len(sourceLayers), newDiskName)
	for _, sourceLayer := range sourceLayers {
		// Copy the layer mapping
		if err := m.stateDB.AddDiskLayerMapping(newDiskName, sourceLayer.LayerID, sourceLayer.Position); err != nil {
			// Cleanup: rollback layer mappings
			m.stateDB.DeleteDiskLayers(newDiskName)
			return nil, fmt.Errorf("failed to add disk-layer mapping: %w", err)
		}
		fmt.Fprintf(os.Stderr, "  [%d] LayerID: %s, Position: %d\n", sourceLayer.Position, sourceLayer.LayerID, sourceLayer.Position)

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

	// Create disk-specific layers directories
	sourceDiskLayersDir := filepath.Join(m.config.DataDir, "layers", sourceDiskName)
	newDiskLayersDir := filepath.Join(m.config.DataDir, "layers", newDiskName)
	if err := os.MkdirAll(sourceDiskLayersDir, 0755); err != nil {
		// Cleanup
		m.stateDB.DeleteDisk(newDiskName)
		m.stateDB.DeleteDiskLayers(newDiskName)
		for _, sl := range sourceLayers {
			m.stateDB.DecrementLayerRefCount(sl.LayerID)
		}
		return nil, fmt.Errorf("failed to create source disk layers directory: %w", err)
	}
	if err := os.MkdirAll(newDiskLayersDir, 0755); err != nil {
		// Cleanup
		os.RemoveAll(sourceDiskLayersDir)
		m.stateDB.DeleteDisk(newDiskName)
		m.stateDB.DeleteDiskLayers(newDiskName)
		for _, sl := range sourceLayers {
			m.stateDB.DecrementLayerRefCount(sl.LayerID)
		}
		return nil, fmt.Errorf("failed to create new disk layers directory: %w", err)
	}

	// Step 4: Build backing file chain for both disks
	// Get cached layer paths (these are the shared physical files)
	fmt.Fprintf(os.Stderr, "[FORK] Building backing chain from %d cached layers:\n", len(sourceLayers))
	layerPaths := make([]string, len(sourceLayers))
	for i, sourceLayer := range sourceLayers {
		layerPath := m.layerCache.GetLayerPath(sourceLayer.LayerID)
		// Verify layer exists in cache
		if _, err := os.Stat(layerPath); os.IsNotExist(err) {
			// Cleanup
			os.RemoveAll(sourceDiskLayersDir)
			os.RemoveAll(newDiskLayersDir)
			m.stateDB.DeleteDisk(newDiskName)
			m.stateDB.DeleteDiskLayers(newDiskName)
			for _, sl := range sourceLayers {
				m.stateDB.DecrementLayerRefCount(sl.LayerID)
			}
			return nil, fmt.Errorf("layer %s not found in cache", sourceLayer.LayerID)
		}
		layerPaths[i] = layerPath
		fmt.Fprintf(os.Stderr, "  [%d] LayerID: %s -> Path: %s\n", i, sourceLayer.LayerID, layerPath)
	}

	// Helper function to build backing chain
	buildBackingChain := func(diskLayersDir string) (string, error) {
		workingLayers := make([]string, 0)

		// First, check if any layer is standalone (flattened) - if so, find the last one
		// Standalone layers contain all data, so we can use them directly without building a chain
		lastStandaloneIndex := -1
		for i := len(layerPaths) - 1; i >= 0; i-- {
			layerBackingFile, err := m.qcow2Client.GetBackingFile(ctx, layerPaths[i])
			if err != nil {
				return "", fmt.Errorf("failed to check layer backing file: %w", err)
			}
			if layerBackingFile == "" {
				// Found a standalone/flattened layer
				lastStandaloneIndex = i
				fmt.Fprintf(os.Stderr, "[FORK] Found standalone (flattened) layer at position %d, will use directly\n", i)
				break
			}
		}

		if lastStandaloneIndex >= 0 {
			// We have a standalone layer - use it directly as the base
			// It contains all data from previous layers, so we don't need them
			basePath := filepath.Join(diskLayersDir, "base.qcow2")
			if err := m.copyFile(layerPaths[lastStandaloneIndex], basePath); err != nil {
				return "", fmt.Errorf("failed to copy standalone layer as base: %w", err)
			}
			workingLayers = append(workingLayers, basePath)

			// Process any layers after the standalone one (shouldn't happen in fork, but handle it)
			for i := lastStandaloneIndex + 1; i < len(layerPaths); i++ {
				deltaPath := filepath.Join(diskLayersDir, fmt.Sprintf("delta-%d.qcow2", i))
				if err := m.copyFile(layerPaths[i], deltaPath); err != nil {
					return "", fmt.Errorf("failed to copy delta layer: %w", err)
				}
				// Rebase to point to previous layer
				if err := m.qcow2Client.RebaseUnsafe(ctx, deltaPath, workingLayers[len(workingLayers)-1]); err != nil {
					return "", fmt.Errorf("failed to rebase delta layer: %w", err)
				}
				workingLayers = append(workingLayers, deltaPath)
			}
		} else {
			// No standalone layer - build normal chain
			// Base layer: create reference to cached layer
			basePath := filepath.Join(diskLayersDir, "base.qcow2")
			if err := m.createLayerReference(ctx, layerPaths[0], basePath); err != nil {
				return "", fmt.Errorf("failed to create base layer reference: %w", err)
			}
			workingLayers = append(workingLayers, basePath)

			// Delta layers: create with backing chain
			for i := 1; i < len(layerPaths); i++ {
				deltaPath := filepath.Join(diskLayersDir, fmt.Sprintf("delta-%d.qcow2", i))
				if err := m.copyFile(layerPaths[i], deltaPath); err != nil {
					return "", fmt.Errorf("failed to copy delta layer: %w", err)
				}
				// Rebase to point to previous layer in chain
				if err := m.qcow2Client.RebaseUnsafe(ctx, deltaPath, workingLayers[len(workingLayers)-1]); err != nil {
					return "", fmt.Errorf("failed to rebase delta layer: %w", err)
				}
				workingLayers = append(workingLayers, deltaPath)
			}
		}

		// Return the top layer in the chain
		return workingLayers[len(workingLayers)-1], nil
	}

	// Build backing chain for new disk
	fmt.Fprintf(os.Stderr, "[FORK] Building backing chain for new disk '%s' in %s\n", newDiskName, newDiskLayersDir)
	newDiskTopLayer, err := buildBackingChain(newDiskLayersDir)
	if err != nil {
		// Cleanup
		os.RemoveAll(sourceDiskLayersDir)
		os.RemoveAll(newDiskLayersDir)
		m.stateDB.DeleteDisk(newDiskName)
		m.stateDB.DeleteDiskLayers(newDiskName)
		for _, sl := range sourceLayers {
			m.stateDB.DecrementLayerRefCount(sl.LayerID)
		}
		return nil, fmt.Errorf("failed to build backing chain for new disk: %w", err)
	}
	fmt.Fprintf(os.Stderr, "[FORK] New disk '%s' top layer: %s\n", newDiskName, newDiskTopLayer)

	// Build backing chain for source disk
	fmt.Fprintf(os.Stderr, "[FORK] Building backing chain for source disk '%s' in %s\n", sourceDiskName, sourceDiskLayersDir)
	sourceDiskTopLayer, err := buildBackingChain(sourceDiskLayersDir)
	if err != nil {
		// Cleanup
		os.RemoveAll(sourceDiskLayersDir)
		os.RemoveAll(newDiskLayersDir)
		m.stateDB.DeleteDisk(newDiskName)
		m.stateDB.DeleteDiskLayers(newDiskName)
		for _, sl := range sourceLayers {
			m.stateDB.DecrementLayerRefCount(sl.LayerID)
		}
		return nil, fmt.Errorf("failed to build backing chain for source disk: %w", err)
	}
	fmt.Fprintf(os.Stderr, "[FORK] Source disk '%s' top layer: %s\n", sourceDiskName, sourceDiskTopLayer)

	// Step 5: Create new empty working layers for both disks
	// New disk's working image path
	newDiskImagePath := filepath.Join(m.config.DataDir, "disks", newDiskName+".qcow2")
	if err := m.qcow2Client.CreateWithBacking(ctx, newDiskTopLayer, newDiskImagePath, int(newDiskState.SizeGB)); err != nil {
		// Cleanup
		os.RemoveAll(sourceDiskLayersDir)
		os.RemoveAll(newDiskLayersDir)
		m.stateDB.DeleteDisk(newDiskName)
		m.stateDB.DeleteDiskLayers(newDiskName)
		for _, sl := range sourceLayers {
			m.stateDB.DecrementLayerRefCount(sl.LayerID)
		}
		return nil, fmt.Errorf("failed to create new working layer for new disk: %w", err)
	}

	// Source disk's new working image (replace the old one)
	tempSourceImage := sourceDisk.imagePath + ".new"
	if err := m.qcow2Client.CreateWithBacking(ctx, sourceDiskTopLayer, tempSourceImage, int(sourceState.SizeGB)); err != nil {
		// Cleanup
		os.Remove(newDiskImagePath)
		os.RemoveAll(sourceDiskLayersDir)
		os.RemoveAll(newDiskLayersDir)
		m.stateDB.DeleteDisk(newDiskName)
		m.stateDB.DeleteDiskLayers(newDiskName)
		for _, sl := range sourceLayers {
			m.stateDB.DecrementLayerRefCount(sl.LayerID)
		}
		return nil, fmt.Errorf("failed to create new working layer for source disk: %w", err)
	}

	// Replace the source disk's working image
	if err := os.Rename(tempSourceImage, sourceDisk.imagePath); err != nil {
		// Cleanup
		os.Remove(newDiskImagePath)
		os.RemoveAll(sourceDiskLayersDir)
		os.RemoveAll(newDiskLayersDir)
		m.stateDB.DeleteDisk(newDiskName)
		m.stateDB.DeleteDiskLayers(newDiskName)
		for _, sl := range sourceLayers {
			m.stateDB.DecrementLayerRefCount(sl.LayerID)
		}
		return nil, fmt.Errorf("failed to replace source disk working image: %w", err)
	}

	// Final validation: ensure both disks have valid backing chains
	if err := m.qcow2Client.ValidateBackingChain(ctx, newDiskImagePath); err != nil {
		// Cleanup
		os.Remove(newDiskImagePath)
		os.RemoveAll(sourceDiskLayersDir)
		os.RemoveAll(newDiskLayersDir)
		m.stateDB.DeleteDisk(newDiskName)
		m.stateDB.DeleteDiskLayers(newDiskName)
		for _, sl := range sourceLayers {
			m.stateDB.DecrementLayerRefCount(sl.LayerID)
		}
		return nil, fmt.Errorf("invalid backing file chain in new disk: %w", err)
	}

	if err := m.qcow2Client.ValidateBackingChain(ctx, sourceDisk.imagePath); err != nil {
		// Cleanup
		os.Remove(newDiskImagePath)
		os.RemoveAll(sourceDiskLayersDir)
		os.RemoveAll(newDiskLayersDir)
		m.stateDB.DeleteDisk(newDiskName)
		m.stateDB.DeleteDiskLayers(newDiskName)
		for _, sl := range sourceLayers {
			m.stateDB.DecrementLayerRefCount(sl.LayerID)
		}
		return nil, fmt.Errorf("invalid backing file chain in source disk: %w", err)
	}

	// Update source disk's database state (unmounted after fork)
	if err := m.stateDB.UpdateMountState(sourceDiskName, false, ""); err != nil {
		// Log warning but don't fail - the disk state is correct in memory
		fmt.Fprintf(os.Stderr, "warning: failed to update source disk mount state: %v\n", err)
	}

	// Update source disk's in-memory state and ensure it's registered
	if existingSourceDisk, exists := m.disks[sourceDiskName]; exists {
		existingSourceDisk.isMounted = false
		existingSourceDisk.mountPath = ""
	} else {
		// Register source disk in memory if it wasn't already there
		// The source disk keeps its original ID (sourceDiskName)
		sourceDisk.isMounted = false
		sourceDisk.mountPath = ""
		m.disks[sourceDiskName] = sourceDisk
	}

	// Create new disk object with the new ID (newDiskName)
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

	// Register new disk with its new ID
	m.disks[newDiskName] = newDisk

	// Final logging: verify layers for both disks
	newDiskLayers, err := m.stateDB.GetDiskLayers(newDiskName)
	if err == nil {
		fmt.Fprintf(os.Stderr, "[FORK] New disk '%s' final layer configuration (%d layers):\n", newDiskName, len(newDiskLayers))
		for _, layer := range newDiskLayers {
			fmt.Fprintf(os.Stderr, "  [%d] LayerID: %s\n", layer.Position, layer.LayerID)
		}
	}

	sourceDiskLayersFinal, err := m.stateDB.GetDiskLayers(sourceDiskName)
	if err == nil {
		fmt.Fprintf(os.Stderr, "[FORK] Source disk '%s' final layer configuration (%d layers):\n", sourceDiskName, len(sourceDiskLayersFinal))
		for _, layer := range sourceDiskLayersFinal {
			fmt.Fprintf(os.Stderr, "  [%d] LayerID: %s\n", layer.Position, layer.LayerID)
		}
	}

	// Add detailed layer information for manual testing
	if logFile != nil && newDiskLayers != nil {
		fmt.Fprintf(logFile, "\n=== NEW DISK LAYER DETAILS FOR MANUAL TESTING ===\n")
		fmt.Fprintf(logFile, "New Disk Name: %s\n", newDiskName)
		fmt.Fprintf(logFile, "New Disk imagePath: %s\n", newDisk.imagePath)
		fmt.Fprintf(logFile, "New Disk top layer path: %s\n", filepath.Join(m.config.DataDir, "layers", newDiskName, "base.qcow2"))
		fmt.Fprintf(logFile, "\nLayer Details:\n")
		for i, layer := range newDiskLayers {
			layerPath := m.layerCache.GetLayerPath(layer.LayerID)
			fmt.Fprintf(logFile, "  Layer %d:\n", i)
			fmt.Fprintf(logFile, "    LayerID: %s\n", layer.LayerID)
			fmt.Fprintf(logFile, "    Position: %d\n", layer.Position)
			fmt.Fprintf(logFile, "    Cache Path: %s\n", layerPath)
			if fileInfo, err := os.Stat(layerPath); err == nil {
				fmt.Fprintf(logFile, "    Size: %d bytes\n", fileInfo.Size())
			}
		}

		fmt.Fprintf(logFile, "\n=== MANUAL TEST COMMANDS ===\n")
		topLayer := filepath.Join(m.config.DataDir, "layers", newDiskName, "base.qcow2")
		fmt.Fprintf(logFile, "# Connect and mount the new disk's top layer:\n")
		fmt.Fprintf(logFile, "sudo qemu-nbd --connect=/dev/nbd15 %s\n", topLayer)
		fmt.Fprintf(logFile, "sudo mount /dev/nbd15 /tmp/test-mount\n")
		fmt.Fprintf(logFile, "ls -la /tmp/test-mount/\n")
		fmt.Fprintf(logFile, "sudo umount /tmp/test-mount\n")
		fmt.Fprintf(logFile, "sudo qemu-nbd --disconnect /dev/nbd15\n")
		fmt.Fprintf(logFile, "\n# Or mount the imagePath directly:\n")
		fmt.Fprintf(logFile, "sudo qemu-nbd --connect=/dev/nbd15 %s\n", newDisk.imagePath)
		fmt.Fprintf(logFile, "sudo mount /dev/nbd15 /tmp/test-mount\n")
		fmt.Fprintf(logFile, "ls -la /tmp/test-mount/\n")
		fmt.Fprintf(logFile, "sudo umount /tmp/test-mount\n")
		fmt.Fprintf(logFile, "sudo qemu-nbd --disconnect /dev/nbd15\n")
		fmt.Fprintf(logFile, "=== END MANUAL TEST INFO ===\n\n")
	}

	return newDisk, nil
}

// ForceRemoveFromPool forcefully removes a disk from the pool without unmounting
// This is useful for clearing stale pool entries
// It also updates the database to clear mount state
func (m *manager) ForceRemoveFromPool(diskId string) {
	if m.pool != nil {
		m.pool.ForceRemove(diskId)
	}

	// Also clear mount state in database to ensure fresh mounts
	if err := m.stateDB.UpdateMountState(diskId, false, ""); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to clear mount state for disk '%s': %v\n", diskId, err)
	}
}
