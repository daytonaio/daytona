// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runner

import (
	"github.com/daytonaio/daytona/pkg/jobs"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/runner/providermanager"
	"github.com/daytonaio/daytona/pkg/telemetry"
)

type IRunnerJobFactory interface {
	Create(job models.Job) jobs.IJob
}

type RunnerJobFactory struct {
	config RunnerJobFactoryConfig
}

type RunnerJobFactoryConfig struct {
	TrackTelemetryEvent func(event telemetry.BuildRunnerEvent, clientId string, props map[string]interface{}) error
	ProviderManager     providermanager.IProviderManager
}

func NewRunnerJobFactory(config RunnerJobFactoryConfig) IRunnerJobFactory {
	return &RunnerJobFactory{
		config: config,
	}
}

func (f *RunnerJobFactory) Create(job models.Job) jobs.IJob {
	return &RunnerJob{
		Job:                 job,
		trackTelemetryEvent: f.config.TrackTelemetryEvent,
		providerManager:     f.config.ProviderManager,
	}
}
