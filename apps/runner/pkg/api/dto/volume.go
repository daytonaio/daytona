// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package dto

// VolumeDTO is the on-the-wire shape the control plane sends to the runner
// for each volume mount declared on a sandbox.
//
// The first three fields are populated for every backend. The Archil*
// fields are populated only when the sandbox uses the experimental
// in-container backend (volumeBackend = "experimental"); they are ignored
// by the host-side s3fuse path.
type VolumeDTO struct {
	VolumeId  string  `json:"volumeId"`
	MountPath string  `json:"mountPath"`
	Subpath   *string `json:"subpath,omitempty"`

	// ReadOnly mounts the volume read-only for this sandbox. It is a
	// per-mount attribute (not a per-volume one), so the same volume can
	// be mounted RW in one sandbox and RO in another. The s3fuse path
	// enforces it via the Docker bind mode (":ro"); the experimental
	// path forwards it to `archil mount --read-only`.
	ReadOnly bool `json:"readOnly,omitempty"`

	// ArchilDisk identifies the Archil disk to mount, in the form
	// "owner/disk-name" or "dsk-XXXXXXXXXXXXXXXX". Required when the
	// sandbox uses the experimental in-container backend.
	ArchilDisk string `json:"archilDisk,omitempty"`
	// ArchilRegion is the Archil region the disk lives in
	// (e.g. "aws-us-east-1"). Required when ArchilDisk is set.
	ArchilRegion string `json:"archilRegion,omitempty"`
	// ArchilMountToken is the per-disk mount token used as ARCHIL_MOUNT_TOKEN
	// inside the sandbox. The runner forwards it to the daemon via env vars
	// and never stores or logs it. Required when ArchilDisk is set.
	ArchilMountToken string `json:"archilMountToken,omitempty"`
}
