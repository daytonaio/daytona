// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package process

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// ErrProcessNotFound is returned when a process is not tracked.
var ErrProcessNotFound = errors.New("process not found")

// ProcessType identifies which subsystem owns a tracked process.
type ProcessType string

const (
	ProcessTypeSession     ProcessType = "session"
	ProcessTypePTY         ProcessType = "pty"
	ProcessTypeInterpreter ProcessType = "interpreter"
	ProcessTypeExec        ProcessType = "exec"
	ProcessTypeCodeRun     ProcessType = "code_run"
)

// ProcessEntry holds metadata about a tracked process.
type ProcessEntry struct {
	PID       int               `json:"pid"`
	Token     uint64            `json:"token"`
	Type      ProcessType       `json:"type"`
	ID        string            `json:"id"`
	Command   string            `json:"command,omitempty"`
	Cwd       string            `json:"cwd,omitempty"`
	Envs      map[string]string `json:"envs,omitempty"`
	Internal  bool              `json:"internal,omitempty"`
	CreatedAt time.Time         `json:"createdAt"`
	cancel    context.CancelFunc
}

// ListProcessesResponse is the HTTP response for GET /process/list.
type ListProcessesResponse struct {
	Processes []ProcessEntry `json:"processes"`
} //	@name	ListProcessesResponse

// ProcessTracker is a central, thread-safe registry of running processes.
type ProcessTracker struct {
	mu        sync.RWMutex
	processes map[int]*ProcessEntry
	tokens    atomic.Uint64
}

// NewProcessTracker creates a new empty tracker.
func NewProcessTracker() *ProcessTracker {
	return &ProcessTracker{
		processes: make(map[int]*ProcessEntry),
	}
}

// WithProcessCancel attaches a subsystem-specific cancel hook to a process entry.
func WithProcessCancel(entry ProcessEntry, cancel context.CancelFunc) ProcessEntry {
	entry.cancel = cancel
	return entry
}

// Register adds a process to the tracker.
func (t *ProcessTracker) Register(entry ProcessEntry) uint64 {
	if t == nil {
		return 0
	}

	token := t.tokens.Add(1)
	entry.Token = token

	t.mu.Lock()
	defer t.mu.Unlock()
	t.processes[entry.PID] = &entry

	return token
}

// Deregister removes a process from the tracker.
func (t *ProcessTracker) Deregister(pid int, token uint64) {
	if t == nil {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	entry, ok := t.processes[pid]
	if !ok || entry.Token != token {
		return
	}

	delete(t.processes, pid)
}

// List returns a snapshot of all tracked processes.
func (t *ProcessTracker) List() []ProcessEntry {
	if t == nil {
		return nil
	}

	t.mu.RLock()
	defer t.mu.RUnlock()

	result := make([]ProcessEntry, 0, len(t.processes))
	for _, entry := range t.processes {
		result = append(result, *entry)
	}

	return result
}

// FindByPID looks up a tracked process by its OS PID.
func (t *ProcessTracker) FindByPID(pid int) (*ProcessEntry, bool) {
	if t == nil {
		return nil, false
	}

	t.mu.RLock()
	defer t.mu.RUnlock()

	entry, ok := t.processes[pid]
	if !ok {
		return nil, false
	}

	cp := *entry
	return &cp, true
}

// Kill terminates a process by PID.
func (t *ProcessTracker) Kill(pid int) error {
	if t == nil {
		return fmt.Errorf("process tracker not configured")
	}

	t.mu.RLock()
	entry, ok := t.processes[pid]
	var snapshot ProcessEntry
	if ok {
		snapshot = *entry
	}
	t.mu.RUnlock()
	if !ok {
		return fmt.Errorf("process with PID %d not found: %w", pid, ErrProcessNotFound)
	}

	if !t.hasToken(pid, snapshot.Token) {
		return fmt.Errorf("process with PID %d not found: %w", pid, ErrProcessNotFound)
	}

	if snapshot.cancel != nil {
		snapshot.cancel()
		return nil
	}

	return killProcessGroup(pid)
}

func (t *ProcessTracker) hasToken(pid int, token uint64) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	entry, ok := t.processes[pid]
	return ok && entry.Token == token
}

func killProcessGroup(pid int) error {
	err := syscall.Kill(-pid, syscall.SIGKILL)
	if err != nil {
		return syscall.Kill(pid, syscall.SIGKILL)
	}

	return nil
}
