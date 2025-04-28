// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package runner

import (
	"log"

	"github.com/daytonaio/runner/pkg/cache"
	"github.com/daytonaio/runner/pkg/docker"
	"github.com/daytonaio/runner/pkg/services"
)

type RunnerInstanceConfig struct {
	Cache          cache.IRunnerCache
	Docker         *docker.DockerClient
	SandboxService *services.SandboxService
}

type Runner struct {
	Cache          cache.IRunnerCache
	Docker         *docker.DockerClient
	SandboxService *services.SandboxService
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
			Cache:          config.Cache,
			Docker:         config.Docker,
			SandboxService: config.SandboxService,
		}
	}

	return runner
}
