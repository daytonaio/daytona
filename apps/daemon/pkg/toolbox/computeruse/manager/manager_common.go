// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package manager

import (
	"sync"

	"github.com/daytonaio/daemon/pkg/toolbox/computeruse"
	"github.com/hashicorp/go-plugin"
)

type pluginRef struct {
	// mu serializes plugin lifecycle: spawn (getOrSpawn) and teardown
	// (KillComputerUse). It guards all fields below.
	mu     sync.Mutex
	client *plugin.Client
	impl   computeruse.IComputerUse
	path   string
}

var ComputerUseHandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "DAYTONA_COMPUTER_USE_PLUGIN",
	MagicCookieValue: "daytona_computer_use",
}

var computerUse = &pluginRef{}

// spawnFunc starts the plugin process and returns the managed client, the
// dispensed impl, and the plugin base path.
type spawnFunc func() (*plugin.Client, computeruse.IComputerUse, string, error)

// getOrSpawn returns the cached plugin impl, or runs spawn under the manager
// lock and caches its result. Exactly one concurrent caller executes spawn;
// the rest block on the lock and receive the cached instance, so no code path
// can start a second plugin process. A failed spawn caches nothing, leaving
// the next caller free to retry.
func getOrSpawn(spawn spawnFunc) (computeruse.IComputerUse, error) {
	computerUse.mu.Lock()
	defer computerUse.mu.Unlock()

	if computerUse.impl != nil {
		return computerUse.impl, nil
	}

	client, impl, basePath, err := spawn()
	if err != nil {
		return nil, err
	}

	computerUse.client = client
	computerUse.impl = impl
	computerUse.path = basePath
	return impl, nil
}

// ComputerUseError represents a computer-use plugin error with context
type ComputerUseError struct {
	Type    string // "dependency", "system", "plugin"
	Message string
	Details string
}

func (e *ComputerUseError) Error() string {
	return e.Message
}

// KillComputerUse terminates the plugin client and clears the cached impl.
// Used by the daemon's /computeruse/stop HTTP handler and shutdown path.
//
// It takes the manager lock, so a kill racing an in-flight spawn WAITS for
// the spawn to finish and then terminates the freshly spawned client (waiting
// is simpler than cancellation and cannot leak the child either way).
func KillComputerUse() {
	computerUse.mu.Lock()
	defer computerUse.mu.Unlock()

	if computerUse.client != nil {
		computerUse.client.Kill()
	}
	computerUse.client = nil
	computerUse.impl = nil
	computerUse.path = ""
}
