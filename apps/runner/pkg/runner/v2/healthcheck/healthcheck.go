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

	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
	"github.com/daytonaio/runner/internal"
	"github.com/daytonaio/runner/internal/metrics"
	runnerapiclient "github.com/daytonaio/runner/pkg/apiclient"
	"github.com/daytonaio/runner/pkg/docker"
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
	Docker     *docker.DockerClient
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
	docker     *docker.DockerClient
}

// NewService creates a new healthcheck service
func NewService(cfg *HealthcheckServiceConfig) (*Service, error) {
	apiClient, err := runnerapiclient.GetApiClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create API client: %w", err)
	}

	if cfg.Docker == nil {
		return nil, fmt.Errorf("docker client is required for healthcheck service")
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
		docker:     cfg.Docker,
	}, nil
}

// Start begins the healthcheck loop
func (s *Service) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// Send initial healthcheck immediately
	if err := s.sendHealthcheck(ctx); err != nil {
		s.log.WarnContext(ctx, "Failed to send initial healthcheck", "error", err)
	}

	for {
		select {
		case <-ctx.Done():
			s.log.InfoContext(ctx, "Healthcheck loop stopped")
			return
		case <-ticker.C:
			if err := s.sendHealthcheck(ctx); err != nil {
				s.log.WarnContext(ctx, "Failed to send healthcheck", "error", err)
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

	// Build healthcheck request
	healthcheck := apiclient.NewRunnerHealthcheck(internal.Version)
	healthcheck.SetDomain(s.domain)

	proxyUrl := fmt.Sprintf("http://%s:%d", s.domain, s.proxyPort)
	apiUrl := fmt.Sprintf("http://%s:%d", s.domain, s.apiPort)

	if s.tlsEnabled {
		apiUrl = fmt.Sprintf("https://%s:%d", s.domain, s.apiPort)
		proxyUrl = fmt.Sprintf("https://%s:%d", s.domain, s.proxyPort)
	}

	healthcheck.SetProxyUrl(proxyUrl)
	healthcheck.SetApiUrl(apiUrl)

	dockerHealth := apiclient.RunnerServiceHealth{
		ServiceName: "docker",
		Healthy:     true,
	}

	err := s.docker.Ping(reqCtx)
	if err != nil {
		s.log.WarnContext(ctx, "Failed to ping Docker daemon", "error", err)

		errStr := err.Error()
		dockerHealth.Healthy = false
		dockerHealth.Error = &errStr
	}

	healthcheck.SetServiceHealth([]apiclient.RunnerServiceHealth{dockerHealth})

	// Collect metrics
	m, err := s.collector.Collect(reqCtx)
	if err != nil {
		s.log.WarnContext(ctx, "Failed to collect metrics for healthcheck", "error", err)
	} else {
		healthcheck.SetMetrics(apiclient.RunnerHealthMetrics{
			CurrentCpuLoadAverage:        m.CPULoadAverage,
			CurrentCpuUsagePercentage:    m.CPUUsagePercentage,
			CurrentMemoryUsagePercentage: m.MemoryUsagePercentage,
			CurrentDiskUsagePercentage:   m.DiskUsagePercentage,
			CurrentAllocatedCpu:          m.AllocatedCPU,
			CurrentAllocatedMemoryGiB:    m.AllocatedMemoryGiB,
			CurrentAllocatedDiskGiB:      m.AllocatedDiskGiB,
			CurrentSnapshotCount:         m.SnapshotCount,
			CurrentStartedSandboxes:      m.StartedSandboxCount,
			Cpu:                          m.TotalCPU,
			MemoryGiB:                    m.TotalRAMGiB,
			DiskGiB:                      m.TotalDiskGiB,
		})
	}

	req := s.client.RunnersAPI.RunnerHealthcheck(reqCtx).RunnerHealthcheck(*healthcheck)
	_, err = req.Execute()
	if err != nil {
		return err
	}

	s.log.DebugContext(ctx, "Healthcheck sent successfully")
	return nil
}
