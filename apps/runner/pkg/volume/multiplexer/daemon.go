package multiplexer

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/daytonaio/runner/pkg/volume"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

// MultiplexerDaemon manages multiple volumes through a single FUSE mount
type MultiplexerDaemon struct {
	mountPoint string
	volumes    sync.Map // volumeId -> *VolumeEntry
	server     *fuse.Server
	root       *rootNode
	logger     *slog.Logger

	// Statistics
	stats *StatsTracker

	// Configuration
	cacheDir string
	maxCache int64
}

// VolumeEntry represents a registered volume
type VolumeEntry struct {
	ID       string
	Provider volume.Provider
	RefCount atomic.Int32 // Number of active bind mounts
	Cache    *VolumeCache
	ReadOnly bool
}

// rootNode is the FUSE root directory that routes to volumes
type rootNode struct {
	fs.Inode
	daemon *MultiplexerDaemon
}

// volumeNode represents a volume subdirectory
type volumeNode struct {
	fs.Inode
	volumeID string
	daemon   *MultiplexerDaemon
}

// fileNode represents a file within a volume
type fileNode struct {
	fs.Inode
	volumeID string
	path     string
	daemon   *MultiplexerDaemon
	mu       sync.Mutex
}

// NewMultiplexerDaemon creates a new multiplexer instance
func NewMultiplexerDaemon(mountPoint string, cacheDir string, logger *slog.Logger) *MultiplexerDaemon {
	return &MultiplexerDaemon{
		mountPoint: mountPoint,
		cacheDir:   cacheDir,
		maxCache:   10 * 1024 * 1024 * 1024, // 10GB default
		logger:     logger,
		stats:      NewStatsTracker(),
	}
}

// Start mounts the FUSE filesystem and starts serving requests
func (d *MultiplexerDaemon) Start(ctx context.Context) error {
	// Create mount point if it doesn't exist
	if err := os.MkdirAll(d.mountPoint, 0755); err != nil {
		return fmt.Errorf("failed to create mount point: %w", err)
	}

	// Create cache directory
	if err := os.MkdirAll(d.cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Create root node
	d.root = &rootNode{daemon: d}

	// Mount options
	opts := &fs.Options{
		MountOptions: fuse.MountOptions{
			AllowOther: true,
			Debug:      false,
			FsName:     "daytona-volumes",
			Name:       "daytona-volume-multiplexer",
		},
		UID: uint32(os.Getuid()),
		GID: uint32(os.Getgid()),
	}

	// Create and start server
	server, err := fs.Mount(d.mountPoint, d.root, opts)
	if err != nil {
		return fmt.Errorf("failed to mount filesystem: %w", err)
	}
	d.server = server

	d.logger.Info("FUSE multiplexer started", "mountPoint", d.mountPoint)

	// Wait for unmount or context cancellation
	go func() {
		<-ctx.Done()
		d.Stop()
	}()

	// Serve requests
	d.server.Wait()
	return nil
}

// Stop unmounts the filesystem and cleans up
func (d *MultiplexerDaemon) Stop() error {
	if d.server != nil {
		err := d.server.Unmount()
		d.server = nil
		return err
	}
	return nil
}

// RegisterVolume adds a volume to the multiplexer
func (d *MultiplexerDaemon) RegisterVolume(ctx context.Context, volumeID string, config volume.ProviderConfig, readOnly bool) error {
	// Check if already registered
	if _, exists := d.volumes.Load(volumeID); exists {
		d.logger.Debug("Volume already registered", "volumeID", volumeID)
		return nil
	}

	// Create provider based on type
	provider, err := createProvider(config.Type)
	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}

	// Connect to storage backend
	if err := provider.Connect(ctx, config); err != nil {
		return fmt.Errorf("failed to connect provider: %w", err)
	}

	// Create cache for this volume
	cache := NewVolumeCache(path.Join(d.cacheDir, volumeID), 1024*1024*1024) // 1GB per volume

	// Register volume
	entry := &VolumeEntry{
		ID:       volumeID,
		Provider: provider,
		Cache:    cache,
		ReadOnly: readOnly,
	}
	d.volumes.Store(volumeID, entry)

	d.logger.Info("Volume registered", "volumeID", volumeID, "type", config.Type)
	d.stats.VolumeRegistered(volumeID)

	return nil
}

