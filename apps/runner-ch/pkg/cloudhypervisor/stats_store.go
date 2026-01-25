// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package cloudhypervisor

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// MemoryStatsRecord represents a single memory stats entry
type MemoryStatsRecord struct {
	ID             int64     `json:"id"`
	Timestamp      time.Time `json:"timestamp"`
	SandboxId      string    `json:"sandbox_id"`
	MaxMemoryKiB   uint64    `json:"max_memory_kib"`
	BalloonSizeKiB uint64    `json:"balloon_size_kib"`
	UsedKiB        uint64    `json:"used_kib"`
	AvailableKiB   uint64    `json:"available_kib"`
	BalloonActive  bool      `json:"balloon_active"`
}

// MemoryStatsRecordInput is the input for recording stats
type MemoryStatsRecordInput struct {
	SandboxId      string
	MaxMemoryKiB   uint64
	BalloonSizeKiB uint64
	UsedKiB        uint64
	AvailableKiB   uint64
	BalloonActive  bool
}

// StatsStoreConfig holds configuration for the stats store
type StatsStoreConfig struct {
	DataPath        string        // Path to data directory
	RetentionDays   int           // How long to keep records (default: 7)
	CleanupInterval time.Duration // How often to run cleanup (default: 1h)
	WriteBufferSize int           // Async write buffer size (default: 1000)
}

// statsData holds all stats in memory with periodic persistence
type statsData struct {
	Records  []MemoryStatsRecord `json:"records"`
	NextID   int64               `json:"next_id"`
	LastSave time.Time           `json:"last_save"`
}

// StatsStore provides persistent storage for memory statistics
type StatsStore struct {
	config    StatsStoreConfig
	dataFile  string
	data      *statsData
	dataMu    sync.RWMutex
	writeChan chan MemoryStatsRecordInput
	logger    *log.Entry
	closeOnce sync.Once
	closed    bool
	closeMu   sync.RWMutex
}

// NewStatsStore creates a new stats store with file backend
func NewStatsStore(config StatsStoreConfig) (*StatsStore, error) {
	// Set defaults
	if config.DataPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			homeDir = "/tmp"
		}
		config.DataPath = filepath.Join(homeDir, ".daytona-runner-ch")
	}
	if config.RetentionDays == 0 {
		config.RetentionDays = 7
	}
	if config.CleanupInterval == 0 {
		config.CleanupInterval = 1 * time.Hour
	}
	if config.WriteBufferSize == 0 {
		config.WriteBufferSize = 1000
	}

	// Ensure directory exists
	if err := os.MkdirAll(config.DataPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create stats directory: %w", err)
	}

	dataFile := filepath.Join(config.DataPath, "memory_stats.json")

	store := &StatsStore{
		config:    config,
		dataFile:  dataFile,
		writeChan: make(chan MemoryStatsRecordInput, config.WriteBufferSize),
		logger:    log.WithField("component", "stats_store"),
		data:      &statsData{Records: []MemoryStatsRecord{}, NextID: 1},
	}

	// Load existing data
	if err := store.loadData(); err != nil {
		store.logger.Warnf("Failed to load existing stats: %v (starting fresh)", err)
	}

	return store, nil
}

// loadData loads stats from file
func (s *StatsStore) loadData() error {
	data, err := os.ReadFile(s.dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist yet, that's OK
		}
		return err
	}

	s.dataMu.Lock()
	defer s.dataMu.Unlock()

	return json.Unmarshal(data, s.data)
}

// saveData saves stats to file
func (s *StatsStore) saveData() error {
	s.dataMu.RLock()
	data, err := json.Marshal(s.data)
	s.dataMu.RUnlock()

	if err != nil {
		return err
	}

	// Write to temp file first, then rename (atomic)
	tmpFile := s.dataFile + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return err
	}

	return os.Rename(tmpFile, s.dataFile)
}

// Start begins the async writer and cleanup goroutines
func (s *StatsStore) Start(ctx context.Context) {
	// Start async writer
	go s.asyncWriter(ctx)

	// Start cleanup routine
	go s.cleanupRoutine(ctx)

	// Start periodic save routine
	go s.saveRoutine(ctx)

	s.logger.Infof("Stats store started (path: %s, retention: %d days)", s.dataFile, s.config.RetentionDays)
}

// asyncWriter processes queued writes in the background
func (s *StatsStore) asyncWriter(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			// Drain remaining writes
			s.drainWrites()
			s.saveData()
			return
		case record := <-s.writeChan:
			s.writeRecord(record)
		}
	}
}

// drainWrites processes any remaining writes in the channel
func (s *StatsStore) drainWrites() {
	for {
		select {
		case record := <-s.writeChan:
			s.writeRecord(record)
		default:
			return
		}
	}
}

