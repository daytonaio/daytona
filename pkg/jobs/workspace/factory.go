// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"

	"github.com/daytonaio/daytona/pkg/jobs"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/provisioner"
	"github.com/daytonaio/daytona/pkg/telemetry"
)

type IWorkspaceJobFactory interface {
	Create(job models.Job) jobs.IJob
}

type WorkspaceJobFactory struct {
	config WorkspaceJobFactoryConfig
}

type WorkspaceJobFactoryConfig struct {
	FindWorkspace         func(ctx context.Context, workspaceId string) (*models.Workspace, error)
	FindTarget            func(ctx context.Context, targetId string) (*models.Target, error)
	FindContainerRegistry func(ctx context.Context, image string) (*models.ContainerRegistry, error)
	FindGitProviderConfig func(ctx context.Context, id string) (*models.GitProviderConfig, error)
	TrackTelemetryEvent   func(event telemetry.ServerEvent, clientId string, props map[string]interface{}) error

	LoggerFactory logs.LoggerFactory
	Provisioner   provisioner.IProvisioner
}

func NewWorkspaceJobFactory(config WorkspaceJobFactoryConfig) IWorkspaceJobFactory {
	return &WorkspaceJobFactory{
		config: config,
	}
}

func (f *WorkspaceJobFactory) Create(job models.Job) jobs.IJob {
	return &WorkspaceJob{
		Job: job,

		findWorkspace:         f.config.FindWorkspace,
		findTarget:            f.config.FindTarget,
		findContainerRegistry: f.config.FindContainerRegistry,
		findGitProviderConfig: f.config.FindGitProviderConfig,
		trackTelemetryEvent:   f.config.TrackTelemetryEvent,
		loggerFactory:         f.config.LoggerFactory,
		provisioner:           f.config.Provisioner,
	}
}
