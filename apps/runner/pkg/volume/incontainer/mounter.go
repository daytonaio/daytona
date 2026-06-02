// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

// Package incontainer implements a volume.Mounter that mounts inside the
// sandbox container rather than on the runner host. The runner only:
//
//  1. Bind-mounts the layered mount binary into the container (read-only).
//  2. Injects an env payload (disk, region, per-(sandbox, volume) token) that
//     the in-container daemon uses to invoke the binary at sandbox start.
//
// Each token is scoped to a single disk, so a sandbox holds credentials only
// for the disks it mounts. The runner needs no layered API key — it just
// shuttles the tokens the control plane already issued.
package incontainer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/daytonaio/runner/pkg/volume"
)

// LayeredBinaryContainerPath is the path the runner bind-mounts the layered
// mount CLI to, and the path the in-container daemon execs.
const LayeredBinaryContainerPath = "/usr/local/bin/daytona-layered"

// Env var names exchanged with the in-container daemon. Namespaced under
// DAYTONA_INCONTAINER_ to avoid colliding with user env and to keep secrets
// out of the third-party CLI's own env var name until the daemon scrubs them.
const (
	// EnvVolumesJSON is a JSON-encoded []Volume (including per-volume tokens).
	EnvVolumesJSON = "DAYTONA_INCONTAINER_VOLUMES"
	// EnvLayeredBinary is the in-container path to the layered mount CLI,
	// always LayeredBinaryContainerPath when this backend is active.
	EnvLayeredBinary = "DAYTONA_INCONTAINER_LAYERED_BINARY"
)

// Config controls how the runner-side mounter prepares sandboxes for the
// in-container layered backend.
type Config struct {
	// LayeredBinaryHostPath is the host path to the layered mount CLI,
	// bind-mounted RO into each sandbox at LayeredBinaryContainerPath.
	// Required for this backend.
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

// Host-side lifecycle — all no-ops. The mount lives inside the container and
// is torn down when the container exits.

func (m *Mounter) Mount(_ context.Context, _ string, _ string) error { return nil }
func (m *Mounter) Unmount(_ context.Context, _ string) error         { return nil }
func (m *Mounter) IsMounted(_ string) bool                           { return false }
func (m *Mounter) WaitUntilReady(_ context.Context, _ string) error  { return nil }

// ContainerBinds returns the RO bind for the layered mount CLI binary,
// needed by every sandbox on this backend regardless of volume count.
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
