package sdisk

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// LayerCache manages the global layer cache
type LayerCache struct {
	cacheDir string
	s3Client *S3Client
	db       *DB
	mu       sync.RWMutex
}

// NewLayerCache creates a new layer cache
func NewLayerCache(cacheDir string, s3Client *S3Client, db *DB) (*LayerCache, error) {
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create layer cache directory: %w", err)
	}

	return &LayerCache{
		cacheDir: cacheDir,
		s3Client: s3Client,
		db:       db,
	}, nil
}

// GetOrDownload retrieves layer from cache or downloads if missing
func (c *LayerCache) GetOrDownload(ctx context.Context, diskName, layerID string, metadata S3LayerInfo) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if layer already exists in cache
	layerPath := c.getLayerPath(layerID)
	if _, err := os.Stat(layerPath); err == nil {
		// Layer exists in cache, increment ref count
		if err := c.db.IncrementLayerRefCount(layerID); err != nil {
			return "", fmt.Errorf("failed to increment ref count for layer %s: %w", layerID, err)
		}
		return layerPath, nil
	}

	// Layer not in cache, download it
	if err := c.s3Client.DownloadLayer(ctx, diskName, layerID, layerPath); err != nil {
		return "", fmt.Errorf("failed to download layer %s: %w", layerID, err)
	}

	// Save layer state to database
	layerState := &LayerState{
		ID:       layerID,
		Checksum: metadata.Checksum,
		Size:     metadata.Size,
		CachedAt: time.Now(),
		RefCount: 1, // First reference
	}

	if err := c.db.SaveLayer(layerState); err != nil {
		// Cleanup downloaded file on database error
		os.Remove(layerPath)
		return "", fmt.Errorf("failed to save layer state: %w", err)
	}

	return layerPath, nil
}

// GetLayerPath returns the path to a cached layer (empty if not cached)
func (c *LayerCache) GetLayerPath(layerID string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.getLayerPath(layerID)
}

// getLayerPath returns the path to a cached layer (internal, no locking)
func (c *LayerCache) getLayerPath(layerID string) string {
	return filepath.Join(c.cacheDir, layerID+".qcow2")
}

// CleanupUnusedLayers removes layers with zero ref count
func (c *LayerCache) CleanupUnusedLayers(ctx context.Context) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Get unused layers from database
	unusedLayers, err := c.db.ListUnusedLayers()
	if err != nil {
		return 0, fmt.Errorf("failed to list unused layers: %w", err)
	}

	cleanedCount := 0
	for _, layer := range unusedLayers {
		layerPath := c.getLayerPath(layer.ID)

		// Remove file from filesystem
		if err := os.Remove(layerPath); err != nil && !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "warning: failed to remove layer file %s: %v\n", layerPath, err)
			continue
		}

		// Remove from database
		query := `DELETE FROM layers WHERE id = ?`
		if _, err := c.db.db.Exec(query, layer.ID); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to delete layer %s from database: %v\n", layer.ID, err)
			continue
		}

		cleanedCount++
	}

	return cleanedCount, nil
}
