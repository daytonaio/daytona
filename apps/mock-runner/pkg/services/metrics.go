// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package services

import (
	"context"
	"sync"
	"time"

	"github.com/daytonaio/mock-runner/pkg/mock"
	"github.com/daytonaio/mock-runner/pkg/models"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"

	log "github.com/sirupsen/logrus"
)

type MetricsServiceConfig struct {
	Mock     *mock.MockClient
	Interval time.Duration
}

type MetricsService struct {
	mockClient *mock.MockClient
	interval   time.Duration
	metrics    *models.SystemMetrics
	mu         sync.RWMutex
}

func NewMetricsService(config MetricsServiceConfig) *MetricsService {
	return &MetricsService{
		mockClient: config.Mock,
		interval:   config.Interval,
		metrics:    &models.SystemMetrics{},
	}
}

func (s *MetricsService) StartMetricsCollection(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		// Collect initial metrics
		s.collectMetrics()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.collectMetrics()
			}
		}
	}()
}

func (s *MetricsService) collectMetrics() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// CPU usage
	cpuPercent, err := cpu.Percent(0, false)
	if err != nil {
		log.Warnf("Failed to get CPU usage: %v", err)
	} else if len(cpuPercent) > 0 {
		s.metrics.CPUUsage = cpuPercent[0]
	}

	// Memory usage
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		log.Warnf("Failed to get memory usage: %v", err)
	} else {
		s.metrics.RAMUsage = memInfo.UsedPercent
		s.metrics.AllocatedMemory = float64(memInfo.Used) / (1024 * 1024 * 1024) // Convert to GB
	}

	// Disk usage
	diskInfo, err := disk.Usage("/")
	if err != nil {
		log.Warnf("Failed to get disk usage: %v", err)
	} else {
		s.metrics.DiskUsage = diskInfo.UsedPercent
		s.metrics.AllocatedDisk = float64(diskInfo.Used) / (1024 * 1024 * 1024) // Convert to GB
	}

	// CPU count as allocated CPU (mock)
	cpuCount, err := cpu.Counts(true)
	if err != nil {
		log.Warnf("Failed to get CPU count: %v", err)
	} else {
		s.metrics.AllocatedCPU = float64(cpuCount)
	}

	// Mock snapshot count (count tracked images)
	s.metrics.SnapshotCount = 0 // Could be enhanced to count images in mock client
}

func (s *MetricsService) GetSystemMetrics(ctx context.Context) *models.SystemMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	/*
		return &models.SystemMetrics{
			CPUUsage:        s.metrics.CPUUsage,
			RAMUsage:        s.metrics.RAMUsage,
			DiskUsage:       s.metrics.DiskUsage,
			AllocatedCPU:    s.metrics.AllocatedCPU,
			AllocatedMemory: s.metrics.AllocatedMemory,
			AllocatedDisk:   s.metrics.AllocatedDisk,
			SnapshotCount:   s.metrics.SnapshotCount,
		}
	*/

	// TODO: Remove this once mock metrics are implemented properly
	return &models.SystemMetrics{
		CPUUsage:        10,
		RAMUsage:        10,
		DiskUsage:       10,
		AllocatedCPU:    10,
		AllocatedMemory: 10,
		AllocatedDisk:   10,
		SnapshotCount:   10,
	}
}
