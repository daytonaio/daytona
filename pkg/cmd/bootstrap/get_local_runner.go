// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package bootstrap

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/build"
	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/db"
	"github.com/daytonaio/daytona/pkg/docker"
	jobs_build "github.com/daytonaio/daytona/pkg/jobs/build"
	jobs_runner "github.com/daytonaio/daytona/pkg/jobs/runner"
	"github.com/daytonaio/daytona/pkg/jobs/target"
	"github.com/daytonaio/daytona/pkg/jobs/workspace"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/runner"
	"github.com/daytonaio/daytona/pkg/runner/providermanager"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/server/headscale"
	"github.com/daytonaio/daytona/pkg/server/targetconfigs"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/daytonaio/daytona/pkg/telemetry"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
)

const LOCAL_RUNNER_ID = "local"

type LocalRunnerParams struct {
	ServerConfig     *server.Config
	RunnerConfig     *runner.Config
	ConfigDir        string
	TelemetryService telemetry.TelemetryService
}

type LocalJobFactoryParams struct {
	ServerConfig     *server.Config
	ConfigDir        string
	TelemetryService telemetry.TelemetryService
	ProviderManager  providermanager.IProviderManager
}

func GetLocalRunner(params LocalRunnerParams) (runner.IRunner, error) {
	loggerFactory := logs.NewLoggerFactory(logs.LoggerFactoryConfig{LogsDir: server.GetRunnerLogsDir(params.ConfigDir)})

	runnerLogger, err := loggerFactory.CreateLogger(LOCAL_RUNNER_ID, LOCAL_RUNNER_ID, logs.LogSourceRunner)
	if err != nil {
		return nil, err
	}

	logger := &log.Logger{
		Out: io.MultiWriter(runnerLogger, os.Stdout),
		Formatter: &log.TextFormatter{
			ForceColors: true,
		},
		Hooks: make(log.LevelHooks),
		Level: log.DebugLevel,
	}

	jobService := server.GetInstance(nil).JobService

	providerManager, err := getProviderManager(params, logger)
	if err != nil {
		return nil, err
	}

	jobFactoryParams := LocalJobFactoryParams{
		ServerConfig:     params.ServerConfig,
		ConfigDir:        params.ConfigDir,
		TelemetryService: params.TelemetryService,
		ProviderManager:  providerManager,
	}

	runnerService := server.GetInstance(nil).RunnerService

	workspaceJobFactory, err := getLocalWorkspaceJobFactory(jobFactoryParams)
	if err != nil {
		return nil, err
	}

	targetJobFactory, err := getLocalTargetJobFactory(jobFactoryParams)
	if err != nil {
		return nil, err
	}

	buildJobFactory, err := getLocalBuildJobFactory(jobFactoryParams)
	if err != nil {
		return nil, err
	}

	runnerJobFactory := jobs_runner.NewRunnerJobFactory(jobs_runner.RunnerJobFactoryConfig{
		TrackTelemetryEvent: func(event telemetry.BuildRunnerEvent, clientId string, props map[string]interface{}) error {
			return params.TelemetryService.TrackBuildRunnerEvent(event, clientId, props)
		},
		ProviderManager: providerManager,
	})

	return runner.NewRunner(runner.RunnerConfig{
		Config:          params.RunnerConfig,
		Logger:          logger,
		ProviderManager: providerManager,
		RegistryUrl:     params.ServerConfig.RegistryUrl,
		ListPendingJobs: func(ctx context.Context) ([]*models.Job, int, error) {
			jobs, err := jobService.List(ctx, &stores.JobFilter{
				RunnerIdOrIsNil: util.Pointer(LOCAL_RUNNER_ID),
				States:          &[]models.JobState{models.JobStatePending},
			})
			return jobs, 0, err
		},
		UpdateJobState: func(ctx context.Context, jobId string, state models.JobState, err *error) error {
			var jobErr *string
			if err != nil {
				jobErr = util.Pointer((*err).Error())
			}
			return jobService.SetState(ctx, jobId, services.UpdateJobStateDTO{
				State:        state,
				ErrorMessage: jobErr,
			})
		},
		SetRunnerMetadata: func(ctx context.Context, runnerId string, metadata models.RunnerMetadata) error {
			return runnerService.SetRunnerMetadata(context.Background(), runnerId, &models.RunnerMetadata{
				Uptime:      uint64(metadata.Uptime),
				Providers:   metadata.Providers,
				RunningJobs: metadata.RunningJobs,
			})
		},
		WorkspaceJobFactory: workspaceJobFactory,
		TargetJobFactory:    targetJobFactory,
		BuildJobFactory:     buildJobFactory,
		RunnerJobFactory:    runnerJobFactory,
	}), nil
}

