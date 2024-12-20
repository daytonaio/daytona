// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"context"

	"github.com/daytonaio/daytona/pkg/jobs"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/runner/providermanager"
	"github.com/daytonaio/daytona/pkg/telemetry"
)

type ITargetJobFactory interface {
	Create(job models.Job) jobs.IJob
}

type TargetJobFactory struct {
	config TargetJobFactoryConfig
}

type TargetJobFactoryConfig struct {
	FindTarget               func(ctx context.Context, targetId string) (*models.Target, error)
	HandleSuccessfulCreation func(ctx context.Context, targetId string) error

	TrackTelemetryEvent          func(event telemetry.ServerEvent, clientId string, props map[string]interface{}) error
	UpdateTargetProviderMetadata func(ctx context.Context, targetId, metadata string) error

	LoggerFactory   logs.ILoggerFactory
	ProviderManager providermanager.IProviderManager
}

func NewTargetJobFactory(config TargetJobFactoryConfig) ITargetJobFactory {
	return &TargetJobFactory{
		config: config,
	}
}

func (f *TargetJobFactory) Create(job models.Job) jobs.IJob {
	return &TargetJob{
		Job: job,

		findTarget:                   f.config.FindTarget,
		handleSuccessfulCreation:     f.config.HandleSuccessfulCreation,
		trackTelemetryEvent:          f.config.TrackTelemetryEvent,
		updateTargetProviderMetadata: f.config.UpdateTargetProviderMetadata,
		loggerFactory:                f.config.LoggerFactory,
		providerManager:              f.config.ProviderManager,
	}
}
