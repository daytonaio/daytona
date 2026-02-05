/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

package metrics

import (
	"context"
	"log/slog"
	"time"

	"github.com/daytonaio/runner-android/pkg/cuttlefish"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

// Metrics holds runner metrics
type Metrics struct {
	CPUUsagePercentage    float32
	MemoryUsagePercentage float32
	DiskUsagePercentage   float32
	AllocatedCPU          float32
	AllocatedMemoryGiB    float32
	AllocatedDiskGiB      float32
	SnapshotCount         float32
	TotalCPU              float32
	TotalRAMGiB           float32
	TotalDiskGiB          float32
}

// Collector collects system metrics
type Collector struct {
	log *slog.Logger
	// localMode indicates if the runner is running in local mode (vs remote SSH mode)
	// When not in local mode, system metrics are collected from the remote host via SSH
	localMode bool
	// cvdClient is used to collect metrics from remote hosts and get sandbox info
	cvdClient *cuttlefish.Client
}

// NewCollector creates a new metrics collector
// localMode should be true when the runner is connecting to a local Cuttlefish instance
// When false (remote mode), system metrics will be collected from the remote host via SSH
func NewCollector(logger *slog.Logger, localMode bool, cvdClient *cuttlefish.Client) *Collector {
	return &Collector{
		log:       logger.With(slog.String("component", "metrics")),
		localMode: localMode,
		cvdClient: cvdClient,
	}
}

// Collect gathers current system metrics
// In local mode, metrics are collected from the local system using gopsutil
// In remote mode, metrics are collected from the remote Cuttlefish host via SSH
func (c *Collector) Collect(ctx context.Context) (*Metrics, error) {
	metrics := &Metrics{}

	if c.localMode {
		c.collectLocalMetrics(ctx, metrics)
	} else {
		c.collectRemoteMetrics(ctx, metrics)
	}

	// Collect allocated resources from CVD client
	c.collectAllocatedResources(ctx, metrics)

	return metrics, nil
}

// collectLocalMetrics collects system metrics from the local machine
func (c *Collector) collectLocalMetrics(ctx context.Context, metrics *Metrics) {
	// Collect CPU count
	cpuCount, err := cpu.CountsWithContext(ctx, true)
	if err != nil {
		c.log.Warn("Failed to collect CPU count", slog.Any("error", err))
	} else {
		metrics.TotalCPU = float32(cpuCount)
	}

	// Collect CPU usage
	cpuPercent, err := cpu.PercentWithContext(ctx, 0, false)
	if err != nil {
		c.log.Warn("Failed to collect CPU metrics", slog.Any("error", err))
	} else if len(cpuPercent) > 0 {
		metrics.CPUUsagePercentage = float32(cpuPercent[0])
	}

	// Collect memory usage and total
	memStats, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		c.log.Warn("Failed to collect memory metrics", slog.Any("error", err))
	} else {
		metrics.MemoryUsagePercentage = float32(memStats.UsedPercent)
		// Convert bytes to GiB (1 GiB = 1024^3 bytes)
		metrics.TotalRAMGiB = float32(memStats.Total) / (1024 * 1024 * 1024)
	}

	// Collect disk usage and total
	diskStats, err := disk.UsageWithContext(ctx, "/")
	if err != nil {
		c.log.Warn("Failed to collect disk metrics", slog.Any("error", err))
	} else {
		metrics.DiskUsagePercentage = float32(diskStats.UsedPercent)
		// Convert bytes to GiB (1 GiB = 1024^3 bytes)
		metrics.TotalDiskGiB = float32(diskStats.Total) / (1024 * 1024 * 1024)
	}
}

