// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/daytonaio/runner/pkg/cache"
	"github.com/daytonaio/runner/pkg/docker"
	"github.com/daytonaio/runner/pkg/models"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"

	log "github.com/sirupsen/logrus"
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

// StartMetricsCollection starts a background goroutine that collects metrics every 20 seconds
func (s *MetricsService) StartMetricsCollection(ctx context.Context) {
	go func() {
		// Collect metrics immediately on startup
		_ = s.collectAndCacheMetrics(ctx)

		// Set up ticker for every 20 seconds
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				_ = s.collectAndCacheMetrics(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()
}

// collectAndCacheMetrics collects all metrics and stores them in cache
func (s *MetricsService) collectAndCacheMetrics(ctx context.Context) error {
	// Get current cached metrics to preserve valid values
	metrics := s.GetSystemMetrics(ctx)

	// Get CPU metrics
	cpuPercent, err := cpu.Percent(15*time.Second, false)
	if err == nil {
		metrics.CPUUsage = cpuPercent[0]
	}

	// Get memory metrics
	memory, err := mem.VirtualMemory()
	if err == nil {
		metrics.RAMUsage = (float64(memory.Total-memory.Available) / float64(memory.Total)) * 100
	}

	// Get disk metrics
	diskUsage, err := disk.Usage("/var/lib/docker")
	if err == nil {
		metrics.DiskUsage = diskUsage.UsedPercent
	}

	// Get snapshot count
	images, err := s.docker.ApiClient().ImageList(ctx, image.ListOptions{})
	if err == nil {
		metrics.SnapshotCount = len(images)
	}

	// Get container allocated resources
	s.getAllocatedResources(ctx, metrics)

	// Store in cache with final values
	return s.cache.Set(ctx, systemMetricsKey, *metrics, 2*time.Hour)
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

func (s *MetricsService) getAllocatedResources(ctx context.Context, metrics *models.SystemMetrics) {
	containers, err := s.docker.ApiClient().ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		log.Errorf("Error listing containers when getting allocated resources: %v", err)
		return
	}

	var totalAllocatedCpuMicroseconds int64 = 0 // CPU quota in microseconds per period
	var totalAllocatedMemoryBytes int64 = 0     // Memory in bytes
	var totalAllocatedDiskGB int64 = 0          // Disk in GB

	for _, ctr := range containers {
		cpu, memory, disk, err := s.getContainerAllocatedResources(ctx, ctr.ID)
		if err == nil {
			// For CPU and memory: only count running containers
			if ctr.State == "running" {
				totalAllocatedCpuMicroseconds += cpu
				totalAllocatedMemoryBytes += memory
			}
			// For disk: count all containers (running and stopped)
			totalAllocatedDiskGB += disk
		}
	}

	// Convert to original API units
	totalAllocatedCpuVCPUs := totalAllocatedCpuMicroseconds / 100000           // Convert back to vCPUs
	totalAllocatedMemoryGB := totalAllocatedMemoryBytes / (1024 * 1024 * 1024) // Convert back to GB

	// Only update if values are valid (not 0 or -1)
	// Note: 0 can be valid for allocated resources (no containers running)
	// so we only check for -1
	if totalAllocatedCpuVCPUs != -1 {
		metrics.AllocatedCPU = totalAllocatedCpuVCPUs
	}
	if totalAllocatedMemoryGB != -1 {
		metrics.AllocatedMemory = totalAllocatedMemoryGB
	}
	if totalAllocatedDiskGB != -1 {
		metrics.AllocatedDisk = totalAllocatedDiskGB
	}
}

func (s *MetricsService) getContainerAllocatedResources(ctx context.Context, containerId string) (int64, int64, int64, error) {
	// Inspect the container to get its resource configuration
	containerJSON, err := s.docker.ContainerInspect(ctx, containerId)
	if err != nil {
		return 0, 0, 0, err
	}

	var allocatedCpu int64 = 0
	var allocatedMemory int64 = 0
	var allocatedDisk int64 = 0

	if containerJSON.HostConfig != nil {
		resources := containerJSON.HostConfig.Resources

		// CPU allocation (convert from quota to total microseconds if quota is set)
		if resources.CPUQuota > 0 {
			allocatedCpu = resources.CPUQuota
		}

		// Memory allocation
		if resources.Memory > 0 {
			allocatedMemory = resources.Memory
		}

		// Disk allocation from StorageOpt (assuming xfs filesystem)
		if containerJSON.HostConfig.StorageOpt != nil {
			if sizeStr, exists := containerJSON.HostConfig.StorageOpt["size"]; exists {
				// Parse size string like "10G" and convert to GB
				if diskGB, err := s.parseStorageQuotaGB(sizeStr); err == nil {
					allocatedDisk = diskGB
				}
			}
		}
	}

	return allocatedCpu, allocatedMemory, allocatedDisk, nil
}

func (s *MetricsService) parseStorageQuotaGB(sizeStr string) (int64, error) {
	// Handle size format like "10G" and return the GB value
	if sizeStr == "" {
		return 0, fmt.Errorf("empty size string")
	}

	// Remove any whitespace
	sizeStr = strings.TrimSpace(sizeStr)

	// Check if it ends with 'G' (assuming xfs format)
	if strings.HasSuffix(sizeStr, "G") {
		// Remove the 'G' and parse the number
		numStr := strings.TrimSuffix(sizeStr, "G")

		gb, err := strconv.ParseInt(numStr, 10, 64)
		if err != nil {
			return 0, err
		}

		return gb, nil
	}

	// If it doesn't end with 'G', return 0 (not xfs format)
	return 0, fmt.Errorf("not in expected xfs format (e.g., '10G')")
}
