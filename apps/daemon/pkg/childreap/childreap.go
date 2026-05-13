// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

// Package childreap wraps github.com/ramr/go-reaper with cooperative recovery
// of child exit status.
//
// When the daemon runs as PID 1, it must call wait4(-1, ...) to reap orphaned
// grandchildren that get reparented to it. But the daemon also calls
// cmd.Wait() (i.e. wait4(specificPid, ...)) on processes it spawned itself.
// Those two calls race: whichever syscall the kernel dispatches first wins,
// and the other returns ECHILD. When the reaper wins, cmd.Wait() reports a
// non-ExitError error and the daemon loses the real exit code — handlers
// then surface exitCode=-1 to API clients even when the process succeeded.
//
// This package wires up go-reaper's StatusChannel so the reaper publishes
// (pid, waitStatus) for every child it claims. The Wait helper consults that
// status if cmd.Wait() lost the race, reconstructing the correct exit code.
package childreap

import (
	"errors"
	"fmt"
	"os/exec"
	"sync"
	"syscall"
	"time"

	reaper "github.com/ramr/go-reaper"
)

const (
	// statusChannelBuf must be large enough that the reaper never has to
	// drop a notification because the dispatcher is briefly slow. 1024 is
	// generous for the daemon's expected exec workload.
	statusChannelBuf = 1024

	// pendingTTL bounds how long an unclaimed reaper status sits in the
	// pending map. Long enough that Wait() registering moments after the
	// child exited can still find it; short enough that an exited PID
	// being reused doesn't deliver stale status to the new process.
	pendingTTL = 30 * time.Second

	// pendingSweepInterval is how often we evict expired pending entries.
	pendingSweepInterval = 5 * time.Second
)

// recoveryTimeout caps how long Wait() will block after cmd.Wait() returned
// ECHILD, waiting for the reaper to publish the status. In practice this
// should fire within microseconds — the timeout exists only to guarantee
// Wait() never hangs forever. var (not const) so tests can shorten it.
var recoveryTimeout = 5 * time.Second

type pidRegistration struct {
	// ch is buffered 1 so the dispatcher's send never blocks.
	ch chan syscall.WaitStatus
}

type pendingStatus struct {
	ws      syscall.WaitStatus
	addedAt time.Time
}

var (
	mu       sync.Mutex
	registry = make(map[int]*pidRegistration) // pid -> waiter, populated by Wait
	pending  = make(map[int]pendingStatus)    // pid -> status, populated by dispatcher when nobody's waiting yet

	startOnce sync.Once
)

// Start installs the PID-1 zombie reaper and the cooperative-status
// dispatcher. Idempotent: safe to call more than once but only the first
// call has effect. Must be called once early in main(), before any
// exec.Cmd is spawned.
func Start() {
	startOnce.Do(func() {
		ch := make(chan reaper.Status, statusChannelBuf)
		go dispatch(ch)
		go reaper.Start(reaper.Config{
			Pid:           -1,
			Options:       0,
			StatusChannel: ch,
		})
	})
}

func dispatch(ch chan reaper.Status) {
	ticker := time.NewTicker(pendingSweepInterval)
	defer ticker.Stop()

	for {
		select {
		case s, ok := <-ch:
			if !ok {
				return
			}
			recordStatus(s.Pid, s.WaitStatus)
		case <-ticker.C:
			sweepPending()
		}
	}
}

func recordStatus(pid int, ws syscall.WaitStatus) {
	mu.Lock()
	defer mu.Unlock()
	if reg, ok := registry[pid]; ok {
		// Waiter is parked; deliver directly. Non-blocking because ch is
		// buffered 1 and a pid can only be reaped once.
		select {
		case reg.ch <- ws:
		default:
		}
		return
	}
	pending[pid] = pendingStatus{ws: ws, addedAt: time.Now()}
}

func sweepPending() {
	cutoff := time.Now().Add(-pendingTTL)
	mu.Lock()
	defer mu.Unlock()
	for pid, ps := range pending {
		if ps.addedAt.Before(cutoff) {
			delete(pending, pid)
		}
	}
}

// Wait wraps cmd.Wait() with cooperative recovery from the PID-1 reaper.
//
// Returns the process exit code (matching os.ProcessState.ExitCode()
// semantics: 0..255 for normal exit, -1 for signal-terminated processes).
// Returns a non-nil error only when the exit status truly cannot be
// recovered — e.g., cmd was never started, or the reaper never published a
// status within recoveryTimeout.
//
// Safe for concurrent use across goroutines, each waiting on its own cmd.
func Wait(cmd *exec.Cmd) (int, error) {
	if cmd == nil || cmd.Process == nil {
		return -1, errors.New("childreap.Wait: cmd not started")
	}
	pid := cmd.Process.Pid

	// Register BEFORE cmd.Wait so a fast reaper still routes our status
	// into reg.ch instead of the pending map.
	reg := register(pid)
	defer unregister(pid)

	err := cmd.Wait()
	if err == nil {
		return 0, nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode(), nil
	}

	// cmd.Wait() returned a non-ExitError — almost always ECHILD because
	// the reaper consumed the zombie first. Recover from the reaper's
	// status report.
	ws, recovered := recoverStatus(pid, reg)
	if !recovered {
		return -1, fmt.Errorf("childreap.Wait: lost exit status for pid %d: %w", pid, err)
	}
	return ws.ExitStatus(), nil
}

func register(pid int) *pidRegistration {
	reg := &pidRegistration{ch: make(chan syscall.WaitStatus, 1)}
	mu.Lock()
	registry[pid] = reg
	mu.Unlock()
	return reg
}

func unregister(pid int) {
	mu.Lock()
	delete(registry, pid)
	mu.Unlock()
}

// recoverStatus is only called after cmd.Wait() returned ECHILD, so we
// know the PID has exited. Check pending first (status arrived before
// register completed); otherwise park on the waiter channel.
func recoverStatus(pid int, reg *pidRegistration) (syscall.WaitStatus, bool) {
	mu.Lock()
	if ps, ok := pending[pid]; ok {
		delete(pending, pid)
		mu.Unlock()
		return ps.ws, true
	}
	mu.Unlock()

	select {
	case ws := <-reg.ch:
		return ws, true
	case <-time.After(recoveryTimeout):
		// One last look in case status raced in during the timeout.
		mu.Lock()
		defer mu.Unlock()
		if ps, ok := pending[pid]; ok {
			delete(pending, pid)
			return ps.ws, true
		}
		return 0, false
	}
}
