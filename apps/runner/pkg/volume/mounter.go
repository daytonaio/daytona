// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package volume

import "context"

// Mounter abstracts how a volume is mounted onto the host filesystem; the
// runner bind-mounts the resulting host path into the sandbox. Implementations
// may use different backends (S3 FUSE, layered, etc.).
type Mounter interface {
	// Mount ensures the volume is accessible at mountPath on the host.
	// volumeID is backend-specific (e.g. an S3 bucket name). Idempotent:
	// returns nil if already mounted.
	Mount(ctx context.Context, volumeID string, mountPath string) error

	// Unmount tears down the mount at the given path.
	Unmount(ctx context.Context, mountPath string) error

	// IsMounted reports whether mountPath is an active mountpoint.
	IsMounted(mountPath string) bool

	// WaitUntilReady blocks until the filesystem at mountPath is responsive
	// (Stat and ReadDir succeed). May return immediately for synchronous
	// backends.
	WaitUntilReady(ctx context.Context, mountPath string) error
}

// Volume describes a single volume to mount. It mirrors dto.VolumeDTO in a
// package-neutral shape to keep this package dependency-free. The Layered*
// fields are populated only for the layered in-container backend.
type Volume struct {
	VolumeID  string `json:"volumeId"`
	MountPath string `json:"mountPath"`
	Subpath   string `json:"subpath,omitempty"`

	// ReadOnly mounts the volume read-only in this sandbox. Honored by both
	// backends (Docker bind mode for s3fuse, the CLI's `--read-only` flag for
	// layered).
	ReadOnly bool `json:"readOnly,omitempty"`

	LayeredDisk       string `json:"layeredDisk,omitempty"`
	LayeredRegion     string `json:"layeredRegion,omitempty"`
	LayeredMountToken string `json:"layeredMountToken,omitempty"`
}

// InContainerMounter is an optional extension for mounters that mount inside
// the sandbox rather than on the host. When a mounter satisfies it, the runner
// skips host-side mounting and instead appends ContainerBinds to
// HostConfig.Binds and ContainerEnv to the sandbox env; the in-container daemon
// consumes that env and mounts before user code runs.
type InContainerMounter interface {
	Mounter

	// ContainerBinds returns host->container bind strings for HostConfig.Binds
	// (e.g. the mount binary mounted RO), independent of the volume list.
	ContainerBinds() []string

	// ContainerEnv returns env vars for the sandbox (volume spec + per-volume
	// credentials + binary path). May perform I/O and should honor ctx.
	// Returns nil when there are no volumes to mount.
	ContainerEnv(ctx context.Context, volumes []Volume) ([]string, error)
}
