// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

// Package incontainer implements a volume.Mounter that performs the mount
// inside the sandbox container instead of on the runner host. The runner's
// responsibility is limited to:
//
//  1. Bind-mounting the layered-volume mount binary into the container
//     (read-only).
//  2. Injecting an env payload describing each volume — disk identifier,
//     region, and per-(sandbox, volume) mount token — that the in-container
//     daemon consumes to invoke the mount binary at sandbox start.
//
// Authentication is per-volume via per-(sandbox, volume) tokens. Each token
// is scoped to a single layered disk, so a sandbox only ever holds
// credentials for the disks it actually mounts. The runner itself does not
// need a layered control-plane API key for this backend — it just shuttles
// the per-volume tokens that the control plane already issued.
package incontainer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/daytonaio/runner/pkg/volume"
)

// LayeredBinaryContainerPath is the well-known path the runner bind-mounts
// the layered-volume mount CLI to inside every sandbox using the
// in-container layered backend. The in-container daemon execs this exact
// path.
const LayeredBinaryContainerPath = "/usr/local/bin/daytona-layered"

// Env var names exchanged between the runner and the in-container daemon.
// Namespaced under DAYTONA_INCONTAINER_ so they don't collide with anything
// the user may set, and so secrets aren't dumped under the third-party CLI's
// own env var name in `env` listings before the daemon translates and
// scrubs them.
const (
	// EnvVolumesJSON is a JSON-encoded []Volume describing every volume to
	// mount, including its per-(sandbox, volume) mount token.
	EnvVolumesJSON = "DAYTONA_INCONTAINER_VOLUMES"
	// EnvLayeredBinary is the absolute in-container path to the layered
	// mount CLI. Always set to LayeredBinaryContainerPath when this
	// backend is active.
	EnvLayeredBinary = "DAYTONA_INCONTAINER_LAYERED_BINARY"
)

// Config controls how the runner-side mounter prepares sandboxes for the
// in-container layered backend.
type Config struct {
	// LayeredBinaryHostPath is the host path to the layered mount CLI
	// binary that gets bind-mounted RO into each sandbox at
	// LayeredBinaryContainerPath. Required for this backend to operate.
	LayeredBinaryHostPath string
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
// needs regardless of volume count: just the layered mount CLI binary,
// mounted at LayeredBinaryContainerPath.
func (m *Mounter) ContainerBinds() []string {
	if m.cfg.LayeredBinaryHostPath == "" {
		return nil
	}
	return []string{fmt.Sprintf("%s:%s:ro", m.cfg.LayeredBinaryHostPath, LayeredBinaryContainerPath)}
}

// ContainerEnv serializes the volume list (including per-volume mount tokens)
// and the in-container layered binary path into env vars for the daemon to
// consume at sandbox start. Returns nil when there are no volumes to mount.
func (m *Mounter) ContainerEnv(_ context.Context, volumes []volume.Volume) ([]string, error) {
	if len(volumes) == 0 {
		return nil, nil
	}

	for i, v := range volumes {
		if v.LayeredDisk == "" {
			return nil, fmt.Errorf("volume %d (%q) is missing layeredDisk; the layered backend requires layered-shaped volumes", i, v.VolumeID)
		}
		if v.LayeredRegion == "" {
			return nil, fmt.Errorf("volume %d (%q) is missing layeredRegion", i, v.VolumeID)
		}
		if v.LayeredMountToken == "" {
			return nil, fmt.Errorf("volume %d (%q) is missing layeredMountToken", i, v.VolumeID)
		}
	}

	volumesJSON, err := json.Marshal(volumes)
	if err != nil {
		return nil, fmt.Errorf("marshal volumes: %w", err)
	}

	return []string{
		EnvVolumesJSON + "=" + string(volumesJSON),
		EnvLayeredBinary + "=" + LayeredBinaryContainerPath,
	}, nil
}

// Compile-time check that Mounter satisfies both interfaces.
var (
	_ volume.Mounter            = (*Mounter)(nil)
	_ volume.InContainerMounter = (*Mounter)(nil)
)
