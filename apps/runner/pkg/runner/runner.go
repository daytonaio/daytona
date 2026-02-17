// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package runner

import (
	"context"
	"log"

	"github.com/daytonaio/runner/internal/metrics"
	"github.com/daytonaio/runner/pkg/cache"
	"github.com/daytonaio/runner/pkg/docker"
	"github.com/daytonaio/runner/pkg/netrules"
	"github.com/daytonaio/runner/pkg/services"
	"github.com/daytonaio/runner/pkg/sshgateway"
)

type RunnerInstanceConfig struct {
	Ctx                context.Context
	StatesCache        *cache.StatesCache
	SnapshotErrorCache *cache.SnapshotErrorCache
	Docker             *docker.DockerClient
	MetricsCollector   *metrics.Collector
	SandboxService     *services.SandboxService
	NetRulesManager    *netrules.NetRulesManager
	SSHGatewayService  *sshgateway.Service
}

type Runner struct {
	Ctx                context.Context
	StatesCache        *cache.StatesCache
	SnapshotErrorCache *cache.SnapshotErrorCache
	Docker             *docker.DockerClient
	MetricsCollector   *metrics.Collector
	SandboxService     *services.SandboxService
	NetRulesManager    *netrules.NetRulesManager
	SSHGatewayService  *sshgateway.Service
}

var runner *Runner

func GetInstance(config *RunnerInstanceConfig) *Runner {
	if config != nil && runner != nil {
		log.Fatal("Runner already initialized")
	}

	if runner == nil {
		if config == nil {
			log.Fatal("Runner not initialized")
		}

		runner = &Runner{
			Ctx:                config.Ctx,
			StatesCache:        config.StatesCache,
			SnapshotErrorCache: config.SnapshotErrorCache,
			Docker:             config.Docker,
			SandboxService:     config.SandboxService,
			MetricsCollector:   config.MetricsCollector,
			NetRulesManager:    config.NetRulesManager,
			SSHGatewayService:  config.SSHGatewayService,
		}
	}

	return runner
}