// UnregisterVolume removes a volume from the multiplexer
func (d *MultiplexerDaemon) UnregisterVolume(ctx context.Context, volumeID string) error {
	value, exists := d.volumes.Load(volumeID)
	if !exists {
		return fmt.Errorf("volume not found: %s", volumeID)
	}

	entry := value.(*VolumeEntry)

	// Check reference count
	if entry.RefCount.Load() > 0 {
		return fmt.Errorf("volume %s still in use (refcount: %d)", volumeID, entry.RefCount.Load())
	}

	// Close provider
	if err := entry.Provider.Close(); err != nil {
		d.logger.Warn("Failed to close provider", "volumeID", volumeID, "error", err)
	}

	// Clean up cache
	if err := entry.Cache.Clear(); err != nil {
		d.logger.Warn("Failed to clear cache", "volumeID", volumeID, "error", err)
	}

	// Remove from map
	d.volumes.Delete(volumeID)

	d.logger.Info("Volume unregistered", "volumeID", volumeID)
	d.stats.VolumeUnregistered(volumeID)

	return nil
}

// IncrementRefCount increases the reference count for a volume
func (d *MultiplexerDaemon) IncrementRefCount(volumeID string) error {
	value, exists := d.volumes.Load(volumeID)
	if !exists {
		return fmt.Errorf("volume not found: %s", volumeID)
	}

	entry := value.(*VolumeEntry)
	entry.RefCount.Add(1)
	return nil
}

// DecrementRefCount decreases the reference count for a volume
func (d *MultiplexerDaemon) DecrementRefCount(volumeID string) error {
	value, exists := d.volumes.Load(volumeID)
	if !exists {
		return fmt.Errorf("volume not found: %s", volumeID)
	}

	entry := value.(*VolumeEntry)
	if entry.RefCount.Add(-1) < 0 {
		entry.RefCount.Store(0) // Prevent negative counts
	}
	return nil
}

// GetStats returns current daemon statistics
func (d *MultiplexerDaemon) GetStats() *DaemonStats {
	return d.stats.GetStats()
}

// FUSE Operations Implementation

// Readdir implements the root directory listing
func (r *rootNode) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	entries := []fuse.DirEntry{
		{Name: ".", Mode: syscall.S_IFDIR},
		{Name: "..", Mode: syscall.S_IFDIR},
	}

	// List all registered volumes
	r.daemon.volumes.Range(func(key, value interface{}) bool {
		volumeID := key.(string)
		entries = append(entries, fuse.DirEntry{
			Name: volumeID,
			Mode: syscall.S_IFDIR,
		})
		return true
	})

	return fs.NewListDirStream(entries), 0
}

// Lookup finds a volume directory
func (r *rootNode) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	// Check if volume exists
	if _, exists := r.daemon.volumes.Load(name); !exists {
		return nil, syscall.ENOENT
	}

	// Create volume node
	node := &volumeNode{
		volumeID: name,
		daemon:   r.daemon,
	}

	// Create inode
	child := r.NewPersistentInode(ctx, node, fs.StableAttr{Mode: syscall.S_IFDIR})
	return child, 0
}

// Getattr returns attributes for the root directory
func (r *rootNode) Getattr(ctx context.Context, f fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Mode = syscall.S_IFDIR | 0755
	out.Uid = uint32(os.Getuid())
	out.Gid = uint32(os.Getgid())
	return 0
}

// Volume node operations

// Readdir lists files in a volume
func (v *volumeNode) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	entry, exists := v.daemon.volumes.Load(v.volumeID)
	if !exists {
		return nil, syscall.ENOENT
	}

	provider := entry.(*VolumeEntry).Provider

	// List directory from provider
	files, err := provider.ListDir(ctx, "/")
	if err != nil {
		v.daemon.logger.Error("Failed to list directory", "volume", v.volumeID, "error", err)
		return nil, syscall.EIO
	}

	entries := []fuse.DirEntry{
		{Name: ".", Mode: syscall.S_IFDIR},
		{Name: "..", Mode: syscall.S_IFDIR},
	}

	for _, file := range files {
		mode := uint32(syscall.S_IFREG)
		if file.IsDir {
			mode = syscall.S_IFDIR
		}
		entries = append(entries, fuse.DirEntry{
			Name: file.Name,
			Mode: mode,
		})
	}

	v.daemon.stats.Operation("readdir", v.volumeID, 0)
	return fs.NewListDirStream(entries), 0
}

