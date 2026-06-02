// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package dto

// VolumeDTO is the on-the-wire shape the control plane sends per volume mount.
// The Layered* fields are populated only for the in-container layered backend
// (volumeBackend = "layered") and ignored by the s3fuse path.
type VolumeDTO struct {
	VolumeId  string  `json:"volumeId"`
	MountPath string  `json:"mountPath"`
	Subpath   *string `json:"subpath,omitempty"`

	// ReadOnly mounts the volume read-only for this sandbox. It is per-mount,
	// not per-volume, so the same volume can be RW in one sandbox and RO in
	// another. s3fuse enforces it via the Docker bind mode (":ro"); the
	// layered path forwards it to the mount binary's `--read-only` flag.
	ReadOnly bool `json:"readOnly,omitempty"`

	// LayeredDisk identifies the disk to mount, as "owner/disk-name" or
	// "dsk-XXXXXXXXXXXXXXXX". Required for the layered backend.
	LayeredDisk string `json:"layeredDisk,omitempty"`
	// LayeredRegion is the disk's region (e.g. "aws-us-east-1"). Required
	// when LayeredDisk is set.
	LayeredRegion string `json:"layeredRegion,omitempty"`
	// LayeredMountToken authenticates the mount inside the sandbox
	// (per-(sandbox, volume)). The runner forwards it to the daemon via env
	// and never stores or logs it. Required when LayeredDisk is set.
	LayeredMountToken string `json:"layeredMountToken,omitempty"`
}
