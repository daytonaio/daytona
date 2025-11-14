// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package services

import (
	"container/ring"
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/daytonaio/runner/pkg/docker"
	"github.com/daytonaio/runner/pkg/models"

	"github.com/docker/docker/api/types/container"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"

	log "github.com/sirupsen/logrus"
)

// CPUSnapshot represents a point-in-time CPU measurement
type CPUSnapshot struct {
	timestamp  time.Time
	cpuPercent float64
}

// MetricsServiceConfig holds configuration for the metrics service
type MetricsServiceConfig struct {
	Docker          *docker.DockerClient
	TimeoutInterval string
	WindowSize      int // CPU metrics window size in seconds
}

// MetricsService provides system metrics with proper encapsulation
type MetricsService struct {
	docker          *docker.DockerClient
	timeoutInterval time.Duration

	// CPU usage metrics - ring buffer for sliding window
	cpuRing  *ring.Ring
	cpuMutex sync.RWMutex

	// Other metrics - cached values
	cpuLoadAvg      float64
	ramUsage        float64
	diskUsage       float64
	allocatedCPU    int64
	allocatedMemory int64
	allocatedDisk   int64
	snapshotCount   int
	lastUpdated     time.Time
	lastError       error
	lastErrorTime   time.Time
	otherMutex      sync.RWMutex
}

// NewMetricsService creates a new metrics service instance
func NewMetricsService(config MetricsServiceConfig) *MetricsService {
	windowSize := config.WindowSize
	if windowSize <= 0 {
		windowSize = 60 // Default to size 60
	}

	timeoutInterval, err := time.ParseDuration(config.TimeoutInterval)
	if err != nil {
		log.Errorf("Error parsing timeout interval: %v - using default of 5s", err)
		timeoutInterval = 5 * time.Second
	}

	return &MetricsService{
		docker:          config.Docker,
		timeoutInterval: timeoutInterval,
		cpuRing:         ring.New(windowSize),
	}
}

// Start begins metrics collection in background goroutines
func (s *MetricsService) Start(ctx context.Context) {
	go s.collectCPUUsageMetrics(ctx)
	go s.collectOtherMetrics(ctx)
}

// GetMetrics returns all current system metrics
func (s *MetricsService) GetMetrics() (*models.SystemMetrics, error) {
	for {
		select {
		case <-time.After(s.timeoutInterval):
			s.otherMutex.RLock()
			defer s.otherMutex.RUnlock()
			if s.lastError != nil && time.Since(s.lastErrorTime) < time.Minute {
				return nil, fmt.Errorf("error getting metrics: %w", s.lastError)
			}
			return nil, fmt.Errorf("timeout waiting for metrics")
		default:
			metrics := s.getMetrics()
			if metrics != nil {
				return metrics, nil
			}
		}
	}
}

// getMetrics returns all current system metrics
func (s *MetricsService) getMetrics() *models.SystemMetrics {
	// Get CPU metrics (requires read lock)
	s.cpuMutex.RLock()
	cpuUsage, err := s.calculateCPUUsageAverage()
	if err != nil {
		s.otherMutex.Lock()
		s.lastError = err
		s.lastErrorTime = time.Now()
		s.otherMutex.Unlock()

		return nil
	}
	s.cpuMutex.RUnlock()

	// Get other metrics (requires read lock)
	s.otherMutex.RLock()
	defer s.otherMutex.RUnlock()

	// Return error if it's within the last minute
	if s.lastError != nil && time.Since(s.lastErrorTime) < time.Minute {
		return nil
	}

	return &models.SystemMetrics{
		CPUUsage:        cpuUsage,
		CPULoadAvg:      s.cpuLoadAvg,
		RAMUsage:        s.ramUsage,
		DiskUsage:       s.diskUsage,
		AllocatedCPU:    s.allocatedCPU,
		AllocatedMemory: s.allocatedMemory,
		AllocatedDisk:   s.allocatedDisk,
		SnapshotCount:   s.snapshotCount,
		LastUpdated:     s.lastUpdated,
	}
}

// collectCPUUsageMetrics runs in a background goroutine, continuously monitoring CPU
func (s *MetricsService) collectCPUUsageMetrics(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			s.addCPUPercentSnapshot()
		}
	}
}

// collectOtherMetrics runs in a background goroutine, collecting other metrics every 20 seconds
func (s *MetricsService) collectOtherMetrics(ctx context.Context) {
	// Collect immediately on startup
	s.updateOtherMetrics(ctx)

	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.updateOtherMetrics(ctx)
		case <-ctx.Done():
			return
		}
	}
}

// addCPUPercentSnapshot adds a new CPU snapshot to the ring buffer
func (s *MetricsService) addCPUPercentSnapshot() {
	// Use 1-second averaging for accurate CPU measurement
	cpuPercent, err := cpu.Percent(1*time.Second, false)
	if err != nil {
		log.Errorf("Error reading CPU percentage: %v", err)
		return
	}

	s.cpuMutex.Lock()
	defer s.cpuMutex.Unlock()

	s.cpuRing.Value = CPUSnapshot{
		timestamp:  time.Now(),
		cpuPercent: cpuPercent[0],
	}
	s.cpuRing = s.cpuRing.Next()
}

