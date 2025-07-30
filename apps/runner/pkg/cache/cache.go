// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package cache

import (
	"context"
	"sync"
	"time"

	"github.com/daytonaio/runner/pkg/models"
	"github.com/daytonaio/runner/pkg/models/enums"
)

type IRunnerCache interface {
	SetSandboxState(ctx context.Context, sandboxId string, state enums.SandboxState)
	SetBackupState(ctx context.Context, sandboxId string, state enums.BackupState)
	SetSystemMetrics(ctx context.Context, metrics models.SystemMetrics)
	GetSystemMetrics(ctx context.Context) *models.SystemMetrics

	Set(ctx context.Context, sandboxId string, data models.CacheData)
	Get(ctx context.Context, sandboxId string) *models.CacheData
	Remove(ctx context.Context, sandboxId string)
	List(ctx context.Context) []string
	Cleanup(ctx context.Context)
}

type InMemoryRunnerCacheConfig struct {
	Cache         map[string]*models.CacheData
	RetentionDays int
}

type InMemoryRunnerCache struct {
	mutex         sync.RWMutex
	cache         map[string]*models.CacheData
	retentionDays int
}

func NewInMemoryRunnerCache(config InMemoryRunnerCacheConfig) IRunnerCache {
	retentionDays := config.RetentionDays
	if retentionDays <= 0 {
		retentionDays = 7
	}

	cache := config.Cache
	if cache == nil {
		cache = make(map[string]*models.CacheData)
	}

	return &InMemoryRunnerCache{
		cache:         cache,
		retentionDays: config.RetentionDays,
	}
}

func (c *InMemoryRunnerCache) SetSandboxState(ctx context.Context, sandboxId string, state enums.SandboxState) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	data, ok := c.cache[sandboxId]
	if !ok {
		data = &models.CacheData{
			SandboxState:    state,
			BackupState:     enums.BackupStateNone,
			DestructionTime: nil,
			SystemMetrics:   nil,
		}
	} else {
		data.SandboxState = state
	}

	c.cache[sandboxId] = data
}

func (c *InMemoryRunnerCache) SetBackupState(ctx context.Context, sandboxId string, state enums.BackupState) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	data, ok := c.cache[sandboxId]
	if !ok {
		data = &models.CacheData{
			SandboxState:    enums.SandboxStateUnknown,
			BackupState:     state,
			DestructionTime: nil,
			SystemMetrics:   nil,
		}
	} else {
		data.BackupState = state
	}

	c.cache[sandboxId] = data
}

func (c *InMemoryRunnerCache) SetSystemMetrics(ctx context.Context, metrics models.SystemMetrics) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Store system metrics under a special key
	const systemMetricsKey = "__system_metrics__"

	data, ok := c.cache[systemMetricsKey]
	if !ok {
		data = &models.CacheData{
			SandboxState:    enums.SandboxStateUnknown,
			BackupState:     enums.BackupStateNone,
			DestructionTime: nil,
			SystemMetrics:   &metrics,
		}
	} else {
		data.SystemMetrics = &metrics
	}

	c.cache[systemMetricsKey] = data
}

func (c *InMemoryRunnerCache) GetSystemMetrics(ctx context.Context) *models.SystemMetrics {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	const systemMetricsKey = "__system_metrics__"

	data, ok := c.cache[systemMetricsKey]
	if !ok || data.SystemMetrics == nil {
		return nil
	}

	return data.SystemMetrics
}

func (c *InMemoryRunnerCache) Set(ctx context.Context, sandboxId string, data models.CacheData) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cache[sandboxId] = &models.CacheData{
		SandboxState:    data.SandboxState,
		BackupState:     data.BackupState,
		DestructionTime: data.DestructionTime,
		SystemMetrics:   data.SystemMetrics,
	}
}

func (c *InMemoryRunnerCache) Get(ctx context.Context, sandboxId string) *models.CacheData {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	data, ok := c.cache[sandboxId]
	if !ok {
		data = &models.CacheData{
			SandboxState:    enums.SandboxStateUnknown,
			BackupState:     enums.BackupStateNone,
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
	c.cache[sandboxId] = &models.CacheData{
		SandboxState:    enums.SandboxStateDestroyed,
		BackupState:     enums.BackupStateNone,
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
