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

// Wait waits for cmd to exit and returns its exit code.
//
// Unlike cmd.Wait(), Wait does not block on the kernel-level wait4 syscall
// for the child PID. Instead it races cmd.Wait() (in a goroutine) against
// the PID-1 reaper's status channel and returns as soon as EITHER resolves.
// This matters because cmd.Wait() can block indefinitely when the reaper
// has already consumed the zombie — the Go runtime's wait machinery does
// not always handle externally-reaped children promptly. The reaper's
// status channel is our own mechanism and is guaranteed to deliver.
//
// Returns the exit code matching os.ProcessState.ExitCode() semantics:
// 0..255 for normal exit, -1 for signal-terminated processes. Returns a
// non-nil error only when no exit status could be recovered within
// hangTimeout (e.g., cmd was never started, or the reaper missed it).
//
// Safe for concurrent use across goroutines, each waiting on its own cmd.
func Wait(cmd *exec.Cmd) (int, error) {
	if cmd == nil || cmd.Process == nil {
		return -1, errors.New("childreap.Wait: cmd not started")
	}
	pid := cmd.Process.Pid

	// Register BEFORE doing anything else so a fast reaper routes our
	// status into reg.ch instead of pending.
	reg := register(pid)
	defer unregister(pid)

	// The reaper may have already published status before we registered.
	if ws, ok := claimPending(pid); ok {
		// Fire-and-forget cmd.Wait for pipe cleanup. We don't wait for it
		// because cmd.Wait can hang in this case; the runtime/GC will
		// eventually clean up file descriptors if the goroutine never
		// returns.
		go func() { _ = cmd.Wait() }()
		return ws.ExitStatus(), nil
	}

	// Race cmd.Wait against the reaper status channel. Each outcome path
	// is independent — whichever arrives first wins.
	type cmdResult struct {
		exitCode int
		err      error
		// resolved is true for nil-error or *exec.ExitError outcomes;
		// false means cmd.Wait failed in a way (likely ECHILD) that we
		// should fall through to channel/pending recovery.
		resolved bool
	}
	waitCh := make(chan cmdResult, 1)
	go func() {
		err := cmd.Wait()
		if err == nil {
			waitCh <- cmdResult{0, nil, true}
			return
		}
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			waitCh <- cmdResult{ee.ExitCode(), nil, true}
			return
		}
		waitCh <- cmdResult{-1, err, false}
	}()

	select {
	case r := <-waitCh:
		if r.resolved {
			return r.exitCode, nil
		}
		// cmd.Wait returned ECHILD-ish. Wait briefly for the reaper.
		if ws, ok := claimPending(pid); ok {
			return ws.ExitStatus(), nil
		}
		select {
		case ws := <-reg.ch:
			return ws.ExitStatus(), nil
		case <-time.After(recoveryTimeout):
			return -1, fmt.Errorf("childreap.Wait: lost exit status for pid %d: %w", pid, r.err)
		}

	case ws := <-reg.ch:
		// Reaper delivered status before cmd.Wait. We have the answer;
		// don't block the caller on cmd.Wait completing.
		return ws.ExitStatus(), nil

	case <-time.After(hangTimeout):
		// Both paths failed to resolve. Either the child never exited or
		// something is wedged. Return what we can.
		if ws, ok := claimPending(pid); ok {
			return ws.ExitStatus(), nil
		}
		return -1, fmt.Errorf("childreap.Wait: timed out waiting for pid %d", pid)
	}
}

// hangTimeout is the upper bound on how long Wait will block. Above this,
// we assume something is wedged (kernel-level wait stuck, lost SIGCHLD,
// daemon not running as PID 1, etc.) and return an error rather than hang
// the caller indefinitely. Sized to comfortably cover normal child exits
// while still surfacing real hangs in API request handling.
var hangTimeout = 30 * time.Second

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

func claimPending(pid int) (syscall.WaitStatus, bool) {
	mu.Lock()
	defer mu.Unlock()
	if ps, ok := pending[pid]; ok {
		delete(pending, pid)
		return ps.ws, true
	}
	return 0, false
}
