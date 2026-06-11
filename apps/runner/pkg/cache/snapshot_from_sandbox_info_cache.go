// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package cache

import (
	"context"
	"time"

	"github.com/daytonaio/runner/pkg/api/dto"
	"github.com/daytonaio/runner/pkg/models"
	"github.com/daytonaio/runner/pkg/models/enums"

	common_cache "github.com/daytonaio/common-go/pkg/cache"
)

type SnapshotFromSandboxInfoCache struct {
	common_cache.ICache[models.SnapshotFromSandboxInfo]
	retention time.Duration
}

func NewSnapshotFromSandboxInfoCache(ctx context.Context, retention time.Duration) *SnapshotFromSandboxInfoCache {
	return &SnapshotFromSandboxInfoCache{
		ICache:    common_cache.NewMapCache[models.SnapshotFromSandboxInfo](ctx),
		retention: retention,
	}
}

func (c *SnapshotFromSandboxInfoCache) SetCaptureState(ctx context.Context, sandboxId, name string, state enums.SnapshotFromSandboxState, info *dto.SnapshotInfoResponse, err error) error {
	entry := models.SnapshotFromSandboxInfo{
		Name:  name,
		State: state,
		Info:  info,
		Error: err,
	}

	return c.Set(ctx, sandboxId, entry, c.retention)
}
