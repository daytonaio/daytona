// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package runner

import (
	"log"

	"github.com/daytonaio/runner/pkg/cache"
	"github.com/daytonaio/runner/pkg/docker"
	"github.com/daytonaio/runner/pkg/netrules"
	"github.com/daytonaio/runner/pkg/sdisk"
	"github.com/daytonaio/runner/pkg/services"
	"github.com/daytonaio/runner/pkg/sshgateway"
)

type RunnerInstanceConfig struct {
	StatesCache       *cache.StatesCache
	Docker            *docker.DockerClient
	SandboxService    *services.SandboxService
	MetricsService    *services.MetricsService
	NetRulesManager   *netrules.NetRulesManager
	SSHGatewayService *sshgateway.Service
	SDisk             *sdisk.DiskManager
}

type Runner struct {
	StatesCache       *cache.StatesCache
	Docker            *docker.DockerClient
	SandboxService    *services.SandboxService
	MetricsService    *services.MetricsService
	NetRulesManager   *netrules.NetRulesManager
	SSHGatewayService *sshgateway.Service
	SDisk             *sdisk.DiskManager
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
			StatesCache:       config.StatesCache,
			Docker:            config.Docker,
			SandboxService:    config.SandboxService,
			MetricsService:    config.MetricsService,
			NetRulesManager:   config.NetRulesManager,
			SSHGatewayService: config.SSHGatewayService,
			SDisk:             config.SDisk,
		}
	}

	return runner
}
