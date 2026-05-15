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
	"bytes"
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
	// pending map before the sweeper evicts it for memory cleanliness.
	// Claim-time staleness (pendingMaxAge) is what prevents stale entries
	// from being matched against recycled pids — TTL only controls when
	// orphan entries that never get claimed are released.
	pendingTTL = 10 * time.Second

	// pendingSweepInterval is how often we evict expired pending entries.
	pendingSweepInterval = 5 * time.Second
)

// recoveryTimeout caps how long Wait() will block after cmd.Wait() returned
// ECHILD, waiting for the reaper to publish the status. In practice this
// should fire within microseconds — the timeout exists only to guarantee
// Wait() never hangs forever. var (not const) so tests can shorten it.
var recoveryTimeout = 5 * time.Second

// pendingMaxAge is the maximum age, relative to the registering caller's
// own time, of a pending entry we'll accept. Entries older than this are
// treated as stale and discarded rather than matched against the caller's
// pid — they're almost certainly from a previous process whose pid has
// since been recycled by the kernel (Linux's default pid_max is 32k and
// cycles quickly under fork-heavy workloads).
//
// The legitimate use of pending is the pre-register race: a child exits
// and the reaper records its status BEFORE Wait/Reap calls register.
// That window is bounded by the (typically sub-millisecond) gap between
// cmd.Start() returning and the caller invoking Wait/Reap; 1 second is
// generous enough to cover heavily-scheduled goroutines while staying
// well below realistic pid-recycle horizons.
var pendingMaxAge = 1 * time.Second

type pidRegistration struct {
	// ch is buffered 1 so the dispatcher's send never blocks.
	ch chan syscall.WaitStatus
	// registeredAt is the wall-clock time the registration was created.
	// claimPending compares pending entries' addedAt against this to
	// reject stale orphans from recycled pids (see pendingMaxAge).
	registeredAt time.Time
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

// Wait reaps cmd, ensures cmd.Wait completes (so any I/O copy goroutines
// drain into bytes.Buffer/strings.Reader-style targets), and returns the
// exit code.
//
// Use Wait when the caller reads cmd.Stdout/Stderr buffers AFTER this
// returns — typical for /process/execute and /process/code-run, which
// set cmd.Stdout/Stderr to *bytes.Buffer. For non-*os.File targets Go
// starts internal goroutines to copy from the child's pipes; those
// goroutines only finish when cmd.Wait completes. Reading the buffer
// before cmd.Wait drains leaves it empty or truncated.
//
// Long-running commands aren't penalized: the wait-for-status phase has
// no timeout. Only the post-status I/O-drain phase has a hangTimeout
// backstop, to prevent a wedged cmd.Wait from blocking forever.
//
// Returns the exit code matching os.ProcessState.ExitCode() semantics:
// 0..255 for normal exit, -1 for signal-terminated processes. Returns a
// non-nil error only when no exit status could be recovered.
func Wait(cmd *exec.Cmd) (int, error) {
	if cmd == nil || cmd.Process == nil {
		return -1, errors.New("childreap.Wait: cmd not started")
	}
	pid := cmd.Process.Pid

	reg := register(pid)
	defer unregister(pid)

	waitCh := startCmdWaitGoroutine(cmd)

	// Status may have arrived before we registered.
	var reaperStatus *syscall.WaitStatus
	if ws, ok := claimPending(pid, reg.registeredAt); ok {
		reaperStatus = &ws
	}

	// Phase 1: wait for exit status. No timeout — long-running commands
	// are legitimate. Returns when EITHER cmd.Wait or the reaper resolves.
	if reaperStatus == nil {
		select {
		case r := <-waitCh:
			// cmd.Wait completed; I/O goroutines drained.
			if r.resolved {
				return r.exitCode, nil
			}
			// cmd.Wait returned ECHILD. Cmd.Wait drains I/O goroutines
			// regardless of Process.Wait's outcome, so I/O is already
			// settled — only the status needs recovery.
			if ws, ok := claimPending(pid, reg.registeredAt); ok {
				return ws.ExitStatus(), nil
			}
			select {
			case ws := <-reg.ch:
				return ws.ExitStatus(), nil
			case <-time.After(recoveryTimeout):
				return -1, fmt.Errorf("childreap.Wait: lost exit status for pid %d: %w", pid, r.err)
			}
		case ws := <-reg.ch:
			reaperStatus = &ws
		}
	}

	// Phase 2: we have status from the reaper, but cmd.Wait hasn't
	// completed yet — I/O copy goroutines may still be running. Wait for
	// cmd.Wait, bounded by hangTimeout so a wedged Go runtime doesn't
	// hold the caller forever.
	select {
	case <-waitCh:
	case <-time.After(hangTimeout):
	}
	return reaperStatus.ExitStatus(), nil
}

// Reap returns the exit code as soon as it's known. Does not wait for
// cmd.Wait to fully complete, so internal I/O copy goroutines may still
// be running when this returns.
//
// Use Reap on cleanup paths where the caller doesn't read
// cmd.Stdout/Stderr after the call — e.g., session/PTY/interpreter
// teardown. Faster than Wait when the PID-1 reaper consumed the zombie
// before cmd.Wait could, and crucially avoids hanging if cmd.Wait itself
// is wedged for unrelated reasons.
func Reap(cmd *exec.Cmd) (int, error) {
	if cmd == nil || cmd.Process == nil {
		return -1, errors.New("childreap.Reap: cmd not started")
	}
	pid := cmd.Process.Pid

	reg := register(pid)
	defer unregister(pid)

	if ws, ok := claimPending(pid, reg.registeredAt); ok {
		// Fire-and-forget cmd.Wait so file descriptors eventually close.
		// The goroutine returns on its own (kernel ECHILD).
		go func() { _ = cmd.Wait() }()
		return ws.ExitStatus(), nil
	}

	waitCh := startCmdWaitGoroutine(cmd)

	select {
	case r := <-waitCh:
		if r.resolved {
			return r.exitCode, nil
		}
		if ws, ok := claimPending(pid, reg.registeredAt); ok {
			return ws.ExitStatus(), nil
		}
		select {
		case ws := <-reg.ch:
			return ws.ExitStatus(), nil
		case <-time.After(recoveryTimeout):
			return -1, fmt.Errorf("childreap.Reap: lost exit status for pid %d: %w", pid, r.err)
		}
	case ws := <-reg.ch:
		return ws.ExitStatus(), nil
	case <-time.After(hangTimeout):
		if ws, ok := claimPending(pid, reg.registeredAt); ok {
			return ws.ExitStatus(), nil
		}
		return -1, fmt.Errorf("childreap.Reap: timed out waiting for pid %d", pid)
	}
}

type cmdWaitResult struct {
	exitCode int
	err      error
	// resolved is true for nil-error or *exec.ExitError outcomes; false
	// means cmd.Wait failed in a way (likely ECHILD) that we should fall
	// through to channel/pending recovery.
	resolved bool
}

func startCmdWaitGoroutine(cmd *exec.Cmd) chan cmdWaitResult {
	waitCh := make(chan cmdWaitResult, 1)
	go func() {
		err := cmd.Wait()
		if err == nil {
			waitCh <- cmdWaitResult{0, nil, true}
			return
		}
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			waitCh <- cmdWaitResult{ee.ExitCode(), nil, true}
			return
		}
		waitCh <- cmdWaitResult{-1, err, false}
	}()
	return waitCh
}

