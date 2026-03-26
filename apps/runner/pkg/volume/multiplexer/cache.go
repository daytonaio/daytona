package multiplexer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
)

// VolumeCache implements a file cache for a volume
type VolumeCache struct {
	cacheDir string
	maxSize  int64
	usedSize int64
	mu       sync.RWMutex
	lruIndex *lru.Cache[string, *cacheEntry]
}

type cacheEntry struct {
	path      string
	offset    int64
	size      int
	timestamp time.Time
	etag      string
}

// NewVolumeCache creates a new cache for a volume
func NewVolumeCache(cacheDir string, maxSize int64) *VolumeCache {
	lruCache, _ := lru.New[string, *cacheEntry](10000)

	// Ensure cache directory exists
	os.MkdirAll(cacheDir, 0755)

	return &VolumeCache{
		cacheDir: cacheDir,
		maxSize:  maxSize,
		lruIndex: lruCache,
	}
}

// Get retrieves data from cache if available
func (c *VolumeCache) Get(path string, offset int64, size int) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := c.cacheKey(path, offset, size)
	entry, exists := c.lruIndex.Get(key)
	if !exists {
		return nil, false
	}

	// Check if cache is still fresh (5 seconds TTL)
	if time.Since(entry.timestamp) > 5*time.Second {
		return nil, false
	}

	// Read from cache file
	cacheFile := filepath.Join(c.cacheDir, key)
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil, false
	}

	return data, true
}

// Put stores data in cache
func (c *VolumeCache) Put(path string, offset int64, data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.cacheKey(path, offset, len(data))
	cacheFile := filepath.Join(c.cacheDir, key)

	// Check if we need to evict old entries
	newSize := c.usedSize + int64(len(data))
	if newSize > c.maxSize {
		c.evictLRU(int64(len(data)))
	}

	// Write to cache file
	if err := os.WriteFile(cacheFile, data, 0644); err != nil {
		return
	}

	// Update index
	entry := &cacheEntry{
		path:      path,
		offset:    offset,
		size:      len(data),
		timestamp: time.Now(),
	}
	c.lruIndex.Add(key, entry)
	c.usedSize += int64(len(data))
}

// Invalidate removes all cache entries for a file
func (c *VolumeCache) Invalidate(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Find all entries for this path
	keysToRemove := []string{}
	for _, key := range c.lruIndex.Keys() {
		if entry, ok := c.lruIndex.Peek(key); ok && entry.path == path {
			keysToRemove = append(keysToRemove, key)
		}
	}

	// Remove entries
	for _, key := range keysToRemove {
		if entry, ok := c.lruIndex.Get(key); ok {
			cacheFile := filepath.Join(c.cacheDir, key)
			os.Remove(cacheFile)
			c.usedSize -= int64(entry.size)
			c.lruIndex.Remove(key)
		}
	}
}

// Clear removes all cache entries
func (c *VolumeCache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Remove all cache files
	entries, err := os.ReadDir(c.cacheDir)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			os.Remove(filepath.Join(c.cacheDir, entry.Name()))
		}
	}

	c.lruIndex.Purge()
	c.usedSize = 0
	return nil
}

// evictLRU evicts least recently used entries to make space
func (c *VolumeCache) evictLRU(needed int64) {
	freedSpace := int64(0)

	for freedSpace < needed && c.lruIndex.Len() > 0 {
		key, entry, _ := c.lruIndex.RemoveOldest()
		cacheFile := filepath.Join(c.cacheDir, key)
		os.Remove(cacheFile)
		freedSpace += int64(entry.size)
		c.usedSize -= int64(entry.size)
	}
}

// cacheKey generates a unique key for a cache entry
func (c *VolumeCache) cacheKey(path string, offset int64, size int) string {
	h := sha256.New()
	io.WriteString(h, path)
	io.WriteString(h, fmt.Sprintf(":%d:%d", offset, size))
	return hex.EncodeToString(h.Sum(nil))[:16]
}
