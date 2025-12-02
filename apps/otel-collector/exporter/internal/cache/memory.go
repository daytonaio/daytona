package cache

import (
	"context"
	"time"

	"github.com/jellydator/ttlcache/v3"
)

// MemoryCache implements an in-memory cache with TTL support using ttlcache.
type MemoryCache struct {
	cache *ttlcache.Cache[string, *EndpointConfig]
}

// NewMemoryCache creates a new in-memory cache instance.
func NewMemoryCache() *MemoryCache {
	cache := ttlcache.New[string, *EndpointConfig](
		ttlcache.WithTTL[string, *EndpointConfig](5 * time.Minute),
	)

	// Start automatic expired item deletion
	go cache.Start()

	return &MemoryCache{
		cache: cache,
	}
}

// Get retrieves a cached configuration for the given sandbox ID.
func (m *MemoryCache) Get(ctx context.Context, sandboxID string) (*EndpointConfig, error) {
	item := m.cache.Get(sandboxID)
	if item == nil {
		return nil, nil
	}

	return item.Value(), nil
}

// Set stores a configuration for the given sandbox ID with TTL.
func (m *MemoryCache) Set(ctx context.Context, sandboxID string, config *EndpointConfig, ttl time.Duration) error {
	config.CachedAt = time.Now()
	m.cache.Set(sandboxID, config, ttl)
	return nil
}

// Delete removes a cached configuration for the given sandbox ID.
func (m *MemoryCache) Delete(ctx context.Context, sandboxID string) error {
	m.cache.Delete(sandboxID)
	return nil
}

// Clear removes all cached configurations.
func (m *MemoryCache) Clear(ctx context.Context) error {
	m.cache.DeleteAll()
	return nil
}

// Stop stops the cache's automatic expired item deletion.
func (m *MemoryCache) Stop() {
	m.cache.Stop()
}
