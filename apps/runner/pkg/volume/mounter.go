// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package volume

import "context"

// Mounter abstracts how a volume is mounted onto the host filesystem.
// The runner bind-mounts the resulting host path into the sandbox container.
// Implementations may use different backends (e.g. S3 FUSE, experimental, etc.).
type Mounter interface {
	// Mount ensures the volume is accessible at mountPath on the host.
	// volumeID is the backend-specific identifier (e.g. S3 bucket name).
	// The call must be idempotent — if already mounted, it returns nil.
	Mount(ctx context.Context, volumeID string, mountPath string) error

	// Unmount tears down the mount at the given path.
	Unmount(ctx context.Context, mountPath string) error

	// IsMounted reports whether mountPath is an active mountpoint.
	IsMounted(mountPath string) bool

	// WaitUntilReady blocks until the filesystem at mountPath is responsive
	// (i.e. Stat and ReadDir succeed). Implementations may return immediately
	// if the backend mounts synchronously.
	WaitUntilReady(ctx context.Context, mountPath string) error
}

// Volume describes a single volume to mount. It mirrors dto.VolumeDTO in a
// package-neutral shape so the volume package can stay free of cross-package
// dependencies.
//
// The Archil* fields are only populated when the sandbox uses the experimental
// in-container (Archil) backend; the host-side s3fuse path ignores them.
type Volume struct {
	VolumeID  string `json:"volumeId"`
	MountPath string `json:"mountPath"`
	Subpath   string `json:"subpath,omitempty"`

	// ReadOnly mounts the volume read-only inside this sandbox. Honored
	// by both backends; see Mounter implementations for how the flag is
	// applied (Docker bind mode for s3fuse, `archil mount --read-only`
	// for the experimental backend).
	ReadOnly bool `json:"readOnly,omitempty"`

	ArchilDisk       string `json:"archilDisk,omitempty"`
	ArchilRegion     string `json:"archilRegion,omitempty"`
	ArchilMountToken string `json:"archilMountToken,omitempty"`
}

// InContainerMounter is an optional extension implemented by mounters that
// perform their actual mount inside the sandbox container rather than on the
// runner host. When the runner sees a mounter satisfies this interface it:
//   - skips host-side mounting and bind creation
//   - appends ContainerBinds to the sandbox HostConfig.Binds
//   - appends ContainerEnv to the sandbox env
//
// The in-container daemon is expected to consume the injected env and perform
// the mount before user code runs.
type InContainerMounter interface {
	Mounter

	// ContainerBinds returns host->container bind strings that must be added
	// to HostConfig.Binds (e.g. the mount-s3 static binary mounted RO).
	// Independent of the volume list — applied on every sandbox that uses
	// this mounter.
	ContainerBinds() []string

	// ContainerEnv returns env vars that must be added to the sandbox (volume
	// spec + per-volume credentials + binary path). Implementations may
	// perform I/O and should honor the provided context. Returns nil when
	// there are no volumes to mount.
	ContainerEnv(ctx context.Context, volumes []Volume) ([]string, error)
}
