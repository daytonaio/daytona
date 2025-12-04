// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package runner

import (
	"log"

	"github.com/daytonaio/mock-runner/pkg/cache"
	"github.com/daytonaio/mock-runner/pkg/mock"
	"github.com/daytonaio/mock-runner/pkg/services"
	"github.com/daytonaio/mock-runner/pkg/sshgateway"
	"github.com/daytonaio/mock-runner/pkg/toolbox"
)

type RunnerInstanceConfig struct {
	StatesCache       *cache.StatesCache
	Mock              *mock.MockClient
	SandboxService    *services.SandboxService
	MetricsService    *services.MetricsService
	ToolboxContainer  *toolbox.ToolboxContainer
	SSHGatewayService *sshgateway.Service
}

type Runner struct {
	StatesCache       *cache.StatesCache
	Mock              *mock.MockClient
	SandboxService    *services.SandboxService
	MetricsService    *services.MetricsService
	ToolboxContainer  *toolbox.ToolboxContainer
	SSHGatewayService *sshgateway.Service
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
			Mock:              config.Mock,
			SandboxService:    config.SandboxService,
			MetricsService:    config.MetricsService,
			ToolboxContainer:  config.ToolboxContainer,
			SSHGatewayService: config.SSHGatewayService,
		}
	}

	return runner
}



