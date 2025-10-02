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

	common_cache "github.com/daytonaio/common-go/pkg/cache"

	"github.com/docker/docker/api/types/container"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
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
	Docker     *docker.DockerClient
	WindowSize int // CPU metrics window size in seconds
}

// MetricsService provides system metrics with proper encapsulation
type MetricsService struct {
	cache    common_cache.ICache[models.SystemMetrics]
	docker   *docker.DockerClient

	// CPU metrics - ring buffer for sliding window
	cpuRing  *ring.Ring
	cpuMutex sync.RWMutex

	// Other metrics - cached values
	ramUsage        float64
	diskUsage       float64
	allocatedCPU    int64
	allocatedMemory int64
	allocatedDisk   int64
	snapshotCount   int
	lastUpdated     time.Time
	otherMutex      sync.RWMutex
}

// NewMetricsService creates a new metrics service instance
func NewMetricsService(config MetricsServiceConfig) *MetricsService {
	metricsCache := common_cache.NewMapCache[models.SystemMetrics]()

	windowSize := config.WindowSize
	if windowSize <= 0 {
		windowSize = 60 // Default to 60 seconds
	}

	return &MetricsService{
		cache:   metricsCache,
		docker:  config.Docker,
		cpuRing: ring.New(windowSize),
	}
}

// Start begins metrics collection in background goroutines
func (s *MetricsService) Start(ctx context.Context) {
	go s.collectCPUMetrics(ctx)
	go s.collectOtherMetrics(ctx)
}

// GetMetrics returns all current system metrics
func (s *MetricsService) GetMetrics() *models.SystemMetrics {
	// Get CPU metrics (requires read lock)
	s.cpuMutex.RLock()
	cpuUsage := s.calculateCPUAverage()
	s.cpuMutex.RUnlock()

	// Get other metrics (requires read lock)
	s.otherMutex.RLock()
	metrics := &models.SystemMetrics{
		CPUUsage:        cpuUsage,
		RAMUsage:        s.ramUsage,
		DiskUsage:       s.diskUsage,
		AllocatedCPU:    s.allocatedCPU,
		AllocatedMemory: s.allocatedMemory,
		AllocatedDisk:   s.allocatedDisk,
		SnapshotCount:   s.snapshotCount,
		LastUpdated:     s.lastUpdated,
	}
	s.otherMutex.RUnlock()

	return metrics
}

// collectCPUMetrics runs in a background goroutine, continuously monitoring CPU
func (s *MetricsService) collectCPUMetrics(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			s.addCPUSnapshot()
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

// addCPUSnapshot adds a new CPU snapshot to the ring buffer
func (s *MetricsService) addCPUSnapshot() {
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

// calculateCPUAverage calculates the average CPU usage from the ring buffer
func (s *MetricsService) calculateCPUAverage() float64 {
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
		return -1.0 // No data available
	}

	return total / float64(count)
}

// updateOtherMetrics updates RAM, disk, and container metrics
func (s *MetricsService) updateOtherMetrics(ctx context.Context) {
	s.otherMutex.Lock()
	defer s.otherMutex.Unlock()

	// Update RAM usage
	if memory, err := mem.VirtualMemory(); err != nil {
		log.Errorf("Error getting memory metrics: %v", err)
	} else {
		s.ramUsage = (float64(memory.Total-memory.Available) / float64(memory.Total)) * 100
	}

	// Update disk usage
	if diskUsage, err := disk.Usage("/var/lib/docker"); err != nil {
		log.Errorf("Error getting disk metrics: %v", err)
	} else {
		s.diskUsage = diskUsage.UsedPercent
	}

	// Update snapshot count
	if info, err := s.docker.ApiClient().Info(ctx); err != nil {
		log.Errorf("Error getting snapshot count: %v", err)
	} else {
		s.snapshotCount = info.Images
	}

	// Update allocated resources
	s.updateAllocatedResources(ctx)

	s.lastUpdated = time.Now()
}

// updateAllocatedResources calculates total allocated container resources
func (s *MetricsService) updateAllocatedResources(ctx context.Context) {
	containers, err := s.docker.ApiClient().ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		log.Errorf("Error listing containers: %v", err)
		return
	}

	var totalCPU, totalMemory, totalDisk int64

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
}

// getContainerResources extracts resource allocation from container config
func (s *MetricsService) getContainerResources(ctx context.Context, containerID string) (int64, int64, int64, error) {
	containerJSON, err := s.docker.ContainerInspect(ctx, containerID)
	if err != nil {
		return 0, 0, 0, err
	}

	var cpu, memory, disk int64

	if containerJSON.HostConfig != nil {
		resources := containerJSON.HostConfig.Resources

		if resources.CPUQuota > 0 {
			cpu = resources.CPUQuota
		}
		if resources.Memory > 0 {
			memory = resources.Memory
		}
	}

	// Parse disk allocation from storage options
	if containerJSON.HostConfig.StorageOpt != nil {
		if sizeStr, exists := containerJSON.HostConfig.StorageOpt["size"]; exists {
			if diskGB, err := s.parseStorageSize(sizeStr); err == nil {
				disk = diskGB
			}
		}
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
