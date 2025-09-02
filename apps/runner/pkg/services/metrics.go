// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package services

import (
	"context"
	"time"

	"github.com/daytonaio/runner/pkg/cache"
	"github.com/daytonaio/runner/pkg/docker"
	"github.com/daytonaio/runner/pkg/models"
	"github.com/docker/docker/api/types/image"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
)

const systemMetricsKey = "__system_metrics__"

type MetricsServiceConfig struct {
	Cache    cache.ICache[models.SystemMetrics]
	Docker   *docker.DockerClient
	Interval time.Duration
}

type MetricsService struct {
	cache    cache.ICache[models.SystemMetrics]
	docker   *docker.DockerClient
	interval time.Duration
}

// NewPrometheusParser creates a new parser instance
func NewMetricsService(config MetricsServiceConfig) *MetricsService {
	return &MetricsService{
		cache:    config.Cache,
		docker:   config.Docker,
		interval: config.Interval,
	}
}

func (s *MetricsService) GetCPUMetrics(ctx context.Context, metrics *models.SystemMetrics) {
	cpuPercent, err := cpu.Percent(15*time.Second, false)
	if err == nil {
		metrics.CPUUsage = cpuPercent[0]
	}

	cpuInfo, err := cpu.Info()
	if err == nil {
		metrics.AllocatedCPU = int64(len(cpuInfo))
	}
}

func (s *MetricsService) GetMemoryMetrics(ctx context.Context, metrics *models.SystemMetrics) {
	memory, err := mem.VirtualMemory()
	if err == nil {
		metrics.RAMUsage = (float64(memory.Total-memory.Available) / float64(memory.Total)) * 100
		metrics.AllocatedMemory = int64(memory.Total / (1024 * 1024 * 1024))
	}
}

func (s *MetricsService) GetDiskMetrics(ctx context.Context, metrics *models.SystemMetrics) {
	diskUsage, err := disk.Usage("/var/lib/docker")
	if err == nil {
		metrics.DiskUsage = diskUsage.UsedPercent
		metrics.AllocatedDisk = int64(diskUsage.Total / (1024 * 1024 * 1024))
	}
}

func (s *MetricsService) GetSnapshotCount(ctx context.Context, metrics *models.SystemMetrics) {
	images, err := s.docker.ApiClient().ImageList(ctx, image.ListOptions{})
	if err == nil {
		metrics.SnapshotCount = len(images)
	}
}

// GetCachedSystemMetrics returns cached metrics if available, otherwise returns defaults
func (s *MetricsService) GetSystemMetrics(ctx context.Context) *models.SystemMetrics {
	metrics, err := s.cache.Get(ctx, systemMetricsKey)
	if err != nil || metrics == nil {
		// Return default values if no metrics are cached
		return &models.SystemMetrics{
			CPUUsage:        -1.0,
			RAMUsage:        -1.0,
			DiskUsage:       -1.0,
			AllocatedCPU:    -1,
			AllocatedMemory: -1,
			AllocatedDisk:   -1,
			SnapshotCount:   -1,
			LastUpdated:     time.Now(),
		}
	}

	return metrics
}

// CollectAndCacheMetrics collects all metrics and stores them in cache
func (s *MetricsService) CollectAndCacheMetrics(ctx context.Context) error {
	// Get current cached metrics to preserve valid values
	metrics := s.GetSystemMetrics(ctx)

	// Get CPU metrics
	s.GetCPUMetrics(ctx, metrics)

	// Get memory metrics
	s.GetMemoryMetrics(ctx, metrics)

	// Get disk metrics
	s.GetDiskMetrics(ctx, metrics)

	// Get snapshot count
	s.GetSnapshotCount(ctx, metrics)

	// Store in cache with final values
	return s.cache.Set(ctx, systemMetricsKey, *metrics, 2*time.Hour)
}

// StartMetricsCollection starts a background goroutine that collects metrics every 20 seconds
func (s *MetricsService) StartMetricsCollection(ctx context.Context) {
	go func() {
		// Collect metrics immediately on startup
		_ = s.CollectAndCacheMetrics(ctx)

		// Set up ticker for every 20 seconds
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				_ = s.CollectAndCacheMetrics(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()
}
