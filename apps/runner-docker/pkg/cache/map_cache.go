// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package cache

import (
	"context"
	"errors"
	"time"

	cmap "github.com/orcaman/concurrent-map/v2"
)

type MapCache[T any] struct {
	cacheMap cmap.ConcurrentMap[string, T]
}

func (c *MapCache[T]) Set(ctx context.Context, key string, value T, expiration time.Duration) error {
	c.cacheMap.Set(key, value)
	return nil
}

func (c *MapCache[T]) Has(ctx context.Context, key string) (bool, error) {
	_, ok := c.cacheMap.Get(key)
	return ok, nil
}

func (c *MapCache[T]) Get(ctx context.Context, key string) (*T, error) {
	value, ok := c.cacheMap.Get(key)
	if !ok {
		return nil, errors.New("key not found")
	}
	return &value, nil
}

func (c *MapCache[T]) Delete(ctx context.Context, key string) error {
	c.cacheMap.Remove(key)
	return nil
}

func NewMapCache[T any]() *MapCache[T] {
	return &MapCache[T]{
		cacheMap: cmap.New[T](),
	}
}
