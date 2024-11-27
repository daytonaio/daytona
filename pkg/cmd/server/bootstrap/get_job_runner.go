// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package bootstrap

import (
	"context"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/build"
	"github.com/daytonaio/daytona/pkg/jobs/target"
	"github.com/daytonaio/daytona/pkg/jobs/workspace"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/provider/manager"
	"github.com/daytonaio/daytona/pkg/provisioner"
	"github.com/daytonaio/daytona/pkg/runners"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/daytonaio/daytona/pkg/telemetry"

	"github.com/daytonaio/daytona/pkg/runners/runner"
)

func GetJobRunner(c *server.Config, configDir string, version string, telemetryService telemetry.TelemetryService) (runners.IJobRunner, error) {
	jobService := server.GetInstance(nil).JobService

	workspaceJobFactory, err := GetWorkspaceJobFactory(c, configDir, version, telemetryService)
	if err != nil {
		return nil, err
	}

	targetJobFactory, err := GetTargetJobFactory(c, configDir, version, telemetryService)
	if err != nil {
		return nil, err
	}

	return runner.NewJobRunner(runner.JobRunnerConfig{
		ListPendingJobs: func(ctx context.Context) ([]*models.Job, error) {
			return jobService.List(&stores.JobFilter{
				States: &[]models.JobState{models.JobStatePending},
			})
		},
		UpdateJobState: func(ctx context.Context, job *models.Job, state models.JobState, err *error) error {
			job.State = state
			if err != nil {
				job.Error = util.Pointer((*err).Error())
			}
			return jobService.Save(job)
		},
		WorkspaceJobFactory: workspaceJobFactory,
		TargetJobFactory:    targetJobFactory,
	}), nil
}

func GetWorkspaceJobFactory(c *server.Config, configDir string, version string, telemetryService telemetry.TelemetryService) (workspace.IWorkspaceJobFactory, error) {
	containerRegistryService := server.GetInstance(nil).ContainerRegistryService

	gitProviderService := server.GetInstance(nil).GitProviderService

	targetLogsDir, err := server.GetTargetLogsDir(configDir)
	if err != nil {
		return nil, err
	}
	buildLogsDir, err := build.GetBuildLogsDir()
	if err != nil {
		return nil, err
	}
	loggerFactory := logs.NewLoggerFactory(&targetLogsDir, &buildLogsDir)

	providerManager := manager.GetProviderManager(nil)

	provisioner := provisioner.NewProvisioner(provisioner.ProvisionerConfig{
		ProviderManager: providerManager,
	})

	targetService := server.GetInstance(nil).TargetService

	workspaceService := server.GetInstance(nil).WorkspaceService

	return workspace.NewWorkspaceJobFactory(workspace.WorkspaceJobFactoryConfig{
		FindWorkspace: func(ctx context.Context, workspaceId string) (*models.Workspace, error) {
			workspaceDto, err := workspaceService.GetWorkspace(ctx, workspaceId, services.WorkspaceRetrievalParams{Verbose: false})
			if err != nil {
				return nil, err
			}
			return &workspaceDto.Workspace, nil
		},
		FindTarget: func(ctx context.Context, targetId string) (*models.Target, error) {
			targetDto, err := targetService.GetTarget(ctx, &stores.TargetFilter{IdOrName: &targetId}, services.TargetRetrievalParams{})
			if err != nil {
				return nil, err
			}
			return &targetDto.Target, nil
		},
		FindContainerRegistry: func(ctx context.Context, image string) (*models.ContainerRegistry, error) {
			return containerRegistryService.Find(image)
		},
		FindGitProviderConfig: func(ctx context.Context, id string) (*models.GitProviderConfig, error) {
			return gitProviderService.GetConfig(id)
		},
		TrackTelemetryEvent: func(event telemetry.ServerEvent, clientId string, props map[string]interface{}) error {
			return telemetryService.TrackServerEvent(event, clientId, props)
		},
		LoggerFactory: loggerFactory,
		Provisioner:   provisioner,
	}), nil
}

func GetTargetJobFactory(c *server.Config, configDir string, version string, telemetryService telemetry.TelemetryService) (target.ITargetJobFactory, error) {
	targetLogsDir, err := server.GetTargetLogsDir(configDir)
	if err != nil {
		return nil, err
	}
	buildLogsDir, err := build.GetBuildLogsDir()
	if err != nil {
		return nil, err
	}
	loggerFactory := logs.NewLoggerFactory(&targetLogsDir, &buildLogsDir)

	providerManager := manager.GetProviderManager(nil)

	provisioner := provisioner.NewProvisioner(provisioner.ProvisionerConfig{
		ProviderManager: providerManager,
	})

	targetService := server.GetInstance(nil).TargetService

	return target.NewTargetJobFactory(target.TargetJobFactoryConfig{
		FindTarget: func(ctx context.Context, targetId string) (*models.Target, error) {
			targetDto, err := targetService.GetTarget(ctx, &stores.TargetFilter{IdOrName: &targetId}, services.TargetRetrievalParams{})
			if err != nil {
				return nil, err
			}
			return &targetDto.Target, nil
		},
		HandleSuccessfulCreation: func(ctx context.Context, targetId string) error {
			return targetService.HandleSuccessfulCreation(ctx, targetId)
		},
		TrackTelemetryEvent: func(event telemetry.ServerEvent, clientId string, props map[string]interface{}) error {
			return telemetryService.TrackServerEvent(event, clientId, props)
		},
		LoggerFactory: loggerFactory,
		Provisioner:   provisioner,
	}), nil
}
