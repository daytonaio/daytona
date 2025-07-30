// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package services

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/daytonaio/runner/pkg/cache"
	"github.com/daytonaio/runner/pkg/docker"
	"github.com/daytonaio/runner/pkg/models"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
)

type MetricsService struct {
	docker *docker.DockerClient
	cache  cache.IRunnerCache
}

type cpuStats struct {
	user    uint64
	nice    uint64
	system  uint64
	idle    uint64
	iowait  uint64
	irq     uint64
	softirq uint64
	steal   uint64
}

func NewMetricsService(docker *docker.DockerClient, cache cache.IRunnerCache) *MetricsService {
	return &MetricsService{
		docker: docker,
		cache:  cache,
	}
}

func (m *MetricsService) GetSystemMetrics(ctx context.Context) (float64, float64, float64, error) {
	cpuUsage, err := m.getCPUUsage()
	if err != nil {
		cpuUsage = -1.0
	}

	ramUsage, err := m.getRAMUsage()
	if err != nil {
		ramUsage = -1.0
	}

	diskUsage, err := m.getDiskUsage()
	if err != nil {
		diskUsage = -1.0
	}

	return cpuUsage, ramUsage, diskUsage, nil
}

func (m *MetricsService) GetAllocatedResources(ctx context.Context) (int64, int64, int64, error) {
	containers, err := m.docker.ApiClient().ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return -1, -1, -1, err
	}

	var totalAllocatedCpuMicroseconds int64 = 0 // CPU quota in microseconds per period
	var totalAllocatedMemoryBytes int64 = 0     // Memory in bytes
	var totalAllocatedDiskBytes int64 = 0       // Disk in bytes

	for _, containerItem := range containers {
		cpu, memory, disk, err := m.getContainerAllocatedResources(ctx, containerItem.ID)
		if err == nil {
			// For CPU and memory: only count running containers
			if containerItem.State == "running" {
				totalAllocatedCpuMicroseconds += cpu
				totalAllocatedMemoryBytes += memory
			}
			// For disk: count all containers (running and stopped)
			totalAllocatedDiskBytes += disk
		}
	}

	// Convert to original API units
	totalAllocatedCpuVCPUs := totalAllocatedCpuMicroseconds / 100000           // Convert back to vCPUs
	totalAllocatedMemoryGB := totalAllocatedMemoryBytes / (1024 * 1024 * 1024) // Convert back to GB
	totalAllocatedDiskGB := totalAllocatedDiskBytes / (1024 * 1024 * 1024)     // Convert back to GB

	return totalAllocatedCpuVCPUs, totalAllocatedMemoryGB, totalAllocatedDiskGB, nil
}

func (m *MetricsService) GetSnapshotCount(ctx context.Context) (int, error) {
	images, err := m.docker.ApiClient().ImageList(ctx, image.ListOptions{})
	if err != nil {
		return -1, nil // Return -1 on error
	}

	return len(images), nil
}

func (m *MetricsService) getCPUUsage() (float64, error) {
	// Read initial CPU stats
	stats1, err := m.readCPUStats()
	if err != nil {
		return -1.0, err
	}

	// Wait a short period
	time.Sleep(100 * time.Millisecond)

	// Read CPU stats again
	stats2, err := m.readCPUStats()
	if err != nil {
		return -1.0, err
	}

	// Calculate the difference
	total1 := stats1.user + stats1.nice + stats1.system + stats1.idle + stats1.iowait + stats1.irq + stats1.softirq + stats1.steal
	total2 := stats2.user + stats2.nice + stats2.system + stats2.idle + stats2.iowait + stats2.irq + stats2.softirq + stats2.steal

	idle1 := stats1.idle + stats1.iowait
	idle2 := stats2.idle + stats2.iowait

	totalDiff := float64(total2 - total1)
	idleDiff := float64(idle2 - idle1)

	if totalDiff == 0 {
		return -1.0, fmt.Errorf("no CPU time difference detected")
	}

	cpuUsage := (1.0 - idleDiff/totalDiff) * 100.0

	// Ensure valid percentage range
	if cpuUsage < 0 || cpuUsage > 100 {
		return -1.0, fmt.Errorf("invalid CPU usage calculated: %f", cpuUsage)
	}

	return cpuUsage, nil
}

