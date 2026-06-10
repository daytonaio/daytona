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
}

var ComputerUseHandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "DAYTONA_COMPUTER_USE_PLUGIN",
	MagicCookieValue: "daytona_computer_use",
}

var computerUse = &pluginRef{}

// spawnFunc starts the plugin process and returns the managed client and the
// dispensed impl.
type spawnFunc func() (*plugin.Client, computeruse.IComputerUse, error)

// getOrSpawn returns the cached plugin impl, or runs spawn under the manager
// lock, caches its result, and announces the fresh impl via publish (when
// non-nil). Exactly one concurrent caller executes spawn; the rest block on
// the lock and receive the cached instance, so no code path can start a
// second plugin process. A failed spawn caches and publishes nothing, leaving
// the next caller free to retry.
//
// publish runs while the manager lock is held, so an external cache (the
// daemon's LazyComputerUse) is updated atomically with the manager state:
// the manager is the single writer of both and they can never diverge.
func getOrSpawn(spawn spawnFunc, publish func(computeruse.IComputerUse)) (computeruse.IComputerUse, error) {
	computerUse.mu.Lock()
	defer computerUse.mu.Unlock()

	if computerUse.impl != nil {
		return computerUse.impl, nil
	}

	client, impl, err := spawn()
	if err != nil {
		return nil, err
	}

	computerUse.client = client
	computerUse.impl = impl
	if publish != nil {
		publish(impl)
	}
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

// KillComputerUse terminates the plugin client, clears the cached impl, and
// announces the teardown by calling publish(nil) (when publish is non-nil).
// Used by the Windows daemon's /computeruse/stop HTTP handler and shutdown
// path; the Linux daemon keeps the plugin alive for the process lifetime.
//
// It takes the manager lock, so a kill racing an in-flight spawn WAITS for
// the spawn to finish and then terminates the freshly spawned client (waiting
// is simpler than cancellation and cannot leak the child either way). When
// nothing was spawned it is a cheap no-op, so callers — the shutdown path in
// particular — must invoke it unconditionally rather than gate it on cache
// readiness, which is false while a spawn is still in flight.
//
// publish runs while the manager lock is held; see getOrSpawn.
func KillComputerUse(publish func(computeruse.IComputerUse)) {
	computerUse.mu.Lock()
	defer computerUse.mu.Unlock()

	if computerUse.client != nil {
		computerUse.client.Kill()
	}
	computerUse.client = nil
	computerUse.impl = nil
	if publish != nil {
		publish(nil)
	}
}
