// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package metrics

import (
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

// Collector implements MetricsCollector using real system metrics
type Collector struct {
	diskPath string
}

// NewCollector creates a new instance of Collector
func NewCollector(diskPath string) *Collector {
	return &Collector{
		diskPath: diskPath,
	}
}

// GetCPUPercentage returns the current CPU usage percentage
func (c *Collector) GetCPUPercentage() (float64, error) {
	percentages, err := cpu.Percent(time.Second, false)
	if err != nil {
		return 0, fmt.Errorf("failed to get CPU percentage: %w", err)
	}
	if len(percentages) == 0 {
		return 0, fmt.Errorf("no CPU percentage data available")
	}
	return percentages[0], nil
}

// GetMemoryPercentage returns the current memory usage percentage
func (c *Collector) GetMemoryPercentage() (float64, error) {
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return 0, fmt.Errorf("failed to get memory info: %w", err)
	}
	return memInfo.UsedPercent, nil
}

// GetDiskPercentage returns the current disk usage percentage
func (c *Collector) GetDiskPercentage() (float64, error) {
	usage, err := disk.Usage(c.diskPath)
	if err != nil {
		return 0, fmt.Errorf("failed to get disk usage for %s: %w", c.diskPath, err)
	}
	return usage.UsedPercent, nil
}