func getLocalWorkspaceJobFactory(params LocalJobFactoryParams) (workspace.IWorkspaceJobFactory, error) {
	workspaceLogsFactory := logs.NewLoggerFactory(logs.LoggerFactoryConfig{LogsDir: server.GetWorkspaceLogsDir(params.ConfigDir)})

	envVarService := server.GetInstance(nil).EnvironmentVariableService

	gitProviderService := server.GetInstance(nil).GitProviderService

	targetService := server.GetInstance(nil).TargetService

	workspaceService := server.GetInstance(nil).WorkspaceService

	return workspace.NewWorkspaceJobFactory(workspace.WorkspaceJobFactoryConfig{
		FindWorkspace: func(ctx context.Context, workspaceId string) (*models.Workspace, error) {
			workspaceDto, err := workspaceService.GetWorkspace(ctx, workspaceId, services.WorkspaceRetrievalParams{})
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
		UpdateWorkspaceProviderMetadata: workspaceService.UpdateWorkspaceProviderMetadata,
		FindGitProviderConfig: func(ctx context.Context, id string) (*models.GitProviderConfig, error) {
			return gitProviderService.GetConfig(ctx, id)
		},
		GetWorkspaceEnvironmentVariables: func(ctx context.Context, w *models.Workspace) (map[string]string, error) {
			serverEnvVars, err := envVarService.Map(ctx)
			if err != nil {
				return nil, err
			}

			return util.MergeEnvVars(serverEnvVars, w.EnvVars), nil
		},
		TrackTelemetryEvent: func(event telemetry.ServerEvent, clientId string, props map[string]interface{}) error {
			return params.TelemetryService.TrackServerEvent(event, clientId, props)
		},
		LoggerFactory:   workspaceLogsFactory,
		ProviderManager: params.ProviderManager,
		BuilderImage:    params.ServerConfig.BuilderImage,
	}), nil
}

func getLocalTargetJobFactory(params LocalJobFactoryParams) (target.ITargetJobFactory, error) {
	targetLogsFactory := logs.NewLoggerFactory(logs.LoggerFactoryConfig{LogsDir: server.GetTargetLogsDir(params.ConfigDir)})

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
		UpdateTargetProviderMetadata: targetService.UpdateTargetProviderMetadata,
		TrackTelemetryEvent: func(event telemetry.ServerEvent, clientId string, props map[string]interface{}) error {
			return params.TelemetryService.TrackServerEvent(event, clientId, props)
		},
		LoggerFactory:   targetLogsFactory,
		ProviderManager: params.ProviderManager,
	}), nil
}