// calculateCPUUsageAverage calculates the average CPU usage from the ring buffer
func (s *MetricsService) calculateCPUUsageAverage() (float64, error) {
	var total float64
	var count int

	s.cpuRing.Do(func(x interface{}) {
		if x != nil {
			snapshot := x.(CPUSnapshot)
			total += snapshot.cpuPercent
			count++
		}
	})

	if count == 0 {
		return -1.0, fmt.Errorf("no CPU usage data available") // No data available
	}

	return total / float64(count), nil
}

// updateOtherMetrics updates RAM, disk, and container metrics
func (s *MetricsService) updateOtherMetrics(ctx context.Context) {
	s.otherMutex.Lock()
	defer s.otherMutex.Unlock()

	// Clear previous error at start
	s.lastError = nil

	// Update CPU load averages
	loadAvg, err := load.Avg()
	if err != nil {
		log.Errorf("Error getting CPU load averages: %v", err)
		s.lastError = err
		return
	}
	cpuCounts, err := cpu.Counts(true)
	if err != nil {
		log.Errorf("Error getting CPU counts: %v", err)
		s.lastError = err
		return
	}
	s.cpuLoadAvg = loadAvg.Load15 / float64(cpuCounts)

	// Update RAM usage
	memory, err := mem.VirtualMemory()
	if err != nil {
		log.Errorf("Error getting memory metrics: %v", err)
		s.lastError = err
		return
	}
	s.ramUsage = (float64(memory.Total-memory.Available) / float64(memory.Total)) * 100

	// Update disk usage
	diskUsage, err := disk.Usage("/var/lib/docker")
	if err != nil {
		log.Errorf("Error getting disk metrics: %v", err)
		s.lastError = err
		return
	}
	s.diskUsage = diskUsage.UsedPercent

	// Update snapshot count
	info, err := s.docker.ApiClient().Info(ctx)
	if err != nil {
		log.Errorf("Error getting snapshot count: %v", err)
		s.lastError = err
		return
	}
	s.snapshotCount = info.Images

	// Update allocated resources
	err = s.updateAllocatedResources(ctx)
	if err != nil {
		log.Errorf("Error updating allocated resources: %v", err)
		s.lastError = err
		return
	}

	s.lastUpdated = time.Now()
}

// updateAllocatedResources calculates total allocated container resources
func (s *MetricsService) updateAllocatedResources(ctx context.Context) error {
	containers, err := s.docker.ApiClient().ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return err
	}

	var totalCPU, totalMemory, totalDisk int64 = 0, 0, 0

	for _, ctr := range containers {
		cpu, memory, disk, err := s.getContainerResources(ctx, ctr.ID)
		if err != nil {
			log.Errorf("Error getting resources for container %s: %v", ctr.ID, err)
			continue
		}

		// CPU and memory: only count running containers
		if ctr.State == "running" {
			totalCPU += cpu
			totalMemory += memory
		}
		// Disk: count all containers
		totalDisk += disk
	}

	// Convert to API units
	s.allocatedCPU = totalCPU / 100000                     // Convert to vCPUs
	s.allocatedMemory = totalMemory / (1024 * 1024 * 1024) // Convert to GB
	s.allocatedDisk = totalDisk

	return nil
}

// getContainerResources extracts resource allocation from container config
func (s *MetricsService) getContainerResources(ctx context.Context, containerID string) (int64, int64, int64, error) {
	var cpu, memory, disk int64 = 0, 0, 0

	containerJSON, err := s.docker.ContainerInspect(ctx, containerID)
	if err != nil {
		return cpu, memory, disk, err
	}

	if containerJSON.HostConfig == nil {
		return cpu, memory, disk, fmt.Errorf("container %s has no host config set and runner is unable to get allocated CPU, memory and disk quotas", containerID)
	}

	resources := containerJSON.HostConfig.Resources

	if resources.CPUQuota > 0 {
		cpu = resources.CPUQuota
	} else {
		log.Warnf("Container %s has no CPU quota", containerID)
	}

	if resources.Memory > 0 {
		memory = resources.Memory
	} else {
		log.Warnf("Container %s has no memory quota", containerID)
	}

	// Parse disk allocation from storage options
	if containerJSON.HostConfig.StorageOpt != nil {
		if sizeStr, exists := containerJSON.HostConfig.StorageOpt["size"]; exists {
			diskGB, err := s.parseStorageSize(sizeStr)
			if err == nil {
				disk = diskGB
			} else {
				log.Warnf("Container %s has no disk quota", containerID)
			}
		}
	} else {
		log.Warnf("Container %s has no storage options", containerID)
	}

	return cpu, memory, disk, nil
}

// parseStorageSize parses storage size string like "10G" to GB
func (s *MetricsService) parseStorageSize(sizeStr string) (int64, error) {
	if sizeStr == "" {
		return 0, fmt.Errorf("empty size string")
	}

	sizeStr = strings.TrimSpace(sizeStr)
	if strings.HasSuffix(sizeStr, "G") {
		numStr := strings.TrimSuffix(sizeStr, "G")
		return strconv.ParseInt(numStr, 10, 64)
	}

	return 0, fmt.Errorf("not in expected format (e.g., '10G')")
}
