/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

package metrics

import (
	"context"
	"log/slog"

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
}

// NewCollector creates a new metrics collector
func NewCollector(logger *slog.Logger) *Collector {
	return &Collector{
		log: logger.With(slog.String("component", "metrics")),
	}
}

// Collect gathers current system metrics
func (c *Collector) Collect(ctx context.Context) (*Metrics, error) {
	metrics := &Metrics{
		AllocatedCPU:       c.allocatedCPU,
		AllocatedMemoryGiB: c.allocatedMemoryGiB,
		AllocatedDiskGiB:   c.allocatedDiskGiB,
		SnapshotCount:      c.snapshotCount,
	}

	// Collect CPU usage
	cpuPercent, err := cpu.PercentWithContext(ctx, 0, false)
	if err != nil {
		c.log.Warn("Failed to collect CPU metrics", slog.Any("error", err))
	} else if len(cpuPercent) > 0 {
		metrics.CPUUsagePercentage = float32(cpuPercent[0])
	}

	// Collect memory usage
	memStats, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		c.log.Warn("Failed to collect memory metrics", slog.Any("error", err))
	} else {
		metrics.MemoryUsagePercentage = float32(memStats.UsedPercent)
	}

	// Collect disk usage
	diskStats, err := disk.UsageWithContext(ctx, "/")
	if err != nil {
		c.log.Warn("Failed to collect disk metrics", slog.Any("error", err))
	} else {
		metrics.DiskUsagePercentage = float32(diskStats.UsedPercent)
	}

	return metrics, nil
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
