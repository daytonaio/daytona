// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package cache

import (
	"context"
	"sync"
	"time"

	pb "github.com/daytonaio/runner-docker/gen/pb/runner/v1"
)

type SystemMetrics struct {
	CPUUsage        float64   `json:"cpu_usage"`
	RAMUsage        float64   `json:"ram_usage"`
	DiskUsage       float64   `json:"disk_usage"`
	AllocatedCPU    int64     `json:"allocated_cpu"`
	AllocatedMemory int64     `json:"allocated_memory"`
	AllocatedDisk   int64     `json:"allocated_disk"`
	SnapshotCount   int64     `json:"snapshot_count"`
	LastUpdated     time.Time `json:"last_updated"`
}

type CacheData struct {
	SandboxState    pb.SandboxState
	BackupState     pb.BackupState
	DestructionTime *time.Time
	SystemMetrics   *SystemMetrics
}

type IRunnerCache interface {
	SetSandboxState(ctx context.Context, sandboxId string, state pb.SandboxState)
	SetBackupState(ctx context.Context, sandboxId string, state pb.BackupState)
	SetSystemMetrics(ctx context.Context, metrics SystemMetrics)
	GetSystemMetrics(ctx context.Context) *SystemMetrics

	Set(ctx context.Context, sandboxId string, data CacheData)
	Get(ctx context.Context, sandboxId string) *CacheData
	Remove(ctx context.Context, sandboxId string)
	List(ctx context.Context) []string
	Cleanup(ctx context.Context)
}

type InMemoryRunnerCacheConfig struct {
	Cache         map[string]*CacheData
	RetentionDays int
}

type InMemoryRunnerCache struct {
	mutex         sync.RWMutex
	cache         map[string]*CacheData
	retentionDays int
}

func NewInMemoryRunnerCache(config InMemoryRunnerCacheConfig) IRunnerCache {
	retentionDays := config.RetentionDays
	if retentionDays <= 0 {
		retentionDays = 7
	}

	cache := config.Cache
	if cache == nil {
		cache = make(map[string]*CacheData)
	}

	return &InMemoryRunnerCache{
		cache:         cache,
		retentionDays: config.RetentionDays,
	}
}

func (c *InMemoryRunnerCache) SetSandboxState(ctx context.Context, sandboxId string, state pb.SandboxState) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	data, ok := c.cache[sandboxId]
	if !ok {
		data = &CacheData{
			SandboxState:    state,
			BackupState:     pb.BackupState_BACKUP_STATE_UNSPECIFIED,
			DestructionTime: nil,
			SystemMetrics:   nil,
		}
	} else {
		data.SandboxState = state
	}

	c.cache[sandboxId] = data
}

func (c *InMemoryRunnerCache) SetBackupState(ctx context.Context, sandboxId string, state pb.BackupState) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	data, ok := c.cache[sandboxId]
	if !ok {
		data = &CacheData{
			SandboxState:    pb.SandboxState_SANDBOX_STATE_UNSPECIFIED,
			BackupState:     state,
			DestructionTime: nil,
			SystemMetrics:   nil,
		}
	} else {
		data.BackupState = state
	}

	c.cache[sandboxId] = data
}

func (c *InMemoryRunnerCache) SetSystemMetrics(ctx context.Context, metrics SystemMetrics) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Store system metrics under a special key
	const systemMetricsKey = "__system_metrics__"

	data, ok := c.cache[systemMetricsKey]
	if !ok {
		data = &CacheData{
			SandboxState:    pb.SandboxState_SANDBOX_STATE_UNSPECIFIED,
			BackupState:     pb.BackupState_BACKUP_STATE_UNSPECIFIED,
			DestructionTime: nil,
			SystemMetrics:   &metrics,
		}
	} else {
		data.SystemMetrics = &metrics
	}

	c.cache[systemMetricsKey] = data
}

func (c *InMemoryRunnerCache) GetSystemMetrics(ctx context.Context) *SystemMetrics {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	const systemMetricsKey = "__system_metrics__"

	data, ok := c.cache[systemMetricsKey]
	if !ok || data.SystemMetrics == nil {
		return nil
	}

	return data.SystemMetrics
}

func (c *InMemoryRunnerCache) Set(ctx context.Context, sandboxId string, data CacheData) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cache[sandboxId] = &CacheData{
		SandboxState:    data.SandboxState,
		BackupState:     data.BackupState,
		DestructionTime: data.DestructionTime,
		SystemMetrics:   data.SystemMetrics,
	}
}

func (c *InMemoryRunnerCache) Get(ctx context.Context, sandboxId string) *CacheData {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	data, ok := c.cache[sandboxId]
	if !ok {
		data = &CacheData{
			SandboxState:    pb.SandboxState_SANDBOX_STATE_UNSPECIFIED,
			BackupState:     pb.BackupState_BACKUP_STATE_UNSPECIFIED,
			DestructionTime: nil,
			SystemMetrics:   nil,
		}
	}

	return data
}

func (c *InMemoryRunnerCache) Remove(ctx context.Context, sandboxId string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	destructionTime := time.Now().Add(time.Duration(c.retentionDays) * 24 * time.Hour)
	c.cache[sandboxId] = &CacheData{
		SandboxState:    pb.SandboxState_SANDBOX_STATE_DESTROYED,
		BackupState:     pb.BackupState_BACKUP_STATE_UNSPECIFIED,
		DestructionTime: &destructionTime,
		SystemMetrics:   nil,
	}
}

func (c *InMemoryRunnerCache) List(ctx context.Context) []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	keys := make([]string, 0, len(c.cache))
	for k := range c.cache {
		keys = append(keys, k)
	}

	return keys
}

func (c *InMemoryRunnerCache) Cleanup(ctx context.Context) {
	go func() {
		// Run cleanup every hour
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				c.cleanupExpiredEntries()
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (c *InMemoryRunnerCache) cleanupExpiredEntries() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	for id, data := range c.cache {
		if data.DestructionTime != nil && (now.After(*data.DestructionTime) || now.Equal(*data.DestructionTime)) {
			delete(c.cache, id)
		}
	}
}
