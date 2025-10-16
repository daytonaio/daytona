package sdisk

import (
	"errors"
	"time"
)

// DiskInfo contains information about a disk
type DiskInfo struct {
	Name         string    // Disk name
	SizeGB       int64     // Allocated size in GB
	ActualSizeGB int64     // Actual disk usage in GB
	Created      time.Time // Creation timestamp
	Modified     time.Time // Last modification timestamp
	IsMounted    bool      // Whether disk is currently mounted
	InS3         bool      // Whether disk exists in S3
	Checksum     string    // SHA256 checksum of disk
}

// Common errors
var (
	ErrDiskNotFound      = errors.New("disk not found")
	ErrDiskExists        = errors.New("disk already exists")
	ErrDiskInUse         = errors.New("disk is in use")
	ErrNotMounted        = errors.New("disk is not mounted")
	ErrAlreadyMounted    = errors.New("disk is already mounted")
	ErrInvalidSize       = errors.New("invalid disk size")
	ErrS3NotConfigured   = errors.New("S3 not configured")
	ErrInvalidConfig     = errors.New("invalid configuration")
	ErrQCOW2NotAvailable = errors.New("qemu-img or qemu-nbd not available")
	ErrMountFailed       = errors.New("failed to mount disk")
	ErrUnmountFailed     = errors.New("failed to unmount disk")
)
