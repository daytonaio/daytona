/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

package healthcheck

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	apiclient "github.com/daytonaio/apiclient"
	"github.com/daytonaio/runner/internal"
	"github.com/daytonaio/runner/internal/metrics"
	runnerapiclient "github.com/daytonaio/runner/pkg/apiclient"
)

type HealthcheckServiceConfig struct {
	Interval   time.Duration
	Timeout    time.Duration
	Collector  *metrics.Collector
	Logger     *slog.Logger
	Domain     string
	ApiPort    int
	ProxyPort  int
	TlsEnabled bool
}

// Service handles healthcheck reporting to the API
type Service struct {
	log        *slog.Logger
	interval   time.Duration
	timeout    time.Duration
	collector  *metrics.Collector
	client     *apiclient.APIClient
	domain     string
	apiPort    int
	proxyPort  int
	tlsEnabled bool
}

// NewService creates a new healthcheck service
func NewService(cfg *HealthcheckServiceConfig) (*Service, error) {
	apiClient, err := runnerapiclient.GetApiClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create API client: %w", err)
	}

	return &Service{
		log:        cfg.Logger.With(slog.String("component", "healthcheck")),
		client:     apiClient,
		interval:   cfg.Interval,
		timeout:    cfg.Timeout,
		collector:  cfg.Collector,
		domain:     cfg.Domain,
		apiPort:    cfg.ApiPort,
		proxyPort:  cfg.ProxyPort,
		tlsEnabled: cfg.TlsEnabled,
	}, nil
}

// Start begins the healthcheck loop
func (s *Service) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
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
	reqCtx, cancel := context.WithTimeout(ctx, s.timeout)
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
			Cpu:                          m.TotalCPU,
			MemoryGiB:                    m.TotalRAMGiB,
			DiskGiB:                      m.TotalDiskGiB,
		}
	}

	// Build healthcheck request
	healthcheck := apiclient.NewRunnerHealthcheck(internal.Version)
	if metricsPtr != nil {
		healthcheck.SetMetrics(*metricsPtr)
	}

	healthcheck.SetDomain(s.domain)
	proxyUrl := fmt.Sprintf("http://%s:%d", s.domain, s.proxyPort)
	apiUrl := fmt.Sprintf("http://%s:%d", s.domain, s.apiPort)

	if s.tlsEnabled {
		apiUrl = fmt.Sprintf("https://%s:%d", s.domain, s.apiPort)
		proxyUrl = fmt.Sprintf("https://%s:%d", s.domain, s.proxyPort)
	}

	healthcheck.SetProxyUrl(proxyUrl)
	healthcheck.SetApiUrl(apiUrl)

	req := s.client.RunnersAPI.RunnerHealthcheck(reqCtx).RunnerHealthcheck(*healthcheck)
	_, err = req.Execute()
	if err != nil {
		return err
	}

	s.log.Debug("Healthcheck sent successfully")
	return nil
}
