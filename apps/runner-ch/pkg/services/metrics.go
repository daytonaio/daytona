// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package services

import (
	"context"
	"sync"
	"time"

	"github.com/daytonaio/runner-ch/pkg/cloudhypervisor"
	"github.com/daytonaio/runner-ch/pkg/models"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"

	log "github.com/sirupsen/logrus"
)

type MetricsServiceConfig struct {
	CHClient  *cloudhypervisor.Client
	Interval  time.Duration
	LocalMode bool // true when running with local CH, false for remote SSH connections
}

type MetricsService struct {
	cache     *models.SystemMetrics
	cacheMu   sync.RWMutex
	chClient  *cloudhypervisor.Client
	interval  time.Duration
	localMode bool
}

// NewMetricsService creates a new metrics service instance
// When localMode is false (remote mode), system metrics (CPU, RAM, disk usage)
// are collected from the remote host via SSH
func NewMetricsService(config MetricsServiceConfig) *MetricsService {
	return &MetricsService{
		chClient:  config.CHClient,
		interval:  config.Interval,
		localMode: config.LocalMode,
	}
}

// StartMetricsCollection starts a background goroutine that collects metrics periodically
func (s *MetricsService) StartMetricsCollection(ctx context.Context) {
	go func() {
		// Collect metrics immediately on startup
		_ = s.collectAndCacheMetrics(ctx)

		// Set up ticker for periodic collection
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
		// Remote mode: collect from remote CH host via SSH
		s.collectRemoteSystemMetrics(ctx, metrics)
	}

	// Get snapshot count with timeout
	snapshotCtx, snapshotCancel := context.WithTimeout(ctx, 3*time.Second)
	defer snapshotCancel()
	snapshots, err := s.chClient.ListSnapshots(snapshotCtx)
	if err != nil {
		log.Warnf("Error getting snapshot count: %v", err)
	} else {
		metrics.SnapshotCount = int64(len(snapshots))
	}

	// Get allocated resources from running VMs
	s.getAllocatedResources(ctx, metrics)

	// Update last updated time
	metrics.LastUpdated = time.Now()

	// Store in cache
	s.cacheMu.Lock()
	s.cache = metrics
	s.cacheMu.Unlock()

	return nil
}

// collectLocalSystemMetrics collects system metrics from the local machine
func (s *MetricsService) collectLocalSystemMetrics(ctx context.Context, metrics *models.SystemMetrics) {
	// Get CPU count
	cpuCount, err := cpu.CountsWithContext(ctx, true)
	if err != nil {
		log.Errorf("Error getting CPU count: %v", err)
	} else {
		metrics.TotalCPU = int64(cpuCount)
	}

	// Get CPU usage
	cpuPercent, err := cpu.PercentWithContext(ctx, time.Second, false)
	if err != nil {
		log.Errorf("Error getting CPU metrics: %v", err)
	} else if len(cpuPercent) > 0 {
		metrics.CPUUsage = cpuPercent[0]
	}

	// Get memory metrics
	memory, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		log.Errorf("Error getting memory metrics: %v", err)
	} else {
		metrics.RAMUsage = (float64(memory.Total-memory.Available) / float64(memory.Total)) * 100
		metrics.TotalRAMGiB = float64(memory.Total) / (1024 * 1024 * 1024)
	}

	// Get disk metrics from root filesystem
	diskUsage, err := disk.UsageWithContext(ctx, "/")
	if err != nil {
		log.Errorf("Error getting disk metrics: %v", err)
	} else {
		metrics.DiskUsage = diskUsage.UsedPercent
		metrics.TotalDiskGiB = float64(diskUsage.Total) / (1024 * 1024 * 1024)
	}
}

