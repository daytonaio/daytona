// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package cache

import (
	"context"
	"sync"
	"time"

	"github.com/daytonaio/mock-runner/pkg/models"
	"github.com/daytonaio/mock-runner/pkg/models/enums"
)

// StatesCache stores sandbox and backup states in memory
type StatesCache struct {
	data               map[string]*cacheEntry
	mu                 sync.RWMutex
	cacheRetentionDays int
}

type cacheEntry struct {
	states    models.CachedStates
	expiresAt time.Time
}

var statesCache *StatesCache

// GetStatesCache returns the singleton states cache
func GetStatesCache(cacheRetentionDays int) *StatesCache {
	if cacheRetentionDays <= 0 {
		cacheRetentionDays = 7
	}

	if statesCache != nil {
		if statesCache.cacheRetentionDays != cacheRetentionDays {
			statesCache.cacheRetentionDays = cacheRetentionDays
		}
		return statesCache
	}

	statesCache = &StatesCache{
		data:               make(map[string]*cacheEntry),
		cacheRetentionDays: cacheRetentionDays,
	}

	return statesCache
}

// Get retrieves cached states for a sandbox
func (sc *StatesCache) Get(ctx context.Context, sandboxId string) (*models.CachedStates, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	entry, ok := sc.data[sandboxId]
	if !ok {
		return nil, nil
	}

	if time.Now().After(entry.expiresAt) {
		return nil, nil
	}

	return &entry.states, nil
}

// Set stores cached states for a sandbox
func (sc *StatesCache) Set(ctx context.Context, sandboxId string, states models.CachedStates, expiration time.Duration) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.data[sandboxId] = &cacheEntry{
		states:    states,
		expiresAt: time.Now().Add(expiration),
	}

	return nil
}

// SetSandboxState updates the sandbox state in cache
func (sc *StatesCache) SetSandboxState(ctx context.Context, sandboxId string, state enums.SandboxState) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	entry, ok := sc.data[sandboxId]
	if !ok {
		entry = &cacheEntry{
			states: models.CachedStates{},
		}
	}

	entry.states.SandboxState = state
	entry.expiresAt = sc.getEntryExpiration()
	sc.data[sandboxId] = entry
}

// SetBackupState updates the backup state in cache
func (sc *StatesCache) SetBackupState(ctx context.Context, sandboxId string, state enums.BackupState, backupErr error) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	entry, ok := sc.data[sandboxId]
	if !ok {
		entry = &cacheEntry{
			states: models.CachedStates{},
		}
	}

	entry.states.BackupState = state

	if backupErr != nil {
		errMsg := backupErr.Error()
		entry.states.BackupErrorReason = &errMsg
	} else {
		entry.states.BackupErrorReason = nil
	}

	entry.expiresAt = sc.getEntryExpiration()
	sc.data[sandboxId] = entry
}

// Delete removes cached states for a sandbox
func (sc *StatesCache) Delete(ctx context.Context, sandboxId string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	delete(sc.data, sandboxId)
}

func (sc *StatesCache) getEntryExpiration() time.Time {
	return time.Now().Add(time.Duration(sc.cacheRetentionDays) * 24 * time.Hour)
}
