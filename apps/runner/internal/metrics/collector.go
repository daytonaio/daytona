/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

package metrics

import (
	"container/ring"
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
)

// CPUSnapshot represents a point-in-time CPU measurement
type CPUSnapshot struct {
	timestamp  time.Time
	cpuPercent float64
}

// Metrics holds runner metrics
type Metrics struct {
	CPULoadAverage        float32
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

	// CPU usage - ring buffer for sliding window
	cpuRing  *ring.Ring
	cpuMutex sync.RWMutex

	// For now, allocations are tracked manually
	// In a real implementation, these would query container runtime
	allocatedCPU       float32
	allocatedMemoryGiB float32
	allocatedDiskGiB   float32
	snapshotCount      float32
}

// NewCollector creates a new metrics collector
func NewCollector(logger *slog.Logger, windowSize int) *Collector {
	if windowSize <= 0 {
		// Default to size 60
		windowSize = 60
	}

	return &Collector{
		log:     logger.With(slog.String("component", "metrics")),
		cpuRing: ring.New(windowSize),
	}
}

// Start begins needed metrics collection processes
func (c *Collector) Start(ctx context.Context) {
	go c.snapshotCPUUsage(ctx)
}

// Collect gathers current system metrics
func (c *Collector) Collect(ctx context.Context) *Metrics {
	metrics := &Metrics{
		AllocatedCPU:       c.allocatedCPU,
		AllocatedMemoryGiB: c.allocatedMemoryGiB,
		AllocatedDiskGiB:   c.allocatedDiskGiB,
		SnapshotCount:      c.snapshotCount,
	}

	// Collect CPU count
	cpuCount, err := cpu.CountsWithContext(ctx, true)
	if err != nil {
		c.log.Warn("Failed to collect CPU count", slog.Any("error", err))
	} else {
		metrics.TotalCPU = float32(cpuCount)
	}

	// Collect CPU usage
	cpuUsage, err := c.collectCPUUsageAverage()
	if err != nil {
		c.log.Warn("Failed to collect CPU metrics", slog.Any("error", err))
	} else {
		metrics.CPUUsagePercentage = float32(cpuUsage)
	}

	// Update CPU load averages
	loadAvg, err := load.Avg()
	if err != nil {
		c.log.Warn("Failed to collect CPU load averages", slog.Any("error", err))
	} else {
		metrics.CPULoadAverage = float32(loadAvg.Load15) / float32(cpuCount)
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

	return metrics
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

// snapshotCPUUsage runs in a background goroutine, continuously monitoring CPU usage
func (c *Collector) snapshotCPUUsage(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			c.log.Info("CPU usage snapshotting stopped")
			return
		default:
			// Add a new CPU snapshot to the ring buffer
			// Use 1-second averaging for accurate CPU measurement
			cpuPercent, err := cpu.PercentWithContext(ctx, 1*time.Second, false)
			if err != nil {
				c.log.Warn("Failed to collect next CPU usage ring", slog.Any("error", err))
				return
			}

			c.cpuMutex.Lock()
			c.cpuRing.Value = CPUSnapshot{
				timestamp:  time.Now(),
				cpuPercent: cpuPercent[0],
			}
			c.cpuRing = c.cpuRing.Next()
			c.cpuMutex.Unlock()
		}
	}
}

// collectCPUUsageAverage calculates the average CPU usage from the ring buffer
func (c *Collector) collectCPUUsageAverage() (float64, error) {
	var total float64
	var count int

	c.cpuMutex.RLock()
	defer c.cpuMutex.RUnlock()

	c.cpuRing.Do(func(x interface{}) {
		if x != nil {
			snapshot := x.(CPUSnapshot)
			total += snapshot.cpuPercent
			count++
		}
	})

	if count == 0 {
		return -1.0, fmt.Errorf("no CPU usage data available")
	}

	return total / float64(count), nil
}
