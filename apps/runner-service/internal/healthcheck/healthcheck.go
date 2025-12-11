/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

package healthcheck

import (
	"context"
	"log/slog"
	"time"

	apiclient "github.com/daytonaio/apiclient"
	"github.com/daytonaio/runner-service/internal/config"
	"github.com/daytonaio/runner-service/internal/metrics"
)

// Service handles healthcheck reporting to the API
type Service struct {
	log       *slog.Logger
	cfg       *config.Config
	client    *apiclient.APIClient
	collector *metrics.Collector
}

// NewService creates a new healthcheck service
func NewService(cfg *config.Config, client *apiclient.APIClient, collector *metrics.Collector, logger *slog.Logger) *Service {
	return &Service{
		log:       logger.With(slog.String("component", "healthcheck")),
		cfg:       cfg,
		client:    client,
		collector: collector,
	}
}

// Start begins the healthcheck loop
func (s *Service) Start(ctx context.Context) {
	ticker := time.NewTicker(s.cfg.HealthcheckInterval)
	defer ticker.Stop()

	// Send initial healthcheck immediately
	if err := s.sendHealthcheck(ctx); err != nil {
		s.log.Warn("Failed to send initial healthcheck", slog.Any("error", err))
	}

	for {
		select {
		case <-ctx.Done():
			s.log.Info("Healthcheck loop stopped")
			return
		case <-ticker.C:
			if err := s.sendHealthcheck(ctx); err != nil {
				s.log.Warn("Failed to send healthcheck", slog.Any("error", err))
				// Continue trying - don't crash
			}
		}
	}
}

// sendHealthcheck sends a healthcheck to the API
func (s *Service) sendHealthcheck(ctx context.Context) error {
	// Create context with timeout
	reqCtx, cancel := context.WithTimeout(ctx, s.cfg.HealthcheckTimeout)
	defer cancel()

	// Collect metrics
	var metricsPtr *apiclient.RunnerHealthMetrics
	m, err := s.collector.Collect(reqCtx)
	if err != nil {
		s.log.Warn("Failed to collect metrics", slog.Any("error", err))
	} else {
		metricsPtr = &apiclient.RunnerHealthMetrics{
			CurrentCpuUsagePercentage:    m.CPUUsagePercentage,
			CurrentMemoryUsagePercentage: m.MemoryUsagePercentage,
			CurrentDiskUsagePercentage:   m.DiskUsagePercentage,
			CurrentAllocatedCpu:          m.AllocatedCPU,
			CurrentAllocatedMemoryGiB:    m.AllocatedMemoryGiB,
			CurrentAllocatedDiskGiB:      m.AllocatedDiskGiB,
			CurrentSnapshotCount:         m.SnapshotCount,
		}
	}

	// Build healthcheck request
	healthcheck := apiclient.NewRunnerHealthcheck()
	if metricsPtr != nil {
		healthcheck.SetMetrics(*metricsPtr)
	}

	// Send healthcheck using the new RunnerServiceAPI
	req := s.client.RunnerServiceAPI.RunnerHealthcheck(reqCtx).RunnerHealthcheck(*healthcheck)
	_, err = req.Execute()
	if err != nil {
		return err
	}

	s.log.Debug("Healthcheck sent successfully")
	return nil
}
