// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package dto

// VolumeDTO is the on-the-wire shape the control plane sends to the runner
// for each volume mount declared on a sandbox.
//
// The first three fields are populated for every backend. The Layered*
// fields are populated only when the sandbox uses the in-container layered
// backend (volumeBackend = "layered"); they are ignored by the host-side
// s3fuse path.
type VolumeDTO struct {
	VolumeId  string  `json:"volumeId"`
	MountPath string  `json:"mountPath"`
	Subpath   *string `json:"subpath,omitempty"`

	// ReadOnly mounts the volume read-only for this sandbox. It is a
	// per-mount attribute (not a per-volume one), so the same volume can
	// be mounted RW in one sandbox and RO in another. The s3fuse path
	// enforces it via the Docker bind mode (":ro"); the layered path
	// forwards it to the in-container mount binary's `--read-only` flag.
	ReadOnly bool `json:"readOnly,omitempty"`

	// LayeredDisk identifies the layered-volume disk to mount, in the
	// form "owner/disk-name" or "dsk-XXXXXXXXXXXXXXXX". Required when the
	// sandbox uses the in-container layered backend.
	LayeredDisk string `json:"layeredDisk,omitempty"`
	// LayeredRegion is the region the layered disk lives in
	// (e.g. "aws-us-east-1"). Required when LayeredDisk is set.
	LayeredRegion string `json:"layeredRegion,omitempty"`
	// LayeredMountToken is the per-(sandbox, volume) mount token used to
	// authenticate the mount inside the sandbox. The runner forwards it
	// to the daemon via env vars and never stores or logs it. Required
	// when LayeredDisk is set.
	LayeredMountToken string `json:"layeredMountToken,omitempty"`
}