// Lookup finds a file or subdirectory in a volume
func (v *volumeNode) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	entry, exists := v.daemon.volumes.Load(v.volumeID)
	if !exists {
		return nil, syscall.ENOENT
	}

	provider := entry.(*VolumeEntry).Provider

	// Get file info from provider
	info, err := provider.GetFileInfo(ctx, name)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, syscall.ENOENT
		}
		v.daemon.logger.Error("Failed to get file info", "volume", v.volumeID, "path", name, "error", err)
		return nil, syscall.EIO
	}

	// Set attributes
	fillAttr(&out.Attr, &info)

	if info.IsDir {
		// Create subdirectory node
		node := &volumeNode{
			volumeID: v.volumeID,
			daemon:   v.daemon,
		}
		child := v.NewPersistentInode(ctx, node, fs.StableAttr{Mode: syscall.S_IFDIR})
		return child, 0
	} else {
		// Create file node
		node := &fileNode{
			volumeID: v.volumeID,
			path:     name,
			daemon:   v.daemon,
		}
		child := v.NewPersistentInode(ctx, node, fs.StableAttr{Mode: syscall.S_IFREG})
		return child, 0
	}
}

// Getattr returns attributes for a volume directory
func (v *volumeNode) Getattr(ctx context.Context, f fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Mode = syscall.S_IFDIR | 0755
	out.Uid = uint32(os.Getuid())
	out.Gid = uint32(os.Getgid())
	return 0
}

// Create creates a new file
func (v *volumeNode) Create(ctx context.Context, name string, flags uint32, mode uint32, out *fuse.EntryOut) (*fs.Inode, fs.FileHandle, uint32, syscall.Errno) {
	entry, exists := v.daemon.volumes.Load(v.volumeID)
	if !exists {
		return nil, nil, 0, syscall.ENOENT
	}

	volumeEntry := entry.(*VolumeEntry)
	if volumeEntry.ReadOnly {
		return nil, nil, 0, syscall.EROFS
	}

	// Create empty file
	err := volumeEntry.Provider.WriteFile(ctx, name, []byte{}, 0)
	if err != nil {
		v.daemon.logger.Error("Failed to create file", "volume", v.volumeID, "path", name, "error", err)
		return nil, nil, 0, syscall.EIO
	}

	// Create file node
	node := &fileNode{
		volumeID: v.volumeID,
		path:     name,
		daemon:   v.daemon,
	}

	child := v.NewPersistentInode(ctx, node, fs.StableAttr{Mode: syscall.S_IFREG})
	fh := &fileHandle{
		volumeID: v.volumeID,
		path:     name,
		daemon:   v.daemon,
		flags:    flags,
	}

	v.daemon.stats.Operation("create", v.volumeID, 0)
	return child, fh, fuse.FOPEN_DIRECT_IO, 0
}

// Mkdir creates a new directory
func (v *volumeNode) Mkdir(ctx context.Context, name string, mode uint32, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	entry, exists := v.daemon.volumes.Load(v.volumeID)
	if !exists {
		return nil, syscall.ENOENT
	}

	volumeEntry := entry.(*VolumeEntry)
	if volumeEntry.ReadOnly {
		return nil, syscall.EROFS
	}

	// Create directory
	err := volumeEntry.Provider.CreateDir(ctx, name)
	if err != nil {
		v.daemon.logger.Error("Failed to create directory", "volume", v.volumeID, "path", name, "error", err)
		return nil, syscall.EIO
	}

	// Create directory node
	node := &volumeNode{
		volumeID: v.volumeID,
		daemon:   v.daemon,
	}

	child := v.NewPersistentInode(ctx, node, fs.StableAttr{Mode: syscall.S_IFDIR})
	v.daemon.stats.Operation("mkdir", v.volumeID, 0)
	return child, 0
}

// File operations

// Open opens a file for reading/writing
func (f *fileNode) Open(ctx context.Context, flags uint32) (fs.FileHandle, uint32, syscall.Errno) {
	fh := &fileHandle{
		volumeID: f.volumeID,
		path:     f.path,
		daemon:   f.daemon,
		flags:    flags,
	}

	f.daemon.stats.Operation("open", f.volumeID, 0)
	return fh, fuse.FOPEN_DIRECT_IO, 0
}

