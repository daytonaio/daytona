// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/daytonaio/runner-win/pkg/libvirt"
	"github.com/daytonaio/runner-win/pkg/models"

	common_cache "github.com/daytonaio/common-go/pkg/cache"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"

	log "github.com/sirupsen/logrus"
)

const systemMetricsKey = "__system_metrics__"

type MetricsServiceConfig struct {
	LibVirt   *libvirt.LibVirt
	Interval  time.Duration
	LocalMode bool // true when running with local libvirt, false for remote SSH connections
}

type MetricsService struct {
	cache     common_cache.ICache[models.SystemMetrics]
	libvirt   *libvirt.LibVirt
	interval  time.Duration
	localMode bool
}

// NewMetricsService creates a new metrics service instance
// When localMode is false (remote mode), system metrics (CPU, RAM, disk usage)
// are not collected as they would represent the runner host, not the remote libvirt host
func NewMetricsService(config MetricsServiceConfig) *MetricsService {
	metricsCache := common_cache.NewMapCache[models.SystemMetrics]()

	return &MetricsService{
		cache:     metricsCache,
		libvirt:   config.LibVirt,
		interval:  config.Interval,
		localMode: config.LocalMode,
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

	// Collect system metrics (CPU, RAM, disk)
	if s.localMode {
		// Local mode: collect from local system using gopsutil
		s.collectLocalSystemMetrics(ctx, metrics)
	} else {
		// Remote mode: collect from remote libvirt host via SSH
		s.collectRemoteSystemMetrics(ctx, metrics)
	}

	// Get snapshot count (this queries libvirt, so it's always valid)
	info, err := s.libvirt.Info(ctx)
	if err != nil {
		log.Errorf("Error getting snapshot count: %v", err)
	} else {
		// For libvirt, we don't track images separately, use total domains
		metrics.SnapshotCount = info.DomainsTotal
	}

	// Get container allocated resources (this queries libvirt, so it's always valid)
	s.getAllocatedResources(ctx, metrics)

	// Store in cache with final values
	return s.cache.Set(ctx, systemMetricsKey, *metrics, 2*time.Hour)
}

// collectLocalSystemMetrics collects system metrics from the local machine
func (s *MetricsService) collectLocalSystemMetrics(ctx context.Context, metrics *models.SystemMetrics) {
	// Get CPU metrics
	cpuPercent, err := cpu.Percent(15*time.Second, false)
	if err != nil {
		log.Errorf("Error getting CPU metrics: %v", err)
	} else {
		metrics.CPUUsage = cpuPercent[0]
	}

	// Get memory metrics
	memory, err := mem.VirtualMemory()
	if err != nil {
		log.Errorf("Error getting memory metrics: %v", err)
	} else {
		metrics.RAMUsage = (float64(memory.Total-memory.Available) / float64(memory.Total)) * 100
	}

	// Get disk metrics from root filesystem
	diskUsage, err := disk.Usage("/")
	if err != nil {
		log.Errorf("Error getting disk metrics: %v", err)
	} else {
		metrics.DiskUsage = diskUsage.UsedPercent
	}
}

// collectRemoteSystemMetrics collects system metrics from the remote libvirt host via SSH
func (s *MetricsService) collectRemoteSystemMetrics(ctx context.Context, metrics *models.SystemMetrics) {
	log.Debug("Collecting system metrics from remote libvirt host")

	remoteMetrics, err := s.libvirt.GetRemoteMetrics(ctx)
	if err != nil {
		log.Warnf("Failed to collect remote metrics, using -1 values: %v", err)
		metrics.CPUUsage = -1.0
		metrics.RAMUsage = -1.0
		metrics.DiskUsage = -1.0
		return
	}

	metrics.CPUUsage = remoteMetrics.CPUUsagePercent
	metrics.RAMUsage = remoteMetrics.MemoryUsagePercent
	metrics.DiskUsage = remoteMetrics.DiskUsagePercent

	log.Debugf("Remote system metrics: CPU=%.2f%%, RAM=%.2f%%, Disk=%.2f%%",
		metrics.CPUUsage, metrics.RAMUsage, metrics.DiskUsage)
}

// GetCachedSystemMetrics returns cached metrics if available, otherwise returns defaults
func (s *MetricsService) GetSystemMetrics(ctx context.Context) *models.SystemMetrics {
	metrics, err := s.cache.Get(ctx, systemMetricsKey)
	if err != nil || metrics == nil {
		// This is expected on first call before metrics are collected
		// Don't log error for "key not found" which is normal
		if err != nil && err.Error() != "key not found" {
			log.Warnf("Error getting system metrics: %v", err)
		}

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
	containers, err := s.libvirt.ContainerList(ctx, libvirt.DomainListOptions{All: true})
	if err != nil {
		log.Errorf("Error listing containers when getting allocated resources: %v", err)
		return
	}

	var totalAllocatedCpuMicroseconds int64 = 0 // CPU quota in microseconds per period
	var totalAllocatedMemoryBytes int64 = 0     // Memory in bytes
	var totalAllocatedDiskGB int64 = 0          // Disk in GB

	for _, ctr := range containers {
		cpu, memory, disk, err := s.getContainerAllocatedResources(ctx, ctr.UUID)
		if err != nil {
			log.Errorf("Error getting allocated resources for container %s: %v", ctr.UUID, err)
		} else {
			// For CPU and memory: only count running containers
			if ctr.State == libvirt.DomainStateRunning {
				totalAllocatedCpuMicroseconds += cpu
				totalAllocatedMemoryBytes += memory
			}
			// For disk: count all containers (running and stopped)
			totalAllocatedDiskGB += disk
		}
	}

	// Convert to original API units
	metrics.AllocatedCPU = totalAllocatedCpuMicroseconds / 100000              // Convert back to vCPUs
	metrics.AllocatedMemory = totalAllocatedMemoryBytes / (1024 * 1024 * 1024) // Convert back to GB
	metrics.AllocatedDisk = totalAllocatedDiskGB
}

func (s *MetricsService) getContainerAllocatedResources(ctx context.Context, containerId string) (int64, int64, int64, error) {
	// Get basic domain info without waiting for IP address (faster)
	domainInfo, err := s.libvirt.ContainerInspectBasic(ctx, containerId)
	if err != nil {
		return 0, 0, 0, err
	}

	var allocatedCpu int64 = 0
	var allocatedMemory int64 = 0
	var allocatedDisk int64 = 0

	// For libvirt domains, get CPU and memory from domain info
	// VCPUs * 100000 microseconds (1 vCPU = 100000 microseconds in CPU quota)
	allocatedCpu = int64(domainInfo.VCPUs) * 100000

	// Memory is in KiB, convert to bytes
	allocatedMemory = int64(domainInfo.Memory) * 1024

	// Disk allocation - for now we'll estimate or set a default
	// TODO: Query actual disk size from domain disk configuration
	allocatedDisk = 10 // Default 10 GB per VM

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
