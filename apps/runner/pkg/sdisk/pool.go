package sdisk

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"
)

// DiskPool manages a pool of mounted disks with LRU eviction
type DiskPool struct {
	maxMounted int                   // Maximum number of concurrently mounted disks
	mounted    map[string]*poolEntry // Currently mounted disks
	mu         sync.RWMutex          // Protects mounted map
	shutdownCh chan struct{}         // Channel to signal shutdown
	wg         sync.WaitGroup        // Wait group for background goroutines
}

// poolEntry tracks a mounted disk and its access time
type poolEntry struct {
	disk         *disk     // The mounted disk
	lastAccessed time.Time // Last time the disk was accessed
	accessCount  int64     // Number of times the disk has been accessed
}

// NewDiskPool creates a new disk pool
func NewDiskPool(maxMounted int) *DiskPool {
	if maxMounted <= 0 {
		maxMounted = 100 // Default to 100 mounted disks
	}

	pool := &DiskPool{
		maxMounted: maxMounted,
		mounted:    make(map[string]*poolEntry),
		shutdownCh: make(chan struct{}),
	}

	return pool
}

// Get retrieves a disk from the pool, mounting it if necessary
// This method handles automatic eviction if the pool is full
func (p *DiskPool) Get(ctx context.Context, vol *disk) (*disk, error) {
	fmt.Fprintf(os.Stderr, "debug: pool.Get called for disk '%s'\n", vol.name)

	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if already mounted and in pool
	if entry, exists := p.mounted[vol.name]; exists {
		fmt.Fprintf(os.Stderr, "debug: disk '%s' already in pool, updating access time\n", vol.name)
		entry.lastAccessed = time.Now()
		entry.accessCount++
		return entry.disk, nil
	}

	// Check if disk is already mounted (but not tracked in pool)
	if vol.IsMounted() {
		fmt.Fprintf(os.Stderr, "debug: disk '%s' is already mounted but not in pool, adding to pool\n", vol.name)
		entry := &poolEntry{
			disk:         vol,
			lastAccessed: time.Now(),
			accessCount:  1,
		}
		p.mounted[vol.name] = entry
		return vol, nil
	}

	fmt.Fprintf(os.Stderr, "debug: disk '%s' needs to be mounted, checking pool capacity\n", vol.name)

	// Need to mount the disk
	// First, check if we have space in the pool
	if len(p.mounted) >= p.maxMounted {
		fmt.Fprintf(os.Stderr, "debug: pool is full (%d/%d), evicting LRU disk\n", len(p.mounted), p.maxMounted)
		// Evict LRU disk
		if err := p.evictLRU(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "debug: failed to evict LRU disk: %v\n", err)
			return nil, fmt.Errorf("failed to evict disk: %w", err)
		}
		fmt.Fprintf(os.Stderr, "debug: successfully evicted LRU disk\n")
	}

	// Mount the disk using internal method to avoid recursion
	fmt.Fprintf(os.Stderr, "debug: calling vol.mountInternal for disk '%s' from pool\n", vol.name)
	if _, err := vol.mountInternal(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "debug: vol.mountInternal failed for disk '%s': %v\n", vol.name, err)
		return nil, fmt.Errorf("failed to mount disk: %w", err)
	}
	fmt.Fprintf(os.Stderr, "debug: vol.mountInternal succeeded for disk '%s'\n", vol.name)

	// Add to pool
	fmt.Fprintf(os.Stderr, "debug: adding disk '%s' to pool\n", vol.name)
	entry := &poolEntry{
		disk:         vol,
		lastAccessed: time.Now(),
		accessCount:  1,
	}
	p.mounted[vol.name] = entry

	fmt.Fprintf(os.Stderr, "debug: pool.Get completed successfully for disk '%s'\n", vol.name)
	return vol, nil
}

// Release marks a disk as no longer being actively used
// The disk remains mounted but becomes eligible for eviction
func (p *DiskPool) Release(name string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if entry, exists := p.mounted[name]; exists {
		entry.lastAccessed = time.Now()
	}
}

// Evict manually evicts a specific disk from the pool
func (p *DiskPool) Evict(ctx context.Context, name string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	entry, exists := p.mounted[name]
	if !exists {
		return nil // Already evicted
	}

	// Unmount the disk
	if err := entry.disk.Unmount(ctx); err != nil {
		return fmt.Errorf("failed to unmount disk %s: %w", name, err)
	}

	delete(p.mounted, name)
	return nil
}

// ForceRemove removes a disk from the pool without unmounting
// This is useful for clearing stale pool entries where the disk is not actually mounted
// It also clears the disk's internal mount state to force a fresh mount on next access
func (p *DiskPool) ForceRemove(name string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// If disk is in pool, clear its mount state before removing
	// This ensures pool.Get() will do a fresh mount instead of trusting stale flags
	if entry, exists := p.mounted[name]; exists {
		disk := entry.disk
		disk.mu.Lock()
		disk.isMounted = false
		disk.mountPath = ""
		disk.mu.Unlock()
	}

	delete(p.mounted, name)
}

// evictLRU evicts the least recently used disk from the pool
// Must be called with lock held
func (p *DiskPool) evictLRU(ctx context.Context) error {
	if len(p.mounted) == 0 {
		return nil
	}

	// Find the least recently used disk
	var oldestName string
	var oldestTime time.Time
	first := true

	for name, entry := range p.mounted {
		if first || entry.lastAccessed.Before(oldestTime) {
			oldestName = name
			oldestTime = entry.lastAccessed
			first = false
		}
	}

	// Unmount it
	entry := p.mounted[oldestName]
	if err := entry.disk.Unmount(ctx); err != nil {
		return fmt.Errorf("failed to unmount LRU disk %s: %w", oldestName, err)
	}

	delete(p.mounted, oldestName)
	return nil
}

// Stats returns statistics about the pool
func (p *DiskPool) Stats() PoolStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := PoolStats{
		MaxMounted:     p.maxMounted,
		CurrentMounted: len(p.mounted),
		Disks:          make([]DiskPoolEntry, 0, len(p.mounted)),
	}

	for name, entry := range p.mounted {
		var mountPath string
		if entry.disk != nil {
			mountPath = entry.disk.MountPath()
		}
		stats.Disks = append(stats.Disks, DiskPoolEntry{
			Name:         name,
			LastAccessed: entry.lastAccessed,
			AccessCount:  entry.accessCount,
			MountPath:    mountPath,
		})
	}

	return stats
}

// Close shuts down the pool and unmounts all disks
func (p *DiskPool) Close() error {
	// Unmount all disks
	ctx := context.Background()
	p.mu.Lock()
	defer p.mu.Unlock()

	var lastErr error
	for name, entry := range p.mounted {
		if err := entry.disk.Unmount(ctx); err != nil {
			fmt.Printf("warning: failed to unmount disk %s during pool shutdown: %v\n", name, err)
			lastErr = err
		}
	}

	p.mounted = make(map[string]*poolEntry)
	return lastErr
}

// PoolStats contains statistics about the disk pool
type PoolStats struct {
	MaxMounted     int             // Maximum number of mounted disks
	CurrentMounted int             // Current number of mounted disks
	Disks          []DiskPoolEntry // Information about mounted disks
}

// DiskPoolEntry contains information about a disk in the pool
type DiskPoolEntry struct {
	Name         string    // Disk name
	LastAccessed time.Time // Last access time
	AccessCount  int64     // Number of accesses
	MountPath    string    // Mount path
}