// Getattr returns file attributes
func (f *fileNode) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	entry, exists := f.daemon.volumes.Load(f.volumeID)
	if !exists {
		return syscall.ENOENT
	}

	provider := entry.(*VolumeEntry).Provider

	// Get file info from provider
	info, err := provider.GetFileInfo(ctx, f.path)
	if err != nil {
		if os.IsNotExist(err) {
			return syscall.ENOENT
		}
		f.daemon.logger.Error("Failed to get file info", "volume", f.volumeID, "path", f.path, "error", err)
		return syscall.EIO
	}

	fillAttr(&out.Attr, &info)
	return 0
}

// Setattr sets file attributes
func (f *fileNode) Setattr(ctx context.Context, fh fs.FileHandle, in *fuse.SetAttrIn, out *fuse.AttrOut) syscall.Errno {
	entry, exists := f.daemon.volumes.Load(f.volumeID)
	if !exists {
		return syscall.ENOENT
	}

	volumeEntry := entry.(*VolumeEntry)
	if volumeEntry.ReadOnly {
		return syscall.EROFS
	}

	// Handle size changes (truncate)
	if in.Valid&fuse.FATTR_SIZE != 0 {
		err := volumeEntry.Provider.Truncate(ctx, f.path, int64(in.Size))
		if err != nil {
			f.daemon.logger.Error("Failed to truncate file", "volume", f.volumeID, "path", f.path, "error", err)
			return syscall.EIO
		}
	}

	// Get updated attributes
	return f.Getattr(ctx, fh, out)
}

// fileHandle implements file operations
type fileHandle struct {
	volumeID string
	path     string
	daemon   *MultiplexerDaemon
	flags    uint32
}

// Read reads data from a file
func (fh *fileHandle) Read(ctx context.Context, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	entry, exists := fh.daemon.volumes.Load(fh.volumeID)
	if !exists {
		return nil, syscall.ENOENT
	}

	volumeEntry := entry.(*VolumeEntry)

	// Try cache first
	if data, ok := volumeEntry.Cache.Get(fh.path, off, len(dest)); ok {
		fh.daemon.stats.CacheHit(fh.volumeID, len(data))
		return fuse.ReadResultData(data), 0
	}

	// Read from provider
	data, err := volumeEntry.Provider.ReadFile(ctx, fh.path, off, len(dest))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, syscall.ENOENT
		}
		fh.daemon.logger.Error("Failed to read file", "volume", fh.volumeID, "path", fh.path, "error", err)
		return nil, syscall.EIO
	}

	// Update cache
	volumeEntry.Cache.Put(fh.path, off, data)
	fh.daemon.stats.Operation("read", fh.volumeID, len(data))

	return fuse.ReadResultData(data), 0
}

// Write writes data to a file
func (fh *fileHandle) Write(ctx context.Context, data []byte, off int64) (uint32, syscall.Errno) {
	entry, exists := fh.daemon.volumes.Load(fh.volumeID)
	if !exists {
		return 0, syscall.ENOENT
	}

	volumeEntry := entry.(*VolumeEntry)
	if volumeEntry.ReadOnly {
		return 0, syscall.EROFS
	}

	// Write to provider
	err := volumeEntry.Provider.WriteFile(ctx, fh.path, data, off)
	if err != nil {
		fh.daemon.logger.Error("Failed to write file", "volume", fh.volumeID, "path", fh.path, "error", err)
		return 0, syscall.EIO
	}

	// Invalidate cache for this file
	volumeEntry.Cache.Invalidate(fh.path)
	fh.daemon.stats.Operation("write", fh.volumeID, len(data))

	return uint32(len(data)), 0
}

// Release closes the file handle
func (fh *fileHandle) Release(ctx context.Context) syscall.Errno {
	// Nothing to do for now
	return 0
}

// Helper functions

func fillAttr(attr *fuse.Attr, info *volume.FileInfo) {
	attr.Size = uint64(info.Size)
	attr.Mode = uint32(info.Mode)
	attr.Mtime = uint64(info.ModTime.Unix())
	attr.Uid = uint32(os.Getuid())
	attr.Gid = uint32(os.Getgid())

	if info.IsDir {
		attr.Mode |= syscall.S_IFDIR
	} else {
		attr.Mode |= syscall.S_IFREG
	}
}

func createProvider(providerType string) (volume.Provider, error) {
	return volume.CreateProvider(providerType)
}
