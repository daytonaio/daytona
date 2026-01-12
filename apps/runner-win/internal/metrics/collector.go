/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

package metrics

import (
	"context"
	"log/slog"

	"github.com/daytonaio/runner-win/pkg/libvirt"
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
	// For now, allocations are tracked manually
	// In a real implementation, these would query container runtime
	allocatedCPU       float32
	allocatedMemoryGiB float32
	allocatedDiskGiB   float32
	snapshotCount      float32
	// localMode indicates if the runner is running in local mode (vs remote SSH mode)
	// When not in local mode, system metrics are collected from the remote host via SSH
	localMode bool
	// libvirtClient is used to collect metrics from remote hosts
	libvirtClient *libvirt.LibVirt
}

// NewCollector creates a new metrics collector
// localMode should be true when the runner is connecting to a local libvirt instance
// When false (remote mode), system metrics will be collected from the remote host via SSH
func NewCollector(logger *slog.Logger, localMode bool, libvirtClient *libvirt.LibVirt) *Collector {
	return &Collector{
		log:           logger.With(slog.String("component", "metrics")),
		localMode:     localMode,
		libvirtClient: libvirtClient,
	}
}

// Collect gathers current system metrics
// In local mode, metrics are collected from the local system using gopsutil
// In remote mode, metrics are collected from the remote libvirt host via SSH
func (c *Collector) Collect(ctx context.Context) (*Metrics, error) {
	metrics := &Metrics{
		AllocatedCPU:       c.allocatedCPU,
		AllocatedMemoryGiB: c.allocatedMemoryGiB,
		AllocatedDiskGiB:   c.allocatedDiskGiB,
		SnapshotCount:      c.snapshotCount,
	}

	if c.localMode {
		c.collectLocalMetrics(ctx, metrics)
	} else {
		c.collectRemoteMetrics(ctx, metrics)
	}

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

// collectRemoteMetrics collects system metrics from the remote libvirt host via SSH
func (c *Collector) collectRemoteMetrics(ctx context.Context, metrics *Metrics) {
	c.log.Debug("Collecting system metrics from remote libvirt host")

	if c.libvirtClient == nil {
		c.log.Warn("No libvirt client available, returning -1 for metrics")
		c.setUnavailableMetrics(metrics)
		return
	}

	remoteMetrics, err := c.libvirtClient.GetRemoteMetrics(ctx)
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

// UpdateAllocations updates the allocation metrics
// These would typically be called when sandboxes are created/destroyed
func (c *Collector) UpdateAllocations(cpu, memoryGiB, diskGiB, snapshots float32) {
	c.allocatedCPU = cpu
	c.allocatedMemoryGiB = memoryGiB
	c.allocatedDiskGiB = diskGiB
	c.snapshotCount = snapshots
}

// IncrementAllocations increments the allocation metrics
func (c *Collector) IncrementAllocations(cpu, memoryGiB, diskGiB float32) {
	c.allocatedCPU += cpu
	c.allocatedMemoryGiB += memoryGiB
	c.allocatedDiskGiB += diskGiB
}

// DecrementAllocations decrements the allocation metrics
func (c *Collector) DecrementAllocations(cpu, memoryGiB, diskGiB float32) {
	c.allocatedCPU -= cpu
	c.allocatedMemoryGiB -= memoryGiB
	c.allocatedDiskGiB -= diskGiB

	// Ensure non-negative
	if c.allocatedCPU < 0 {
		c.allocatedCPU = 0
	}
	if c.allocatedMemoryGiB < 0 {
		c.allocatedMemoryGiB = 0
	}
	if c.allocatedDiskGiB < 0 {
		c.allocatedDiskGiB = 0
	}
}

// IncrementSnapshots increments the snapshot count
func (c *Collector) IncrementSnapshots() {
	c.snapshotCount++
}

// DecrementSnapshots decrements the snapshot count
func (c *Collector) DecrementSnapshots() {
	if c.snapshotCount > 0 {
		c.snapshotCount--
	}
}
