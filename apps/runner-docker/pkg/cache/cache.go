// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package cache

import (
	"context"
	"sync"
	"time"

	pb "github.com/daytonaio/runner-docker/gen/pb/runner/v1"
)

type CacheData struct {
	SandboxState      pb.SandboxState
	BackupState       pb.BackupState
	BackupErrorReason *string
	DestructionTime   *time.Time
}

type IRunnerCache interface {
	SetSandboxState(ctx context.Context, sandboxId string, state pb.SandboxState)
	SetBackupState(ctx context.Context, sandboxId string, state pb.BackupState, err error)

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
		}
	} else {
		data.SandboxState = state
	}

	c.cache[sandboxId] = data
}

func (c *InMemoryRunnerCache) SetBackupState(ctx context.Context, sandboxId string, state pb.BackupState, err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	data, ok := c.cache[sandboxId]
	if !ok {
		backupErrorReason := ""
		if err != nil {
			backupErrorReason = err.Error()
		}
		data = &CacheData{
			SandboxState:      pb.SandboxState_SANDBOX_STATE_UNSPECIFIED,
			BackupState:       state,
			BackupErrorReason: &backupErrorReason,
			DestructionTime:   nil,
		}
	} else {
		data.BackupState = state
		if err != nil {
			errorReason := err.Error()
			data.BackupErrorReason = &errorReason
		}
	}

	c.cache[sandboxId] = data
}

func (c *InMemoryRunnerCache) Set(ctx context.Context, sandboxId string, data CacheData) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cache[sandboxId] = &CacheData{
		SandboxState:    data.SandboxState,
		BackupState:     data.BackupState,
		DestructionTime: data.DestructionTime,
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
