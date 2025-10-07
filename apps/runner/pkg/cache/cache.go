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
}

var statesCache *StatesCache

func GetStatesCache() *StatesCache {
	if statesCache == nil {
		statesCache = &StatesCache{
			ICache: common_cache.NewMapCache[models.CachedStates](),
		}
	}

	return statesCache
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
	_ = sc.Set(ctx, sandboxId, *existing, 7*24*time.Hour)
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
	_ = sc.Set(ctx, sandboxId, *existing, 7*24*time.Hour)
}
