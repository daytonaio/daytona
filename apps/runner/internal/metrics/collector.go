/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

package metrics

import (
	"container/ring"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/daytonaio/runner/pkg/common"
	"github.com/daytonaio/runner/pkg/docker"
	"github.com/docker/docker/api/types/container"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
)

// CollectorConfig holds configuration for the metrics collector
type CollectorConfig struct {
	Logger                             *slog.Logger
	Docker                             *docker.DockerClient
	WindowSize                         int
	CPUUsageSnapshotInterval           time.Duration
	AllocatedResourcesSnapshotInterval time.Duration
}

// Collector collects system metrics
type Collector struct {
	docker *docker.DockerClient
	log    *slog.Logger

	// CPU usage - ring buffer for sliding window
	cpuRing  *ring.Ring
	cpuMutex sync.RWMutex

	resourcesMutex      sync.RWMutex
	allocatedCPU        float32
	allocatedMemoryGiB  float32
	allocatedDiskGiB    float32
	startedSandboxCount float32

	// Intervals for snapshotting metrics in seconds
	cpuUsageSnapshotInterval           time.Duration
	allocatedResourcesSnapshotInterval time.Duration
}

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
	StartedSandboxCount   float32
}

// NewCollector creates a new metrics collector
func NewCollector(cfg CollectorConfig) *Collector {
	return &Collector{
		log:                                cfg.Logger.With(slog.String("component", "metrics")),
		docker:                             cfg.Docker,
		cpuRing:                            ring.New(cfg.WindowSize),
		cpuUsageSnapshotInterval:           cfg.CPUUsageSnapshotInterval,
		allocatedResourcesSnapshotInterval: cfg.AllocatedResourcesSnapshotInterval,
	}
}

// Start begins needed metrics collection processes
func (c *Collector) Start(ctx context.Context) {
	go c.snapshotCPUUsage(ctx)
	go c.snapshotAllocatedResources(ctx)
}

// Collect gathers current system metrics
func (c *Collector) Collect(ctx context.Context) (*Metrics, error) {
	timeout := 30 * time.Second
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		select {
		case <-timeoutCtx.Done():
			return nil, errors.New("timeout collecting metrics")
		default:
			metrics, err := c.collect(ctx)
			if err != nil {
				c.log.WarnContext(ctx, "Failed to collect metrics", "error", err)
				time.Sleep(1 * time.Second)
				continue
			}

			return metrics, nil
		}
	}
}

func (c *Collector) collect(ctx context.Context) (*Metrics, error) {
	metrics := &Metrics{}

	// Collect CPU count
	cpuCount, err := cpu.CountsWithContext(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("failed to collect CPU count: %v", err)
	}
	metrics.TotalCPU = float32(cpuCount)

	// Update CPU load averages
	// Make sure that `cpuCount` exists and is greater than 0
	loadAvg, err := load.Avg()
	if err != nil {
		return nil, fmt.Errorf("failed to collect CPU load averages: %v", err)
	}
	if cpuCount <= 0 {
		return nil, errors.New("CPU count must be greater than zero")
	}
	metrics.CPULoadAverage = float32(loadAvg.Load15) / float32(cpuCount)

	// Collect CPU usage
	cpuUsage, err := c.collectCPUUsageAverage()
	if err != nil {
		return nil, fmt.Errorf("failed to collect CPU usage: %v", err)
	}
	metrics.CPUUsagePercentage = float32(cpuUsage)

	// Collect memory usage and total
	memStats, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to collect memory usage: %v", err)
	}
	metrics.MemoryUsagePercentage = float32(memStats.UsedPercent)
	// Convert bytes to GiB (1 GiB = 1024^3 bytes)
	metrics.TotalRAMGiB = float32(memStats.Total) / (1024 * 1024 * 1024)

	// Collect disk usage and total
	diskStats, err := disk.UsageWithContext(ctx, "/var/lib/docker")
	if err != nil {
		return nil, fmt.Errorf("failed to collect disk usage: %v", err)
	}
	metrics.DiskUsagePercentage = float32(diskStats.UsedPercent)
	// Convert bytes to GiB (1 GiB = 1024^3 bytes)
	metrics.TotalDiskGiB = float32(diskStats.Total) / (1024 * 1024 * 1024)

	// Get snapshot count
	info, err := c.docker.ApiClient().Info(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot count: %v", err)
	}
	metrics.SnapshotCount = float32(info.Images)

	c.resourcesMutex.RLock()
	metrics.AllocatedCPU = c.allocatedCPU
	metrics.AllocatedMemoryGiB = c.allocatedMemoryGiB
	metrics.AllocatedDiskGiB = c.allocatedDiskGiB
	metrics.StartedSandboxCount = c.startedSandboxCount
	c.resourcesMutex.RUnlock()

	return metrics, nil
}

