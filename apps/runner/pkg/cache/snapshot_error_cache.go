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
	ctx       context.Context
}

func NewSnapshotErrorCache(ctx context.Context, retention time.Duration) *SnapshotErrorCache {
	return &SnapshotErrorCache{
		ICache:    common_cache.NewMapCache[string](ctx),
		retention: retention,
		ctx:       ctx,
	}
}

func (c *SnapshotErrorCache) SetError(snapshot string, errReason string) error {
	return c.Set(c.ctx, snapshot, errReason, c.retention)
}

func (c *SnapshotErrorCache) GetError(snapshot string) (*string, error) {
	return c.Get(c.ctx, snapshot)
}

func (c *SnapshotErrorCache) RemoveError(snapshot string) error {
	return c.Delete(c.ctx, snapshot)
}
