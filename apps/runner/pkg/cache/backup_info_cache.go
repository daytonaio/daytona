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
	retention time.Duration
}

func NewBackupInfoCache(ctx context.Context, retention time.Duration) *BackupInfoCache {
	return &BackupInfoCache{
		ICache:    common_cache.NewMapCache[models.BackupInfo](ctx),
		retention: retention,
	}
}

func (c *BackupInfoCache) SetBackupState(ctx context.Context, sandboxId string, state enums.BackupState, err error) error {
	entry := models.BackupInfo{
		State: state,
		Error: err,
	}

	return c.Set(ctx, sandboxId, entry, c.retention)
}