// snapshotCPUUsage runs in a background goroutine, continuously monitoring CPU usage
func (c *Collector) snapshotCPUUsage(ctx context.Context) {
	ticker := time.NewTicker(c.cpuUsageSnapshotInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			c.log.InfoContext(ctx, "CPU usage snapshotting stopped")
			return
		case <-ticker.C:
			// Add a new CPU snapshot to the ring buffer
			cpuPercent, err := cpu.PercentWithContext(ctx, 0, false)
			if err != nil {
				c.log.WarnContext(ctx, "Failed to collect next CPU usage ring", "error", err)
				continue
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
			snapshot, ok := x.(CPUSnapshot)
			if !ok {
				return
			}

			total += snapshot.cpuPercent
			count++
		}
	})

	if count == 0 {
		return -1.0, errors.New("CPU metrics not yet available")
	}

	return total / float64(count), nil
}

func (c *Collector) snapshotAllocatedResources(ctx context.Context) {
	ticker := time.NewTicker(c.allocatedResourcesSnapshotInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			c.log.InfoContext(ctx, "Allocated resources snapshotting stopped")
			return
		case <-ticker.C:
			containers, err := c.docker.ApiClient().ContainerList(ctx, container.ListOptions{All: true})
			if err != nil {
				c.log.ErrorContext(ctx, "Error listing containers when getting allocated resources", "error", err)
				continue
			}

			var totalAllocatedCpuMicroseconds float32 = 0 // CPU quota in microseconds per period
			var totalAllocatedMemoryBytes float32 = 0     // Memory in bytes
			var totalAllocatedDiskGB float32 = 0          // Disk in GB
			var startedSandboxCount float32 = 0           // Count of running containers

			for _, ctr := range containers {
				cpu, memory, disk, err := c.getContainerAllocatedResources(ctx, ctr.ID)
				if err != nil {
					c.log.WarnContext(ctx, "Failed to get allocated resources for container", "container_id", ctr.ID, "error", err)
					continue
				}

				// For CPU and memory: only count running containers
				if ctr.State == "running" {
					totalAllocatedCpuMicroseconds += cpu
					totalAllocatedMemoryBytes += memory
					startedSandboxCount++
				}

				// For disk: count all containers (running and stopped)
				totalAllocatedDiskGB += disk
			}

			// Convert to original API units
			c.resourcesMutex.Lock()
			c.allocatedCPU = totalAllocatedCpuMicroseconds / 100000                 // Convert back to vCPUs
			c.allocatedMemoryGiB = totalAllocatedMemoryBytes / (1024 * 1024 * 1024) // Convert back to GB
			c.allocatedDiskGiB = totalAllocatedDiskGB
			c.startedSandboxCount = startedSandboxCount
			c.resourcesMutex.Unlock()
		}
	}
}

func (c *Collector) getContainerAllocatedResources(ctx context.Context, containerId string) (float32, float32, float32, error) {
	// Inspect the container to get its resource configuration
	containerJSON, err := c.docker.ContainerInspect(ctx, containerId)
	if err != nil {
		return 0, 0, 0, err
	}

	if containerJSON.HostConfig == nil {
		return 0, 0, 0, nil
	}

	var allocatedCpu, allocatedMemory, allocatedDisk float32 = 0, 0, 0

	resources := containerJSON.HostConfig.Resources

	if resources.CPUQuota > 0 {
		allocatedCpu = float32(resources.CPUQuota)
	}

	if resources.Memory > 0 {
		allocatedMemory = float32(resources.Memory)
	}

	if containerJSON.HostConfig.StorageOpt == nil {
		return allocatedCpu, allocatedMemory, 0, nil
	}

	// Disk allocation from StorageOpt (assuming xfs filesystem)
	storageGB, err := common.ParseStorageOptSizeGB(containerJSON.HostConfig.StorageOpt)
	if err != nil {
		return allocatedCpu, allocatedMemory, 0, fmt.Errorf("error parsing storage quota for container %s: %v", containerId, err)
	}

	if storageGB > 0 {
		allocatedDisk = float32(storageGB)
	}

	return allocatedCpu, allocatedMemory, allocatedDisk, nil
}
