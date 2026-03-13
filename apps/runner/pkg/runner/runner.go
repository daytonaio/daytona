// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package runner

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/daytonaio/runner/internal/metrics"
	"github.com/daytonaio/runner/pkg/cache"
	"github.com/daytonaio/runner/pkg/docker"
	"github.com/daytonaio/runner/pkg/models"
	"github.com/daytonaio/runner/pkg/netrules"
	"github.com/daytonaio/runner/pkg/services"
	"github.com/daytonaio/runner/pkg/sshgateway"
)

type RunnerInstanceConfig struct {
	Logger             *slog.Logger
	StatesCache        *cache.StatesCache
	SnapshotErrorCache *cache.SnapshotErrorCache
	Docker             *docker.DockerClient
	MetricsCollector   *metrics.Collector
	SandboxService     *services.SandboxService
	NetRulesManager    *netrules.NetRulesManager
	SSHGatewayService  *sshgateway.Service
}

type Runner struct {
	Logger             *slog.Logger
	StatesCache        *cache.StatesCache
	SnapshotErrorCache *cache.SnapshotErrorCache
	Docker             *docker.DockerClient
	MetricsCollector   *metrics.Collector
	SandboxService     *services.SandboxService
	NetRulesManager    *netrules.NetRulesManager
	SSHGatewayService  *sshgateway.Service
}

var runner *Runner

func GetInstance(config *RunnerInstanceConfig) (*Runner, error) {
	if config != nil && runner != nil {
		return nil, errors.New("runner instance already initialized")
	}

	if runner == nil {
		if config == nil {
			return nil, errors.New("runner instance not initialized and no config provided")
		}

		logger := slog.Default()
		if config.Logger != nil {
			logger = config.Logger
		}

		runner = &Runner{
			Logger:             logger.With(slog.String("component", "runner")),
			StatesCache:        config.StatesCache,
			SnapshotErrorCache: config.SnapshotErrorCache,
			Docker:             config.Docker,
			SandboxService:     config.SandboxService,
			MetricsCollector:   config.MetricsCollector,
			NetRulesManager:    config.NetRulesManager,
			SSHGatewayService:  config.SSHGatewayService,
		}
	}

	return runner, nil
}

func (r *Runner) InspectRunnerServices(ctx context.Context) []models.RunnerServiceInfo {
	runnerServicesInfo := make([]models.RunnerServiceInfo, 0)

	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	dockerHealth := models.RunnerServiceInfo{
		ServiceName: "docker",
		Healthy:     true,
	}

	err := r.Docker.Ping(pingCtx)
	if err != nil {
		r.Logger.WarnContext(ctx, "Failed to ping Docker daemon", "error", err)
		dockerHealth.Healthy = false
		dockerHealth.Err = err
	}

	runnerServicesInfo = append(runnerServicesInfo, dockerHealth)

	return runnerServicesInfo
}
