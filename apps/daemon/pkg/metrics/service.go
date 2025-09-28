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
	mu           sync.RWMutex
	stopChan     chan struct{}
	stoppedChan  chan struct{}
}

// NewService creates a new metrics service with the given collector and update period
func NewService(collector MetricsCollector, updatePeriod time.Duration) *Service {
	return &Service{
		collector:    collector,
		updatePeriod: updatePeriod,
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
	cpu, err := s.collector.GetCPUPercentage()
	if err != nil {
		return err
	}

	memory, err := s.collector.GetMemoryPercentage()
	if err != nil {
		return err
	}

	disk, err := s.collector.GetDiskPercentage()
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.cache.CPUPercentage = cpu
	s.cache.MemoryPercentage = memory
	s.cache.DiskPercentage = disk
	s.mu.Unlock()

	return nil
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
