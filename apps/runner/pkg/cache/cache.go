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

type BackupInfoCache struct {
	common_cache.ICache[models.BackupInfo]
	cacheRetentionDays int
}

var backupInfoCache *BackupInfoCache

func GetBackupInfoCache(cacheRetentionDays int) *BackupInfoCache {
	if cacheRetentionDays <= 0 {
		cacheRetentionDays = 7
	}

	if backupInfoCache != nil {
		if backupInfoCache.cacheRetentionDays != cacheRetentionDays {
			backupInfoCache.cacheRetentionDays = cacheRetentionDays
		}

		return backupInfoCache
	}

	return &BackupInfoCache{
		ICache:             common_cache.NewMapCache[models.BackupInfo](),
		cacheRetentionDays: cacheRetentionDays,
	}

}

func (sc *BackupInfoCache) SetBackupState(ctx context.Context, sandboxId string, state enums.BackupState, backupErr error) {
	// Get existing state or create new one
	existing, err := sc.Get(ctx, sandboxId)
	if err != nil {
		// Key doesn't exist, create new entry
		existing = &models.BackupInfo{}
	}

	// Update backup state
	existing.State = state

	// Set error reason if error is provided
	if backupErr != nil {
		errMsg := backupErr.Error()
		existing.ErrReason = &errMsg
	} else {
		existing.ErrReason = nil
	}

	// Save back to cache
	_ = sc.Set(ctx, sandboxId, *existing, sc.getEntryExpiration())
}

func (sc *BackupInfoCache) getEntryExpiration() time.Duration {
	return time.Duration(sc.cacheRetentionDays) * 24 * time.Hour
}
