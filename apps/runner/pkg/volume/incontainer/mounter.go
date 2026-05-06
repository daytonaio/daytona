// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

// Package incontainer implements a volume.Mounter that performs the mount
// inside the sandbox container instead of on the runner host. The runner's
// responsibility is limited to:
//
//  1. Bind-mounting the `archil` CLI binary into the container (read-only).
//  2. Injecting an env payload describing each volume — disk identifier,
//     region, and per-disk mount token — that the in-container daemon
//     consumes to invoke `archil mount` at sandbox start.
//
// Authentication is per-volume via Archil "disk tokens" (ARCHIL_MOUNT_TOKEN).
// Each token is scoped to a single Archil disk, so a sandbox only ever holds
// credentials for the disks it actually mounts. The runner itself does not
// need an Archil API key for this backend — it just shuttles the per-volume
// tokens that the control plane already issued.
package incontainer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/daytonaio/runner/pkg/volume"
)

// ArchilBinaryContainerPath is the well-known path the runner bind-mounts the
// `archil` CLI binary to inside every sandbox using the experimental
// in-container backend. The in-container daemon execs this exact path.
const ArchilBinaryContainerPath = "/usr/local/bin/daytona-archil"

// Env var names exchanged between the runner and the in-container daemon.
// Namespaced under DAYTONA_INCONTAINER_ so they don't collide with anything
// the user may set, and so secrets aren't dumped under the standard
// ARCHIL_MOUNT_TOKEN name in `env` listings before the daemon translates and
// scrubs them.
const (
	// EnvVolumesJSON is a JSON-encoded []Volume describing every volume to
	// mount, including its per-disk Archil mount token.
	EnvVolumesJSON = "DAYTONA_INCONTAINER_VOLUMES"
	// EnvArchilBinary is the absolute in-container path to the archil CLI
	// binary. Always set to ArchilBinaryContainerPath when this backend is
	// active.
	EnvArchilBinary = "DAYTONA_INCONTAINER_ARCHIL_BINARY"
)

// Config controls how the runner-side mounter prepares sandboxes for the
// experimental in-container (Archil) backend.
type Config struct {
	// ArchilBinaryHostPath is the host path to the `archil` CLI binary that
	// gets bind-mounted RO into each sandbox at ArchilBinaryContainerPath.
	// Required for this backend to operate.
	ArchilBinaryHostPath string
}

// Mounter is a volume.Mounter whose host-side methods are deliberate no-ops.
// The mount happens inside the sandbox container; the runner-side work is
// limited to bind-mount + env injection.
type Mounter struct {
	cfg Config
}

func NewMounter(cfg Config) *Mounter {
	return &Mounter{cfg: cfg}
}

// Host-side lifecycle — all no-ops. The mount lives inside the container
// and is torn down naturally when the container exits.

func (m *Mounter) Mount(_ context.Context, _ string, _ string) error { return nil }
func (m *Mounter) Unmount(_ context.Context, _ string) error         { return nil }
func (m *Mounter) IsMounted(_ string) bool                           { return false }
func (m *Mounter) WaitUntilReady(_ context.Context, _ string) error  { return nil }

// ContainerBinds returns the RO binds every sandbox using this backend
// needs regardless of volume count: just the `archil` CLI binary, mounted
// at ArchilBinaryContainerPath.
func (m *Mounter) ContainerBinds() []string {
	if m.cfg.ArchilBinaryHostPath == "" {
		return nil
	}
	return []string{fmt.Sprintf("%s:%s:ro", m.cfg.ArchilBinaryHostPath, ArchilBinaryContainerPath)}
}

// ContainerEnv serializes the volume list (including per-volume mount tokens)
// and the in-container archil binary path into env vars for the daemon to
// consume at sandbox start. Returns nil when there are no volumes to mount.
func (m *Mounter) ContainerEnv(_ context.Context, volumes []volume.Volume) ([]string, error) {
	if len(volumes) == 0 {
		return nil, nil
	}

	for i, v := range volumes {
		if v.ArchilDisk == "" {
			return nil, fmt.Errorf("volume %d (%q) is missing archilDisk; the experimental backend requires Archil-shaped volumes", i, v.VolumeID)
		}
		if v.ArchilRegion == "" {
			return nil, fmt.Errorf("volume %d (%q) is missing archilRegion", i, v.VolumeID)
		}
		if v.ArchilMountToken == "" {
			return nil, fmt.Errorf("volume %d (%q) is missing archilMountToken", i, v.VolumeID)
		}
	}

	volumesJSON, err := json.Marshal(volumes)
	if err != nil {
		return nil, fmt.Errorf("marshal volumes: %w", err)
	}

	return []string{
		EnvVolumesJSON + "=" + string(volumesJSON),
		EnvArchilBinary + "=" + ArchilBinaryContainerPath,
	}, nil
}

// Compile-time check that Mounter satisfies both interfaces.
var (
	_ volume.Mounter            = (*Mounter)(nil)
	_ volume.InContainerMounter = (*Mounter)(nil)
)
