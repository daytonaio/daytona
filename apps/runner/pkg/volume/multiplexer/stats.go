package multiplexer

import (
	"sync"
	"sync/atomic"
	"time"
)

// StatsTracker tracks statistics for the multiplexer daemon
type StatsTracker struct {
	mu sync.RWMutex

	// Volume statistics
	volumeStats map[string]*VolumeStats

	// Global statistics
	totalReads      atomic.Uint64
	totalWrites     atomic.Uint64
	totalBytesRead  atomic.Uint64
	totalBytesWrite atomic.Uint64
	cacheHits       atomic.Uint64
	cacheMisses     atomic.Uint64
}

// VolumeStats tracks per-volume statistics
type VolumeStats struct {
	VolumeID          string
	RegisteredAt      time.Time
	LastAccessedAt    time.Time
	ReadOperations    uint64
	WriteOperations   uint64
	BytesRead         uint64
	BytesWritten      uint64
	CacheHits         uint64
	CacheMisses       uint64
	ActiveFileHandles int32
}

// DaemonStats represents the overall daemon statistics
type DaemonStats struct {
	StartTime       time.Time
	Uptime          time.Duration
	TotalVolumes    int
	ActiveVolumes   int
	TotalReads      uint64
	TotalWrites     uint64
	TotalBytesRead  uint64
	TotalBytesWrite uint64
	CacheHitRate    float64
	VolumeStats     map[string]*VolumeStats
}

// NewStatsTracker creates a new statistics tracker
func NewStatsTracker() *StatsTracker {
	return &StatsTracker{
		volumeStats: make(map[string]*VolumeStats),
	}
}

// VolumeRegistered records a new volume registration
func (s *StatsTracker) VolumeRegistered(volumeID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.volumeStats[volumeID] = &VolumeStats{
		VolumeID:     volumeID,
		RegisteredAt: time.Now(),
	}
}

// VolumeUnregistered records a volume unregistration
func (s *StatsTracker) VolumeUnregistered(volumeID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.volumeStats, volumeID)
}

// Operation records a filesystem operation
func (s *StatsTracker) Operation(op string, volumeID string, bytes int) {
	s.mu.Lock()
	stats, exists := s.volumeStats[volumeID]
	if !exists {
		s.mu.Unlock()
		return
	}
	stats.LastAccessedAt = time.Now()
	s.mu.Unlock()

	switch op {
	case "read":
		s.totalReads.Add(1)
		s.totalBytesRead.Add(uint64(bytes))
		atomic.AddUint64(&stats.ReadOperations, 1)
		atomic.AddUint64(&stats.BytesRead, uint64(bytes))
	case "write":
		s.totalWrites.Add(1)
		s.totalBytesWrite.Add(uint64(bytes))
		atomic.AddUint64(&stats.WriteOperations, 1)
		atomic.AddUint64(&stats.BytesWritten, uint64(bytes))
	}
}

// CacheHit records a cache hit
func (s *StatsTracker) CacheHit(volumeID string, bytes int) {
	s.cacheHits.Add(1)

	s.mu.Lock()
	if stats, exists := s.volumeStats[volumeID]; exists {
		atomic.AddUint64(&stats.CacheHits, 1)
	}
	s.mu.Unlock()
}

// CacheMiss records a cache miss
func (s *StatsTracker) CacheMiss(volumeID string) {
	s.cacheMisses.Add(1)

	s.mu.Lock()
	if stats, exists := s.volumeStats[volumeID]; exists {
		atomic.AddUint64(&stats.CacheMisses, 1)
	}
	s.mu.Unlock()
}

// GetStats returns current statistics
func (s *StatsTracker) GetStats() *DaemonStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Copy volume stats
	volumeStatsCopy := make(map[string]*VolumeStats)
	activeVolumes := 0
	for id, stats := range s.volumeStats {
		statsCopy := *stats
		volumeStatsCopy[id] = &statsCopy
		if stats.ActiveFileHandles > 0 {
			activeVolumes++
		}
	}

	totalCacheRequests := s.cacheHits.Load() + s.cacheMisses.Load()
	cacheHitRate := float64(0)
	if totalCacheRequests > 0 {
		cacheHitRate = float64(s.cacheHits.Load()) / float64(totalCacheRequests)
	}

	return &DaemonStats{
		StartTime:       time.Now(), // This should be stored at daemon start
		Uptime:          time.Since(time.Now()),
		TotalVolumes:    len(s.volumeStats),
		ActiveVolumes:   activeVolumes,
		TotalReads:      s.totalReads.Load(),
		TotalWrites:     s.totalWrites.Load(),
		TotalBytesRead:  s.totalBytesRead.Load(),
		TotalBytesWrite: s.totalBytesWrite.Load(),
		CacheHitRate:    cacheHitRate,
		VolumeStats:     volumeStatsCopy,
	}
}
