// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package runner

import (
	"log"

	"github.com/daytonaio/runner-win/pkg/cache"
	"github.com/daytonaio/runner-win/pkg/libvirt"
	"github.com/daytonaio/runner-win/pkg/netrules"
	"github.com/daytonaio/runner-win/pkg/services"
	"github.com/daytonaio/runner-win/pkg/sshgateway"
)

type RunnerInstanceConfig struct {
	StatesCache       *cache.StatesCache
	LibVirt           *libvirt.LibVirt
	SandboxService    *services.SandboxService
	MetricsService    *services.MetricsService
	NetRulesManager   *netrules.NetRulesManager
	SSHGatewayService *sshgateway.Service
	StatsStore        *libvirt.StatsStore
}

type Runner struct {
	StatesCache       *cache.StatesCache
	LibVirt           *libvirt.LibVirt
	SandboxService    *services.SandboxService
	MetricsService    *services.MetricsService
	NetRulesManager   *netrules.NetRulesManager
	SSHGatewayService *sshgateway.Service
	StatsStore        *libvirt.StatsStore
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
			LibVirt:           config.LibVirt,
			SandboxService:    config.SandboxService,
			MetricsService:    config.MetricsService,
			NetRulesManager:   config.NetRulesManager,
			SSHGatewayService: config.SSHGatewayService,
			StatsStore:        config.StatsStore,
		}
	}

	return runner
}