func (m *MetricsService) readCPUStats() (*cpuStats, error) {
	file, err := os.Open("/proc/stat")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return nil, fmt.Errorf("failed to read CPU line from /proc/stat")
	}

	line := scanner.Text()
	fields := strings.Fields(line)
	if len(fields) < 8 || fields[0] != "cpu" {
		return nil, fmt.Errorf("invalid CPU line format")
	}

	stats := &cpuStats{}
	stats.user, _ = strconv.ParseUint(fields[1], 10, 64)
	stats.nice, _ = strconv.ParseUint(fields[2], 10, 64)
	stats.system, _ = strconv.ParseUint(fields[3], 10, 64)
	stats.idle, _ = strconv.ParseUint(fields[4], 10, 64)
	stats.iowait, _ = strconv.ParseUint(fields[5], 10, 64)
	stats.irq, _ = strconv.ParseUint(fields[6], 10, 64)
	stats.softirq, _ = strconv.ParseUint(fields[7], 10, 64)
	if len(fields) > 8 {
		stats.steal, _ = strconv.ParseUint(fields[8], 10, 64)
	}

	return stats, nil
}

func (m *MetricsService) getRAMUsage() (float64, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return -1.0, err
	}
	defer file.Close()

	var memTotal, memAvailable uint64
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			switch fields[0] {
			case "MemTotal:":
				memTotal, _ = strconv.ParseUint(fields[1], 10, 64)
			case "MemAvailable:":
				memAvailable, _ = strconv.ParseUint(fields[1], 10, 64)
			}
		}
	}

	if memTotal == 0 {
		return -1.0, fmt.Errorf("could not read memory total")
	}

	if memAvailable > memTotal {
		return -1.0, fmt.Errorf("invalid memory values: available > total")
	}

	memUsed := memTotal - memAvailable
	ramUsage := (float64(memUsed) / float64(memTotal)) * 100.0

	// Ensure valid percentage range
	if ramUsage < 0 || ramUsage > 100 {
		return -1.0, fmt.Errorf("invalid RAM usage calculated: %f", ramUsage)
	}

	return ramUsage, nil
}

func (m *MetricsService) getDiskUsage() (float64, error) {
	// Get disk usage for the root filesystem
	var stat syscall.Statfs_t
	err := syscall.Statfs("/", &stat)
	if err != nil {
		return m.getDiskUsageCommand()
	}

	// Available blocks * block size = available space
	totalBytes := stat.Blocks * uint64(stat.Bsize)
	availableBytes := stat.Bavail * uint64(stat.Bsize)

	if totalBytes == 0 {
		return -1.0, fmt.Errorf("total disk space is zero")
	}

	if availableBytes > totalBytes {
		return -1.0, fmt.Errorf("invalid disk values: available > total")
	}

	usedBytes := totalBytes - availableBytes
	diskUsage := (float64(usedBytes) / float64(totalBytes)) * 100.0

	// Ensure valid percentage range
	if diskUsage < 0 || diskUsage > 100 {
		return -1.0, fmt.Errorf("invalid disk usage calculated: %f", diskUsage)
	}

	return diskUsage, nil
}

func (m *MetricsService) getDiskUsageCommand() (float64, error) {
	// Fallback to df command if syscall fails
	cmd := exec.Command("df", "/")
	output, err := cmd.Output()
	if err != nil {
		return -1.0, err
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return -1.0, fmt.Errorf("unexpected df output")
	}

	fields := strings.Fields(lines[1])
	if len(fields) < 5 {
		return -1.0, fmt.Errorf("unexpected df output format")
	}

	// fields[4] contains the usage percentage like "45%"
	usageStr := strings.TrimSuffix(fields[4], "%")
	usage, err := strconv.ParseFloat(usageStr, 64)
	if err != nil {
		return -1.0, err
	}

	// Ensure valid percentage range
	if usage < 0 || usage > 100 {
		return -1.0, fmt.Errorf("invalid disk usage from df: %f", usage)
	}

	return usage, nil
}

func (m *MetricsService) getContainerAllocatedResources(ctx context.Context, containerID string) (int64, int64, int64, error) {
	// Inspect the container to get its resource configuration
	containerJSON, err := m.docker.ApiClient().ContainerInspect(ctx, containerID)
	if err != nil {
		return 0, 0, 0, err
	}

	var allocatedCpu int64 = 0
	var allocatedMemory int64 = 0
	var allocatedDisk int64 = 0

	if containerJSON.HostConfig != nil {
		resources := containerJSON.HostConfig.Resources

		// CPU allocation (convert from quota to total microseconds if quota is set)
		if resources.CPUQuota > 0 {
			allocatedCpu = resources.CPUQuota
		}

		// Memory allocation
		if resources.Memory > 0 {
			allocatedMemory = resources.Memory
		}

		// Disk allocation from StorageOpt (assuming xfs filesystem)
		if containerJSON.HostConfig.StorageOpt != nil {
			if sizeStr, exists := containerJSON.HostConfig.StorageOpt["size"]; exists {
				// Parse size string like "10G" and convert to GB
				if diskGB, err := m.parseStorageQuotaGB(sizeStr); err == nil {
					allocatedDisk = diskGB
				}
			}
		}
	}

	return allocatedCpu, allocatedMemory, allocatedDisk, nil
}

