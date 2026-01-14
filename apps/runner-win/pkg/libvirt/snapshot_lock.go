// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package libvirt

import (
	"context"
	"fmt"
	"os"
	"sync"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

// snapshotLockManager manages per-snapshot locks to prevent concurrent pull operations
// from corrupting each other. It uses both in-memory mutexes (for same-process coordination)
// and file-based locks (for cross-process coordination).
type snapshotLockManager struct {
	mu    sync.Mutex
	locks map[string]*snapshotLock
}

type snapshotLock struct {
	mu       sync.Mutex
	refCount int
}

var globalSnapshotLockManager = &snapshotLockManager{
	locks: make(map[string]*snapshotLock),
}

// acquireSnapshotLock acquires both an in-memory and file-based lock for a snapshot.
// This ensures that only one pull operation can proceed for a given snapshot at a time,
// both within the same process and across multiple processes.
//
// Returns a release function that MUST be called when done (use defer).
func (l *LibVirt) acquireSnapshotLock(ctx context.Context, snapshotPath string) (func(), error) {
	lockPath := snapshotPath + ".lock"

	// First, acquire in-memory lock for same-process coordination
	memLock := globalSnapshotLockManager.getLock(snapshotPath)
	memLock.mu.Lock()

	// Then, acquire file-based lock for cross-process coordination
	lockFile, err := acquireFileLock(ctx, lockPath)
	if err != nil {
		memLock.mu.Unlock()
		globalSnapshotLockManager.releaseLock(snapshotPath)
		return nil, fmt.Errorf("failed to acquire file lock for snapshot: %w", err)
	}

	log.Debugf("Acquired snapshot lock for %s", snapshotPath)

	// Return a release function
	return func() {
		releaseFileLock(lockFile, lockPath)
		memLock.mu.Unlock()
		globalSnapshotLockManager.releaseLock(snapshotPath)
		log.Debugf("Released snapshot lock for %s", snapshotPath)
	}, nil
}

// getLock returns or creates a lock for the given snapshot path
func (m *snapshotLockManager) getLock(path string) *snapshotLock {
	m.mu.Lock()
	defer m.mu.Unlock()

	lock, exists := m.locks[path]
	if !exists {
		lock = &snapshotLock{}
		m.locks[path] = lock
	}
	lock.refCount++
	return lock
}

// releaseLock decrements the reference count and removes the lock if no longer needed
func (m *snapshotLockManager) releaseLock(path string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	lock, exists := m.locks[path]
	if !exists {
		return
	}

	lock.refCount--
	if lock.refCount <= 0 {
		delete(m.locks, path)
	}
}

// acquireFileLock creates and locks a file for cross-process synchronization
func acquireFileLock(ctx context.Context, lockPath string) (*os.File, error) {
	// Create the lock file (and parent directories if needed)
	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create lock file: %w", err)
	}

	// Try to acquire an exclusive lock with timeout
	timeout := 30 * time.Minute // Snapshot pulls can take a while
	deadline := time.Now().Add(timeout)

	for {
		// Check context cancellation
		select {
		case <-ctx.Done():
			lockFile.Close()
			return nil, ctx.Err()
		default:
		}

		// Try to acquire exclusive lock (non-blocking)
		err := syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			// Lock acquired successfully
			return lockFile, nil
		}

		if err != syscall.EWOULDBLOCK {
			lockFile.Close()
			return nil, fmt.Errorf("failed to acquire file lock: %w", err)
		}

		// Lock is held by another process, wait and retry
		if time.Now().After(deadline) {
			lockFile.Close()
			return nil, fmt.Errorf("timeout waiting for snapshot lock (another pull operation may be in progress)")
		}

		log.Debugf("Waiting for snapshot lock on %s (another operation in progress)", lockPath)
		time.Sleep(1 * time.Second)
	}
}

// releaseFileLock releases the file lock and removes the lock file
func releaseFileLock(lockFile *os.File, lockPath string) {
	if lockFile == nil {
		return
	}

	// Release the lock
	syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN)
	lockFile.Close()

	// Remove the lock file (best effort - don't fail if it doesn't work)
	os.Remove(lockPath)
}