func getLocalBuildJobFactory(params LocalJobFactoryParams) (jobs_build.IBuildJobFactory, error) {
	buildLogsFactory := logs.NewLoggerFactory(logs.LoggerFactoryConfig{LogsDir: server.GetBuildLogsDir(params.ConfigDir)})

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	dockerClient := docker.NewDockerClient(docker.DockerClientConfig{
		ApiClient: cli,
	})

	buildService := server.GetInstance(nil).BuildService

	buildImageNamespace := params.ServerConfig.BuildImageNamespace
	if buildImageNamespace != "" {
		buildImageNamespace = fmt.Sprintf("/%s", buildImageNamespace)
	}
	buildImageNamespace = strings.TrimSuffix(buildImageNamespace, "/")

	var builderRegistry *models.ContainerRegistry

	envVarService := server.GetInstance(nil).EnvironmentVariableService

	envVars, err := envVarService.Map(context.Background())
	if err != nil {
		builderRegistry = &models.ContainerRegistry{
			Server: params.ServerConfig.BuilderRegistryServer,
		}
	} else {
		builderRegistry = envVars.FindContainerRegistry(params.ServerConfig.BuilderRegistryServer)
	}

	if builderRegistry == nil {
		builderRegistry = &models.ContainerRegistry{
			Server: util.GetFrpcRegistryDomain(params.ServerConfig.Id, params.ServerConfig.Frps.Domain),
		}
	}

	_, containerRegistries := common.ExtractContainerRegistryFromEnvVars(envVars)

	return jobs_build.NewBuildJobFactory(jobs_build.BuildJobFactoryConfig{
		FindBuild: func(ctx context.Context, buildId string) (*services.BuildDTO, error) {
			return buildService.Find(ctx, &services.BuildFilter{
				StoreFilter: stores.BuildFilter{
					Id: &buildId,
				},
			})
		},
		ListSuccessfulBuilds: func(ctx context.Context, repoUrl string) ([]*models.Build, error) {
			buildDtos, err := buildService.List(ctx, &services.BuildFilter{
				StateNames: &[]models.ResourceStateName{models.ResourceStateNameRunSuccessful},
				StoreFilter: stores.BuildFilter{
					RepositoryUrl: &repoUrl,
				},
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
			return server.GetInstance(nil).GitProviderService.ListConfigsForUrl(ctx, repoUrl)
		},
		CheckImageExists: func(ctx context.Context, image string) bool {
			_, _, err = cli.ImageInspectWithRaw(context.Background(), image)
			return err == nil
		},
		DeleteImage: func(ctx context.Context, image string, force bool) error {
			return dockerClient.DeleteImage(image, force, nil)
		},
		TrackTelemetryEvent: func(event telemetry.BuildRunnerEvent, clientId string, props map[string]interface{}) error {
			return params.TelemetryService.TrackBuildRunnerEvent(event, clientId, props)
		},
		LoggerFactory: buildLogsFactory,
		BuilderFactory: build.NewBuilderFactory(build.BuilderFactoryConfig{
			ContainerRegistries:         containerRegistries,
			Image:                       params.ServerConfig.BuilderImage,
			BuildImageContainerRegistry: builderRegistry,

			BuildImageNamespace:   buildImageNamespace,
			LoggerFactory:         buildLogsFactory,
			DefaultWorkspaceImage: params.ServerConfig.DefaultWorkspaceImage,
			DefaultWorkspaceUser:  params.ServerConfig.DefaultWorkspaceUser,
		}),
		BasePath: filepath.Join(params.ConfigDir, "builds"),
	}), nil
}

func getProviderManager(params LocalRunnerParams, logger *log.Logger) (providermanager.IProviderManager, error) {
	headscaleServer := headscale.NewHeadscaleServer(&headscale.HeadscaleServerConfig{
		ServerId:      params.ServerConfig.Id,
		FrpsDomain:    params.ServerConfig.Frps.Domain,
		FrpsProtocol:  params.ServerConfig.Frps.Protocol,
		HeadscalePort: params.ServerConfig.HeadscalePort,
		ConfigDir:     filepath.Join(params.ConfigDir, "headscale"),
		Frps:          params.ServerConfig.Frps,
	})
	err := headscaleServer.Init()
	if err != nil {
		return nil, err
	}

	headscaleUrl := util.GetFrpcHeadscaleUrl(params.ServerConfig.Frps.Protocol, params.ServerConfig.Id, params.ServerConfig.Frps.Domain)
	binaryUrl, _ := url.JoinPath(util.GetFrpcApiUrl(params.ServerConfig.Frps.Protocol, params.ServerConfig.Id, params.ServerConfig.Frps.Domain), "binary", "script")

	dbPath, err := getDbPath()
	if err != nil {
		return nil, err
	}

	dbConnection := db.GetSQLiteConnection(dbPath)

	store := db.NewStore(dbConnection)

	targetConfigStore, err := db.NewTargetConfigStore(store)
	if err != nil {
		return nil, err
	}

	targetConfigService := targetconfigs.NewTargetConfigService(targetconfigs.TargetConfigServiceConfig{
		TargetConfigStore: targetConfigStore,
	})

	return providermanager.GetProviderManager(&providermanager.ProviderManagerConfig{
		TargetLogsDir:      server.GetTargetLogsDir(params.ConfigDir),
		WorkspaceLogsDir:   server.GetWorkspaceLogsDir(params.ConfigDir),
		Logger:             logger,
		ApiUrl:             util.GetFrpcApiUrl(params.ServerConfig.Frps.Protocol, params.ServerConfig.Id, params.ServerConfig.Frps.Domain),
		RunnerName:         params.RunnerConfig.Name,
		RunnerId:           params.RunnerConfig.Id,
		DaytonaDownloadUrl: binaryUrl,
		ServerUrl:          headscaleUrl,
		BaseDir:            params.RunnerConfig.ProvidersDir,
		CreateProviderNetworkKey: func(ctx context.Context, providerName string) (string, error) {
			return headscaleServer.CreateAuthKey()
		},
		ServerPort: params.ServerConfig.HeadscalePort,
		ApiPort:    params.ServerConfig.ApiPort,
		GetTargetConfigMap: func(ctx context.Context) (map[string]*models.TargetConfig, error) {
			return targetConfigService.Map(ctx)
		},
		CreateTargetConfig: func(ctx context.Context, name, options string, providerInfo models.ProviderInfo) error {
			_, err := targetConfigService.Add(ctx, services.AddTargetConfigDTO{
				Name:         name,
				Options:      options,
				ProviderInfo: providerInfo,
			})
			return err
		},
	}), nil
}