// hangTimeout caps how long Wait/Reap will block AFTER the exit status
// is known (Phase 2) waiting for cmd.Wait to drain. Phase 1 (waiting for
// status itself) has no timeout — long-running children are legitimate.
var hangTimeout = 30 * time.Second

// Run starts cmd and waits for it to finish. The reaper-safe analog of
// cmd.Run().
//
// Unlike cmd.Run(), this returns (exitCode, err) instead of folding
// non-zero exits into err — that's intentional: cmd.Run()'s behavior of
// returning *exec.ExitError only when cmd.Wait() recovers a real
// ProcessState breaks under PID-1 reaping (cmd.Wait() loses the race
// and returns ECHILD-wrapped SyscallError, which callers using a
// type-switch to detect non-zero exits misclassify as "command
// failed"). Callers who want the same semantic should check
// `exitCode != 0` explicitly.
//
// Returns a non-nil error only when cmd couldn't be started or its exit
// status couldn't be recovered (see Wait for details).
func Run(cmd *exec.Cmd) (int, error) {
	if cmd == nil {
		return -1, errors.New("childreap.Run: nil cmd")
	}
	if err := cmd.Start(); err != nil {
		return -1, err
	}
	return Wait(cmd)
}

// CombinedOutput runs cmd with stdout and stderr both captured into a
// single buffer. The reaper-safe analog of cmd.CombinedOutput().
//
// Returns (output, exitCode, err); see Run for the (exitCode, err)
// contract. err being nil does NOT mean the command succeeded — check
// exitCode for that.
func CombinedOutput(cmd *exec.Cmd) ([]byte, int, error) {
	if cmd == nil {
		return nil, -1, errors.New("childreap.CombinedOutput: nil cmd")
	}
	if cmd.Stdout != nil {
		return nil, -1, errors.New("childreap.CombinedOutput: cmd.Stdout already set")
	}
	if cmd.Stderr != nil {
		return nil, -1, errors.New("childreap.CombinedOutput: cmd.Stderr already set")
	}
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	code, err := Run(cmd)
	return buf.Bytes(), code, err
}

// Output runs cmd and captures stdout into a buffer. The reaper-safe
// analog of cmd.Output(). Stderr is left as-is (typically discarded by
// Go's exec when unset).
//
// Returns (stdout, exitCode, err); see Run for the (exitCode, err)
// contract.
func Output(cmd *exec.Cmd) ([]byte, int, error) {
	if cmd == nil {
		return nil, -1, errors.New("childreap.Output: nil cmd")
	}
	if cmd.Stdout != nil {
		return nil, -1, errors.New("childreap.Output: cmd.Stdout already set")
	}
	var buf bytes.Buffer
	cmd.Stdout = &buf
	code, err := Run(cmd)
	return buf.Bytes(), code, err
}

func register(pid int) *pidRegistration {
	reg := &pidRegistration{
		ch:           make(chan syscall.WaitStatus, 1),
		registeredAt: time.Now(),
	}
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

// claimPending returns and removes a pending status for pid if one exists
// AND it's recent enough (within pendingMaxAge of registeredAt) to
// plausibly belong to the registering caller's process. Stale entries —
// older than pendingMaxAge relative to registeredAt — are discarded as
// orphan statuses whose pid has been recycled by the kernel since.
func claimPending(pid int, registeredAt time.Time) (syscall.WaitStatus, bool) {
	mu.Lock()
	defer mu.Unlock()
	ps, ok := pending[pid]
	if !ok {
		return 0, false
	}
	if registeredAt.Sub(ps.addedAt) > pendingMaxAge {
		// Stale: this status predates our registration by more than the
		// fast-exit recovery window, so it can't be for our process.
		// Drop it so subsequent calls don't keep seeing it.
		delete(pending, pid)
		return 0, false
	}
	delete(pending, pid)
	return ps.ws, true
}
