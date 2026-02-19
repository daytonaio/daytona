// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package cache

import (
	"context"
	"time"

	"github.com/daytonaio/runner/pkg/models"
	"github.com/daytonaio/runner/pkg/models/enums"

	common_cache "github.com/daytonaio/common-go/pkg/cache"
)

type StatesCache struct {
	common_cache.ICache[models.CachedStates]
	cacheRetentionDays int
}

var statesCache *StatesCache

func GetStatesCache(ctx context.Context, cacheRetentionDays int) *StatesCache {
	if cacheRetentionDays <= 0 {
		cacheRetentionDays = 7
	}

	if statesCache != nil {
		if statesCache.cacheRetentionDays != cacheRetentionDays {
			statesCache.cacheRetentionDays = cacheRetentionDays
		}

		return statesCache
	}

	return &StatesCache{
		ICache:             common_cache.NewMapCache[models.CachedStates](ctx),
		cacheRetentionDays: cacheRetentionDays,
	}
}

func (sc *StatesCache) SetSandboxState(ctx context.Context, sandboxId string, state enums.SandboxState) {
	// Get existing state or create new one
	existing, err := sc.Get(ctx, sandboxId)
	if err != nil {
		// Key doesn't exist, create new entry
		existing = &models.CachedStates{}
	}

	// Update sandbox state
	existing.SandboxState = state

	// Save back to cache
	_ = sc.Set(ctx, sandboxId, *existing, sc.getEntryExpiration())
}

func (sc *StatesCache) SetBackupState(ctx context.Context, sandboxId string, state enums.BackupState, backupErr error) {
	// Get existing state or create new one
	existing, err := sc.Get(ctx, sandboxId)
	if err != nil {
		// Key doesn't exist, create new entry
		existing = &models.CachedStates{}
	}

	// Update backup state
	existing.BackupState = state

	// Set error reason if error is provided
	if backupErr != nil {
		errMsg := backupErr.Error()
		existing.BackupErrorReason = &errMsg
	} else {
		existing.BackupErrorReason = nil
	}

	// Save back to cache
	_ = sc.Set(ctx, sandboxId, *existing, sc.getEntryExpiration())
}

func (sc *StatesCache) getEntryExpiration() time.Duration {
	return time.Duration(sc.cacheRetentionDays) * 24 * time.Hour
}
