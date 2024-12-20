// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"context"

	"github.com/daytonaio/daytona/pkg/build"
	"github.com/daytonaio/daytona/pkg/jobs"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/telemetry"
)

type IBuildJobFactory interface {
	Create(job models.Job) jobs.IJob
}

type BuildJobFactory struct {
	config BuildJobFactoryConfig
}

type BuildJobFactoryConfig struct {
	FindBuild            func(ctx context.Context, buildId string) (*services.BuildDTO, error)
	ListSuccessfulBuilds func(ctx context.Context, repoUrl string) ([]*models.Build, error)
	ListConfigsForUrl    func(ctx context.Context, repoUrl string) ([]*models.GitProviderConfig, error)
	CheckImageExists     func(ctx context.Context, image string) bool
	DeleteImage          func(ctx context.Context, image string, force bool) error

	TrackTelemetryEvent func(event telemetry.BuildRunnerEvent, clientId string, props map[string]interface{}) error
	LoggerFactory       logs.ILoggerFactory
	BuilderFactory      build.IBuilderFactory

	BasePath string
}

func NewBuildJobFactory(config BuildJobFactoryConfig) IBuildJobFactory {
	return &BuildJobFactory{
		config: config,
	}
}

func (f *BuildJobFactory) Create(job models.Job) jobs.IJob {
	return &BuildJob{
		Job: job,

		findBuild:            f.config.FindBuild,
		listSuccessfulBuilds: f.config.ListSuccessfulBuilds,
		listConfigsForUrl:    f.config.ListConfigsForUrl,
		checkImageExists:     f.config.CheckImageExists,
		deleteImage:          f.config.DeleteImage,

		trackTelemetryEvent: f.config.TrackTelemetryEvent,
		loggerFactory:       f.config.LoggerFactory,
		builderFactory:      f.config.BuilderFactory,
		basePath:            f.config.BasePath,
	}
}