// collectRemoteSystemMetrics collects system metrics from the remote CH host via SSH
func (s *MetricsService) collectRemoteSystemMetrics(ctx context.Context, metrics *models.SystemMetrics) {
	log.WithField("component", "metrics").Debug("Collecting system metrics from remote Cloud Hypervisor host")

	remoteMetrics, err := s.chClient.GetRemoteMetrics(ctx)
	if err != nil {
		log.WithField("component", "metrics").Warnf("Failed to collect remote metrics, returning -1 values: %v", err)
		metrics.CPUUsage = -1.0
		metrics.RAMUsage = -1.0
		metrics.DiskUsage = -1.0
		metrics.TotalCPU = -1
		metrics.TotalRAMGiB = -1
		metrics.TotalDiskGiB = -1
		return
	}

	metrics.CPUUsage = remoteMetrics.CPUUsagePercent
	metrics.RAMUsage = remoteMetrics.MemoryUsagePercent
	metrics.DiskUsage = remoteMetrics.DiskUsagePercent
	metrics.TotalCPU = int64(remoteMetrics.TotalCPUs)
	metrics.TotalRAMGiB = remoteMetrics.TotalMemoryGiB
	metrics.TotalDiskGiB = remoteMetrics.TotalDiskGiB

	log.Debugf("Remote system metrics: CPU=%.2f%%, RAM=%.2f%%, Disk=%.2f%%, TotalCPU=%d, TotalRAM=%.2fGiB, TotalDisk=%.2fGiB",
		metrics.CPUUsage, metrics.RAMUsage, metrics.DiskUsage, metrics.TotalCPU, metrics.TotalRAMGiB, metrics.TotalDiskGiB)
}

// GetSystemMetrics returns cached metrics if available, otherwise returns defaults
func (s *MetricsService) GetSystemMetrics(ctx context.Context) *models.SystemMetrics {
	s.cacheMu.RLock()
	cached := s.cache
	s.cacheMu.RUnlock()

	if cached == nil {
		// Return default values if no metrics are cached
		return &models.SystemMetrics{
			CPUUsage:        -1.0,
			RAMUsage:        -1.0,
			DiskUsage:       -1.0,
			TotalCPU:        -1,
			TotalRAMGiB:     -1,
			TotalDiskGiB:    -1,
			AllocatedCPU:    0,
			AllocatedMemory: 0,
			AllocatedDisk:   0,
			SnapshotCount:   0,
			LastUpdated:     time.Now(),
		}
	}

	// Return a copy to avoid race conditions
	return &models.SystemMetrics{
		CPUUsage:        cached.CPUUsage,
		RAMUsage:        cached.RAMUsage,
		DiskUsage:       cached.DiskUsage,
		TotalCPU:        cached.TotalCPU,
		TotalRAMGiB:     cached.TotalRAMGiB,
		TotalDiskGiB:    cached.TotalDiskGiB,
		AllocatedCPU:    cached.AllocatedCPU,
		AllocatedMemory: cached.AllocatedMemory,
		AllocatedDisk:   cached.AllocatedDisk,
		SnapshotCount:   cached.SnapshotCount,
		LastUpdated:     cached.LastUpdated,
	}
}

func (s *MetricsService) getAllocatedResources(ctx context.Context, metrics *models.SystemMetrics) {
	// Use a shorter timeout for this operation
	listCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	sandboxes, err := s.chClient.List(listCtx)
	if err != nil {
		log.Warnf("Error listing sandboxes when getting allocated resources: %v", err)
		return
	}

	// Collect sandbox info in parallel with individual timeouts
	type sandboxResult struct {
		info *cloudhypervisor.SandboxInfo
		err  error
	}

	results := make(chan sandboxResult, len(sandboxes))

	for _, sandboxId := range sandboxes {
		go func(id string) {
			// Short timeout per sandbox
			infoCtx, infoCancel := context.WithTimeout(ctx, 3*time.Second)
			defer infoCancel()

			info, err := s.chClient.GetSandboxInfo(infoCtx, id)
			results <- sandboxResult{info: info, err: err}
		}(sandboxId)
	}

	var totalAllocatedCPU int64 = 0
	var totalAllocatedMemoryGiB int64 = 0
	var totalAllocatedDiskGB int64 = 0

	// Collect results
	for i := 0; i < len(sandboxes); i++ {
		select {
		case result := <-results:
			if result.err != nil {
				// Don't log every error, just count disk
				totalAllocatedDiskGB += 20 // Default 20GB per sandbox even on error
				continue
			}
			if result.info != nil {
				// Only count running sandboxes for CPU and memory
				if result.info.State == cloudhypervisor.VmStateRunning {
					totalAllocatedCPU += int64(result.info.Vcpus)
					totalAllocatedMemoryGiB += int64(result.info.MemoryMB / 1024) // Convert MB to GiB
				}
				totalAllocatedDiskGB += 20 // Default 20GB per sandbox
			}
		case <-ctx.Done():
			return
		}
	}

	metrics.AllocatedCPU = totalAllocatedCPU
	metrics.AllocatedMemory = totalAllocatedMemoryGiB
	metrics.AllocatedDisk = totalAllocatedDiskGB
}
