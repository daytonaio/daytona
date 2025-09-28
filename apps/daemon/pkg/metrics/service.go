// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package metrics

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// Service handles the caching and periodic updates of sandbox metrics
type Service struct {
	collector    MetricsCollector
	cache        *ResourceMetrics
	updatePeriod time.Duration
	sandboxId    string
	mu           sync.RWMutex
	stopChan     chan struct{}
	stoppedChan  chan struct{}
}

// NewService creates a new metrics service with the given collector and update period
func NewService(collector MetricsCollector, updatePeriod time.Duration, sandboxId string) *Service {
	return &Service{
		collector:    collector,
		updatePeriod: updatePeriod,
		sandboxId:    sandboxId,
		cache:        &ResourceMetrics{},
		stopChan:     make(chan struct{}),
		stoppedChan:  make(chan struct{}),
	}
}

// Start begins the periodic metrics collection
func (s *Service) Start() error {
	// Initial collection
	if err := s.collect(); err != nil {
		return err
	}

	go s.periodicCollection()
	return nil
}

// Stop halts the periodic metrics collection
func (s *Service) Stop() {
	close(s.stopChan)
	<-s.stoppedChan
}

// GetMetrics returns the currently cached metrics
func (s *Service) GetMetrics() ResourceMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return *s.cache
}

func (s *Service) collect() error {
	var lastError error
	metrics := &ResourceMetrics{}

	// Track which metrics were successfully collected
	var cpuCollected, memCollected, diskCollected bool

	// Collect CPU metrics
	cpu, err := s.collector.GetCPUPercentage()
	if err != nil {
		log.Warnf("Failed to collect CPU metrics: %v", err)
		lastError = err
	} else {
		metrics.CPUPercentage = cpu
		cpuCollected = true
	}

	// Collect Memory metrics
	memory, err := s.collector.GetMemoryPercentage()
	if err != nil {
		log.Warnf("Failed to collect Memory metrics: %v", err)
		lastError = err
	} else {
		metrics.MemoryPercentage = memory
		memCollected = true
	}

	// Collect Disk metrics
	disk, err := s.collector.GetDiskPercentage()
	if err != nil {
		log.Warnf("Failed to collect Disk metrics: %v", err)
		lastError = err
	} else {
		metrics.DiskPercentage = disk
		diskCollected = true
	}

	// Update cache if we collected any metrics successfully
	if cpuCollected || memCollected || diskCollected {
		s.mu.Lock()
		// Only update metrics that were successfully collected
		if cpuCollected {
			s.cache.CPUPercentage = metrics.CPUPercentage
		}
		if memCollected {
			s.cache.MemoryPercentage = metrics.MemoryPercentage
		}
		if diskCollected {
			s.cache.DiskPercentage = metrics.DiskPercentage
		}
		s.mu.Unlock()

		log.Infof("%s: CPU: %.1f%%, MEM: %.1f%%, DISK: %.1f%%",
			s.sandboxId,
			metrics.CPUPercentage,
			metrics.MemoryPercentage,
			metrics.DiskPercentage)
	}

	return lastError
}

func (s *Service) periodicCollection() {
	ticker := time.NewTicker(s.updatePeriod)
	defer ticker.Stop()
	defer close(s.stoppedChan)

	for {
		select {
		case <-ticker.C:
			if err := s.collect(); err != nil {
				log.Errorf("Failed to collect metrics: %v", err)
			}
		case <-s.stopChan:
			return
		}
	}
}
