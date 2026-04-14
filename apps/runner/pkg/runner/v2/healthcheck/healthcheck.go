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

	"github.com/daytonaio/runner/internal"
	"github.com/daytonaio/runner/internal/metrics"
	"github.com/daytonaio/runner/pkg/docker"
	"github.com/daytonaio/runner/pkg/runner/v2/client"
	specsgen "github.com/daytonaio/runner/pkg/runner/v2/specs/gen"
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

type Service struct {
	log        *slog.Logger
	interval   time.Duration
	timeout    time.Duration
	collector  *metrics.Collector
	client     *client.APIClient
	domain     string
	apiPort    int
	proxyPort  int
	tlsEnabled bool
	docker     *docker.DockerClient
}

func NewService(cfg *HealthcheckServiceConfig) (*Service, error) {
	apiClient, err := client.NewAPIClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create API client: %w", err)
	}

	if cfg.Docker == nil {
		return nil, fmt.Errorf("docker client is required for healthcheck service")
	}

	logger := slog.Default()
	if cfg.Logger != nil {
		logger = cfg.Logger
	}

	return &Service{
		log:        logger.With(slog.String("component", "healthcheck")),
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

func (s *Service) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

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
			}
		}
	}
}

func (s *Service) sendHealthcheck(ctx context.Context) error {
	reqCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	proxyUrl := fmt.Sprintf("http://%s:%d", s.domain, s.proxyPort)
	apiUrl := fmt.Sprintf("http://%s:%d", s.domain, s.apiPort)
	if s.tlsEnabled {
		apiUrl = fmt.Sprintf("https://%s:%d", s.domain, s.apiPort)
		proxyUrl = fmt.Sprintf("https://%s:%d", s.domain, s.proxyPort)
	}

	hc := &specsgen.RunnerHealthcheck{
		AppVersion: internal.Version,
		Domain:     &s.domain,
		ProxyUrl:   &proxyUrl,
		ApiUrl:     &apiUrl,
	}

	dockerHealth := &specsgen.RunnerServiceHealth{
		ServiceName: "docker",
		Healthy:     true,
	}

	if err := s.docker.Ping(reqCtx); err != nil {
		s.log.WarnContext(reqCtx, "Failed to ping Docker daemon", "error", err)
		errStr := err.Error()
		dockerHealth.Healthy = false
		dockerHealth.ErrorReason = &errStr
	}

	hc.ServiceHealth = []*specsgen.RunnerServiceHealth{dockerHealth}

	m, err := s.collector.Collect(reqCtx)
	if err != nil {
		s.log.WarnContext(reqCtx, "Failed to collect metrics for healthcheck", "error", err)
	} else {
		hc.Metrics = &specsgen.RunnerHealthMetrics{
			CurrentCpuLoadAverage:        float64(m.CPULoadAverage),
			CurrentCpuUsagePercentage:    float64(m.CPUUsagePercentage),
			CurrentMemoryUsagePercentage: float64(m.MemoryUsagePercentage),
			CurrentDiskUsagePercentage:   float64(m.DiskUsagePercentage),
			CurrentAllocatedCpu:          float64(m.AllocatedCPU),
			CurrentAllocatedMemoryGiB:    float64(m.AllocatedMemoryGiB),
			CurrentAllocatedDiskGiB:      float64(m.AllocatedDiskGiB),
			CurrentSnapshotCount:         int32(m.SnapshotCount),
			CurrentStartedSandboxes:      int64(m.StartedSandboxCount),
			Cpu:                          float64(m.TotalCPU),
			MemoryGiB:                    float64(m.TotalRAMGiB),
			DiskGiB:                      float64(m.TotalDiskGiB),
		}
	}

	_, err = s.client.Do(reqCtx, "POST", "/runners/healthcheck", hc, nil)
	if err != nil {
		return err
	}

	s.log.DebugContext(reqCtx, "Healthcheck sent successfully")
	return nil
}
