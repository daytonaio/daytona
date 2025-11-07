package sdisk

import "context"

// DiskManager is the main interface for managing disks
type DiskManager interface {
	// Create creates a new disk with the specified name and size in GB
	Create(ctx context.Context, name string, sizeGB int) (Disk, error)

	// Open opens an existing local disk
	Open(ctx context.Context, name string) (Disk, error)

	// Pull downloads a disk from S3 and opens it locally
	Pull(ctx context.Context, name string) (Disk, error)

	// List returns information about all local disks
	List(ctx context.Context) ([]DiskInfo, error)

	// Delete removes a local disk
	Delete(ctx context.Context, name string) error

	// PoolStats returns statistics about the disk pool (nil if pooling disabled)
	PoolStats() *PoolStats

	// Close closes the manager and releases resources
	Close() error

	// CleanupUnusedLayers removes cached layers with zero references
	CleanupUnusedLayers(ctx context.Context) (int, error)

	// Fork creates a new disk that shares all existing layers of the source disk
	// Both disks will have independent write layers for independent operation
	Fork(ctx context.Context, sourceDiskName, newDiskName string) (Disk, error)
}

// Disk represents a managed disk volume
type Disk interface {
	// Name returns the disk name
	Name() string

	// Size returns the disk size in GB
	Size() int64

	// Mount mounts the disk and returns the mount path
	Mount(ctx context.Context) (string, error)

	// Unmount unmounts the disk
	Unmount(ctx context.Context) error

	// IsMounted returns whether the disk is currently mounted
	IsMounted() bool

	// MountPath returns the current mount path (empty if not mounted)
	MountPath() string

	// Push uploads the disk to S3
	Push(ctx context.Context) error

	// Info returns disk information
	Info() DiskInfo

	// Resize changes the disk size
	Resize(ctx context.Context, newSizeGB int) error

	// Close closes the disk and releases resources
	Close() error
}
