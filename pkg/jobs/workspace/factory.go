// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"

	"github.com/daytonaio/daytona/pkg/jobs"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/runner/providermanager"
	"github.com/daytonaio/daytona/pkg/telemetry"
)

type IWorkspaceJobFactory interface {
	Create(job models.Job) jobs.IJob
}

type WorkspaceJobFactory struct {
	config WorkspaceJobFactoryConfig
}

type WorkspaceJobFactoryConfig struct {
	FindWorkspace                    func(ctx context.Context, workspaceId string) (*models.Workspace, error)
	FindTarget                       func(ctx context.Context, targetId string) (*models.Target, error)
	FindGitProviderConfig            func(ctx context.Context, id string) (*models.GitProviderConfig, error)
	GetWorkspaceEnvironmentVariables func(ctx context.Context, w *models.Workspace) (map[string]string, error)
	UpdateWorkspaceProviderMetadata  func(ctx context.Context, workspaceId, metadata string) error
	TrackTelemetryEvent              func(event telemetry.ServerEvent, clientId string, props map[string]interface{}) error

	LoggerFactory   logs.ILoggerFactory
	ProviderManager providermanager.IProviderManager
	BuilderImage    string
}

func NewWorkspaceJobFactory(config WorkspaceJobFactoryConfig) IWorkspaceJobFactory {
	return &WorkspaceJobFactory{
		config: config,
	}
}

func (f *WorkspaceJobFactory) Create(job models.Job) jobs.IJob {
	return &WorkspaceJob{
		Job: job,

		findWorkspace:                    f.config.FindWorkspace,
		findTarget:                       f.config.FindTarget,
		findGitProviderConfig:            f.config.FindGitProviderConfig,
		getWorkspaceEnvironmentVariables: f.config.GetWorkspaceEnvironmentVariables,
		updateWorkspaceProviderMetadata:  f.config.UpdateWorkspaceProviderMetadata,
		trackTelemetryEvent:              f.config.TrackTelemetryEvent,
		loggerFactory:                    f.config.LoggerFactory,
		providerManager:                  f.config.ProviderManager,
		builderImage:                     f.config.BuilderImage,
	}
}
