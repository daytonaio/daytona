// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package bootstrap

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/build"
	"github.com/daytonaio/daytona/pkg/docker"
	jobs_build "github.com/daytonaio/daytona/pkg/jobs/build"
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
	"github.com/docker/docker/client"

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

	buildJobFactory, err := GetBuildJobFactory(c, configDir, version, telemetryService)
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
			return jobService.Update(job)
		},
		WorkspaceJobFactory: workspaceJobFactory,
		TargetJobFactory:    targetJobFactory,
		BuildJobFactory:     buildJobFactory,
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
		BuilderImage:  c.BuilderImage,
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

func GetBuildJobFactory(c *server.Config, configDir string, version string, telemetryService telemetry.TelemetryService) (jobs_build.IBuildJobFactory, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	dockerClient := docker.NewDockerClient(docker.DockerClientConfig{
		ApiClient: cli,
	})

	logsDir, err := build.GetBuildLogsDir()
	if err != nil {
		return nil, err
	}
	loggerFactory := logs.NewLoggerFactory(nil, &logsDir)

	buildService := server.GetInstance(nil).BuildService

	containerRegistryService := server.GetInstance(nil).ContainerRegistryService

	cr, err := containerRegistryService.FindByImageName(c.BuilderImage)
	if err != nil && !stores.IsContainerRegistryNotFound(err) {
		return nil, err
	}

	buildImageNamespace := c.BuildImageNamespace
	if buildImageNamespace != "" {
		buildImageNamespace = fmt.Sprintf("/%s", buildImageNamespace)
	}
	buildImageNamespace = strings.TrimSuffix(buildImageNamespace, "/")

	var builderRegistry *models.ContainerRegistry

	if c.BuilderRegistryServer != "local" {
		builderRegistry, err = containerRegistryService.Find(c.BuilderRegistryServer)
		if err != nil {
			builderRegistry = &models.ContainerRegistry{
				Server: c.BuilderRegistryServer,
			}
		}
	}

	return jobs_build.NewBuildJobFactory(jobs_build.BuildJobFactoryConfig{
		FindBuild: func(ctx context.Context, buildId string) (*services.BuildDTO, error) {
			return buildService.Find(&services.BuildFilter{
				StoreFilter: stores.BuildFilter{
					Id: &buildId,
				},
			})
		},
		ListSuccessfulBuilds: func(ctx context.Context) ([]*models.Build, error) {
			buildDtos, err := buildService.List(&services.BuildFilter{
				StateNames: &[]models.ResourceStateName{models.ResourceStateNameRunSuccessful},
			})
			if err != nil {
				return nil, err
			}

			var builds []*models.Build
			for _, buildDto := range buildDtos {
				builds = append(builds, &buildDto.Build)
			}
			return builds, nil
		},
		ListConfigsForUrl: func(ctx context.Context, repoUrl string) ([]*models.GitProviderConfig, error) {
			return server.GetInstance(nil).GitProviderService.ListConfigsForUrl(repoUrl)
		},
		CheckImageExists: func(ctx context.Context, image string) bool {
			_, _, err = cli.ImageInspectWithRaw(context.Background(), image)
			return err == nil
		},
		DeleteImage: func(ctx context.Context, image string, force bool) error {
			return dockerClient.DeleteImage(image, force, nil)
		},
		TrackTelemetryEvent: func(event telemetry.BuildRunnerEvent, clientId string, props map[string]interface{}) error {
			return telemetryService.TrackBuildRunnerEvent(event, clientId, props)
		},
		LoggerFactory: loggerFactory,
		BuilderFactory: build.NewBuilderFactory(build.BuilderFactoryConfig{
			Image:                       c.BuilderImage,
			ContainerRegistry:           cr,
			BuildImageContainerRegistry: builderRegistry,
			BuildService:                buildService,
			BuildImageNamespace:         buildImageNamespace,
			LoggerFactory:               loggerFactory,
			DefaultWorkspaceImage:       c.DefaultWorkspaceImage,
			DefaultWorkspaceUser:        c.DefaultWorkspaceUser,
		}),
		BasePath: filepath.Join(configDir, "builds"),
	}), nil
}