// collectRemoteMetrics collects system metrics from the remote Cuttlefish host via SSH
func (c *Collector) collectRemoteMetrics(ctx context.Context, metrics *Metrics) {
	c.log.Debug("Collecting system metrics from remote Cuttlefish host")

	if c.cvdClient == nil {
		c.log.Warn("No Cuttlefish client available, returning -1 for metrics")
		c.setUnavailableMetrics(metrics)
		return
	}

	remoteMetrics, err := c.cvdClient.GetRemoteMetrics(ctx)
	if err != nil {
		c.log.Warn("Failed to collect remote metrics, returning -1 values", slog.Any("error", err))
		c.setUnavailableMetrics(metrics)
		return
	}

	metrics.CPUUsagePercentage = float32(remoteMetrics.CPUUsagePercent)
	metrics.MemoryUsagePercentage = float32(remoteMetrics.MemoryUsagePercent)
	metrics.DiskUsagePercentage = float32(remoteMetrics.DiskUsagePercent)
	metrics.TotalCPU = float32(remoteMetrics.TotalCPUs)
	metrics.TotalRAMGiB = float32(remoteMetrics.TotalMemoryGiB)
	metrics.TotalDiskGiB = float32(remoteMetrics.TotalDiskGiB)

	c.log.Debug("Remote metrics collected",
		slog.Float64("cpu_usage", float64(metrics.CPUUsagePercentage)),
		slog.Float64("mem_usage", float64(metrics.MemoryUsagePercentage)),
		slog.Float64("disk_usage", float64(metrics.DiskUsagePercentage)),
		slog.Float64("total_cpu", float64(metrics.TotalCPU)),
		slog.Float64("total_ram_gib", float64(metrics.TotalRAMGiB)),
		slog.Float64("total_disk_gib", float64(metrics.TotalDiskGiB)))
}

// setUnavailableMetrics sets all system metrics to -1 to indicate unavailability
func (c *Collector) setUnavailableMetrics(metrics *Metrics) {
	metrics.CPUUsagePercentage = -1
	metrics.MemoryUsagePercentage = -1
	metrics.DiskUsagePercentage = -1
	metrics.TotalCPU = -1
	metrics.TotalRAMGiB = -1
	metrics.TotalDiskGiB = -1
}

// collectAllocatedResources collects allocated resources from running instances
func (c *Collector) collectAllocatedResources(ctx context.Context, metrics *Metrics) {
	if c.cvdClient == nil {
		c.log.Warn("No Cuttlefish client available for collecting allocated resources")
		return
	}

	// Use a short timeout for listing sandboxes
	listCtx, listCancel := context.WithTimeout(ctx, 3*time.Second)
	defer listCancel()

	// Get list of sandboxes with their info
	sandboxes, err := c.cvdClient.ListWithInfo(listCtx)
	if err != nil {
		c.log.Warn("Failed to list sandboxes for allocated resources", slog.Any("error", err))
		return
	}

	var totalAllocatedCPU float32 = 0
	var totalAllocatedMemoryGiB float32 = 0
	var totalAllocatedDiskGiB float32 = 0

	for _, info := range sandboxes {
		if info != nil {
			// Only count running instances for CPU and memory
			if info.State == cuttlefish.InstanceStateRunning {
				totalAllocatedCPU += float32(info.Vcpus)
				totalAllocatedMemoryGiB += float32(info.MemoryMB) / 1024 // Convert MB to GiB
			}
			totalAllocatedDiskGiB += 20 // Assume 20GB per instance
		}
	}

	metrics.AllocatedCPU = totalAllocatedCPU
	metrics.AllocatedMemoryGiB = totalAllocatedMemoryGiB
	metrics.AllocatedDiskGiB = totalAllocatedDiskGiB
	metrics.SnapshotCount = 0 // Cuttlefish doesn't have the same snapshot concept

	c.log.Debug("Collected allocated resources",
		slog.Float64("allocated_cpu", float64(metrics.AllocatedCPU)),
		slog.Float64("allocated_memory_gib", float64(metrics.AllocatedMemoryGiB)),
		slog.Float64("allocated_disk_gib", float64(metrics.AllocatedDiskGiB)),
		slog.Float64("snapshot_count", float64(metrics.SnapshotCount)))
}
