// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package cache

import (
	"context"
	"time"

	common_cache "github.com/daytonaio/common-go/pkg/cache"
)

type SnapshotErrorCache struct {
	common_cache.ICache[string]
	retention time.Duration
}

func NewSnapshotErrorCache(ctx context.Context, retention time.Duration) *SnapshotErrorCache {
	return &SnapshotErrorCache{
		ICache:    common_cache.NewMapCache[string](ctx),
		retention: retention,
	}
}

func (c *SnapshotErrorCache) SetError(ctx context.Context, snapshot string, errReason string) error {
	return c.Set(ctx, snapshot, errReason, c.retention)
}

func (c *SnapshotErrorCache) GetError(ctx context.Context, snapshot string) (*string, error) {
	return c.Get(ctx, snapshot)
}

func (c *SnapshotErrorCache) RemoveError(ctx context.Context, snapshot string) error {
	return c.Delete(ctx, snapshot)
}