func (m *MetricsService) parseStorageQuotaGB(sizeStr string) (int64, error) {
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

// GetCachedSystemMetrics returns cached metrics if available, otherwise returns defaults
func (m *MetricsService) GetCachedSystemMetrics(ctx context.Context) (float64, float64, float64, int64, int64, int64, int) {
	metrics := m.cache.GetSystemMetrics(ctx)
	if metrics == nil {
		// Return default values if no metrics are cached
		return -1.0, -1.0, -1.0, -1, -1, -1, -1
	}

	return metrics.CPUUsage, metrics.RAMUsage, metrics.DiskUsage,
		metrics.AllocatedCPU, metrics.AllocatedMemory, metrics.AllocatedDisk,
		metrics.SnapshotCount
}

// CollectAndCacheMetrics collects all metrics and stores them in cache
func (m *MetricsService) CollectAndCacheMetrics(ctx context.Context) error {
	// Get current cached metrics to preserve valid values
	cachedMetrics := m.cache.GetSystemMetrics(ctx)

	// Initialize with default values or preserve existing ones
	var finalCpuUsage, finalRamUsage, finalDiskUsage float64 = -1.0, -1.0, -1.0
	var finalAllocatedCpu, finalAllocatedMemory, finalAllocatedDisk int64 = -1, -1, -1
	var finalSnapshotCount int = -1

	// If we have cached metrics, use them as starting point
	if cachedMetrics != nil {
		finalCpuUsage = cachedMetrics.CPUUsage
		finalRamUsage = cachedMetrics.RAMUsage
		finalDiskUsage = cachedMetrics.DiskUsage
		finalAllocatedCpu = cachedMetrics.AllocatedCPU
		finalAllocatedMemory = cachedMetrics.AllocatedMemory
		finalAllocatedDisk = cachedMetrics.AllocatedDisk
		finalSnapshotCount = cachedMetrics.SnapshotCount
	}

	// Collect system metrics
	cpuUsage, ramUsage, diskUsage, err := m.GetSystemMetrics(ctx)
	if err == nil {
		// Only update if values are valid (not 0 or -1)
		if cpuUsage > 0 && cpuUsage != -1.0 {
			finalCpuUsage = cpuUsage
		}
		if ramUsage > 0 && ramUsage != -1.0 {
			finalRamUsage = ramUsage
		}
		if diskUsage > 0 && diskUsage != -1.0 {
			finalDiskUsage = diskUsage
		}
	}

	// Get allocated resources
	allocatedCpu, allocatedMemory, allocatedDisk, err := m.GetAllocatedResources(ctx)
	if err == nil {
		// Only update if values are valid (not 0 or -1)
		// Note: 0 can be valid for allocated resources (no containers running)
		// so we only check for -1
		if allocatedCpu != -1 {
			finalAllocatedCpu = allocatedCpu
		}
		if allocatedMemory != -1 {
			finalAllocatedMemory = allocatedMemory
		}
		if allocatedDisk != -1 {
			finalAllocatedDisk = allocatedDisk
		}
	}

	// Get snapshot count
	snapshotCount, err := m.GetSnapshotCount(ctx)
	if err == nil && snapshotCount != -1 {
		// 0 is a valid snapshot count, so only check for -1
		finalSnapshotCount = snapshotCount
	}

	// Store in cache with final values
	metrics := models.SystemMetrics{
		CPUUsage:        finalCpuUsage,
		RAMUsage:        finalRamUsage,
		DiskUsage:       finalDiskUsage,
		AllocatedCPU:    finalAllocatedCpu,
		AllocatedMemory: finalAllocatedMemory,
		AllocatedDisk:   finalAllocatedDisk,
		SnapshotCount:   finalSnapshotCount,
		LastUpdated:     time.Now(),
	}

	m.cache.SetSystemMetrics(ctx, metrics)
	return nil
}

// StartMetricsCollection starts a background goroutine that collects metrics every 20 seconds
func (m *MetricsService) StartMetricsCollection(ctx context.Context) {
	go func() {
		// Collect metrics immediately on startup
		_ = m.CollectAndCacheMetrics(ctx)

		// Set up ticker for every 20 seconds
		ticker := time.NewTicker(20 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				_ = m.CollectAndCacheMetrics(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()
}
