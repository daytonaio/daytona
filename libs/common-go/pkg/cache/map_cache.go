// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"context"
	"errors"
	"time"

	"github.com/jellydator/ttlcache/v3"
)

type MapCache[T any] struct {
	ttlCache *ttlcache.Cache[string, T]
}

func (c *MapCache[T]) Set(ctx context.Context, key string, value T, expiration time.Duration) error {
	c.ttlCache.Set(key, value, expiration)
	return nil
}

func (c *MapCache[T]) Has(ctx context.Context, key string) (bool, error) {
	return c.ttlCache.Has(key), nil
}

func (c *MapCache[T]) Get(ctx context.Context, key string) (*T, error) {
	item := c.ttlCache.Get(key)
	if item == nil {
		return nil, errors.New("key not found")
	}

	value := item.Value()
	return &value, nil
}

func (c *MapCache[T]) Delete(ctx context.Context, key string) error {
	c.ttlCache.Delete(key)
	return nil
}

func (c *MapCache[T]) start(ctx context.Context) {
	go c.ttlCache.Start()
	<-ctx.Done()
	c.ttlCache.Stop()
}

func NewMapCache[T any](ctx context.Context) *MapCache[T] {
	cache := &MapCache[T]{
		ttlCache: ttlcache.New[string, T](),
	}

	go cache.start(ctx)

	return cache
}