// writeRecord writes a single record to memory
func (s *StatsStore) writeRecord(input MemoryStatsRecordInput) {
	s.dataMu.Lock()
	defer s.dataMu.Unlock()

	record := MemoryStatsRecord{
		ID:             s.data.NextID,
		Timestamp:      time.Now(),
		SandboxId:      input.SandboxId,
		MaxMemoryKiB:   input.MaxMemoryKiB,
		BalloonSizeKiB: input.BalloonSizeKiB,
		UsedKiB:        input.UsedKiB,
		AvailableKiB:   input.AvailableKiB,
		BalloonActive:  input.BalloonActive,
	}

	s.data.Records = append(s.data.Records, record)
	s.data.NextID++
}

// saveRoutine periodically saves data to disk
func (s *StatsStore) saveRoutine(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.saveData(); err != nil {
				s.logger.Warnf("Failed to save stats: %v", err)
			}
		}
	}
}

// cleanupRoutine periodically removes old records
func (s *StatsStore) cleanupRoutine(ctx context.Context) {
	ticker := time.NewTicker(s.config.CleanupInterval)
	defer ticker.Stop()

	// Run cleanup immediately on start
	s.cleanup()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.cleanup()
		}
	}
}

// cleanup removes records older than retention period
func (s *StatsStore) cleanup() {
	cutoff := time.Now().AddDate(0, 0, -s.config.RetentionDays)

	s.dataMu.Lock()
	defer s.dataMu.Unlock()

	// Filter records
	newRecords := make([]MemoryStatsRecord, 0, len(s.data.Records))
	removed := 0
	for _, r := range s.data.Records {
		if r.Timestamp.After(cutoff) {
			newRecords = append(newRecords, r)
		} else {
			removed++
		}
	}

	if removed > 0 {
		s.data.Records = newRecords
		s.logger.Infof("Cleaned up %d old stats records (older than %v)", removed, cutoff)
	}
}

// RecordMemoryStats queues memory stats for async writing
// This is non-blocking - if buffer is full, the record is dropped
func (s *StatsStore) RecordMemoryStats(stats *VMMemoryStats) {
	s.closeMu.RLock()
	if s.closed {
		s.closeMu.RUnlock()
		return
	}
	s.closeMu.RUnlock()

	record := MemoryStatsRecordInput{
		SandboxId:      stats.SandboxId,
		MaxMemoryKiB:   stats.MaxMemoryKiB,
		BalloonSizeKiB: stats.BalloonSizeKiB,
		UsedKiB:        stats.UsedMemoryKiB(),
		AvailableKiB:   stats.MemAvailableKiB,
		BalloonActive:  stats.IsBalloonDriverActive(),
	}

	select {
	case s.writeChan <- record:
		// Successfully queued
	default:
		// Buffer full, drop the record (non-blocking)
		s.logger.Debug("Stats write buffer full, dropping record")
	}
}

// RecordBatch records multiple stats at once
func (s *StatsStore) RecordBatch(stats map[string]*VMMemoryStats) {
	for _, vmStats := range stats {
		s.RecordMemoryStats(vmStats)
	}
}

// GetMemoryStats retrieves memory stats for a time range
func (s *StatsStore) GetMemoryStats(ctx context.Context, sandboxId string, from, to time.Time) ([]MemoryStatsRecord, error) {
	s.dataMu.RLock()
	defer s.dataMu.RUnlock()

	var records []MemoryStatsRecord
	for _, r := range s.data.Records {
		// Filter by time
		if r.Timestamp.Before(from) || r.Timestamp.After(to) {
			continue
		}
		// Filter by sandbox ID if specified
		if sandboxId != "" && r.SandboxId != sandboxId {
			continue
		}
		records = append(records, r)
	}

	return records, nil
}

// GetAllSandboxIds returns a list of all sandbox IDs with recorded stats
func (s *StatsStore) GetAllSandboxIds(ctx context.Context) ([]string, error) {
	s.dataMu.RLock()
	defer s.dataMu.RUnlock()

	sandboxSet := make(map[string]bool)
	for _, r := range s.data.Records {
		sandboxSet[r.SandboxId] = true
	}

	ids := make([]string, 0, len(sandboxSet))
	for id := range sandboxSet {
		ids = append(ids, id)
	}

	return ids, nil
}

// GetLatestStats returns the most recent stats for each sandbox
func (s *StatsStore) GetLatestStats(ctx context.Context) (map[string]*MemoryStatsRecord, error) {
	s.dataMu.RLock()
	defer s.dataMu.RUnlock()

	stats := make(map[string]*MemoryStatsRecord)

	for i := range s.data.Records {
		r := &s.data.Records[i]
		existing, ok := stats[r.SandboxId]
		if !ok || r.Timestamp.After(existing.Timestamp) {
			stats[r.SandboxId] = r
		}
	}

	return stats, nil
}

// GetStatsCount returns the total number of records
func (s *StatsStore) GetStatsCount(ctx context.Context) (int64, error) {
	s.dataMu.RLock()
	defer s.dataMu.RUnlock()
	return int64(len(s.data.Records)), nil
}

// Close closes the stats store
func (s *StatsStore) Close() error {
	var err error
	s.closeOnce.Do(func() {
		s.closeMu.Lock()
		s.closed = true
		s.closeMu.Unlock()

		close(s.writeChan)
		err = s.saveData()
	})
	return err
}
