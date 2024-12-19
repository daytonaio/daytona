// Copyright 2024 Daytona Platforms Inparams.ServerConfig.
// SPDX-License-Identifier: Apache-2.0

package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/internal/util/apiclient/conversion"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/build"
	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/docker"
	jobs_build "github.com/daytonaio/daytona/pkg/jobs/build"
	jobs_runner "github.com/daytonaio/daytona/pkg/jobs/runner"
	"github.com/daytonaio/daytona/pkg/jobs/target"
	"github.com/daytonaio/daytona/pkg/jobs/workspace"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/runner/providermanager"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/telemetry"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"

	"github.com/daytonaio/daytona/pkg/runner"
)

type RemoteRunnerParams struct {
	ApiClient        *apiclient.APIClient
	ServerConfig     *apiclient.ServerConfig
	RunnerConfig     *runner.Config
	ConfigDir        string
	LogWriter        io.Writer
	TelemetryService telemetry.TelemetryService
}

type RemoteJobFactoryParams struct {
	ApiClient        *apiclient.APIClient
	ServerConfig     *apiclient.ServerConfig
	RunnerConfig     *runner.Config
	ConfigDir        string
	TelemetryService telemetry.TelemetryService
	ProviderManager  providermanager.IProviderManager
}

func GetRemoteRunner(params RemoteRunnerParams) (runner.IRunner, error) {
	runnerLogsDir := runner.GetLogsDir(params.ConfigDir)
	loggerFactory := logs.NewLoggerFactory(logs.LoggerFactoryConfig{
		LogsDir:     runnerLogsDir,
		ApiUrl:      &params.RunnerConfig.ServerApiUrl,
		ApiKey:      &params.RunnerConfig.ServerApiKey,
		ApiBasePath: &logs.ApiBasePathRunner,
	})

	runnerLogger, err := loggerFactory.CreateLogger(params.RunnerConfig.Id, params.RunnerConfig.Name, logs.LogSourceRunner)
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

	providerManager := getRemoteProviderManager(params, logger)

	jobFactoryParams := RemoteJobFactoryParams{
		ApiClient:        params.ApiClient,
		ServerConfig:     params.ServerConfig,
		RunnerConfig:     params.RunnerConfig,
		ConfigDir:        params.ConfigDir,
		TelemetryService: params.TelemetryService,
		ProviderManager:  providerManager,
	}

	workspaceJobFactory, err := getRemoteWorkspaceJobFactory(jobFactoryParams)
	if err != nil {
		return nil, err
	}

	targetJobFactory, err := getRemoteTargetJobFactory(jobFactoryParams)
	if err != nil {
		return nil, err
	}

	buildJobFactory, err := getRemoteBuildJobFactory(jobFactoryParams)
	if err != nil {
		return nil, err
	}

	runnerJobFactory := jobs_runner.NewRunnerJobFactory(jobs_runner.RunnerJobFactoryConfig{
		TrackTelemetryEvent: func(event telemetry.BuildRunnerEvent, clientId string, props map[string]interface{}) error {
			return params.TelemetryService.TrackRunnerEvent(event, clientId, props)
		},
		ProviderManager: providerManager,
	})

	return runner.NewRunner(runner.RunnerConfig{
		Config:          params.RunnerConfig,
		Logger:          logger,
		ProviderManager: providerManager,
		RegistryUrl:     params.ServerConfig.RegistryUrl,
		ListPendingJobs: func(ctx context.Context) ([]*models.Job, int, error) {
			jobs, res, err := params.ApiClient.RunnerAPI.ListRunnerJobs(ctx, params.RunnerConfig.Id).Execute()
			if err != nil {
				statusCode := -1
				if res != nil {
					statusCode = res.StatusCode
				}
				return nil, statusCode, err
			}

			var response []*models.Job
			for _, job := range jobs {
				response = append(response, &models.Job{
					Id:           job.Id,
					ResourceId:   job.ResourceId,
					RunnerId:     job.RunnerId,
					ResourceType: models.ResourceType(job.ResourceType),
					State:        models.JobState(job.State),
					Action:       models.JobAction(job.Action),
					Metadata:     job.Metadata,
					Error:        job.Error,
					// TODO: Convert
					// CreatedAt:    parseTime(job.CreatedAt),
					// UpdatedAt:    parseTime(job.UpdatedAt),
				})
			}
			return response, res.StatusCode, nil
		},
		UpdateJobState: func(ctx context.Context, jobId string, state models.JobState, jobError *error) error {
			var jobErr *string
			if jobError != nil {
				jobErr = util.Pointer((*jobError).Error())
			}
			_, err := params.ApiClient.RunnerAPI.UpdateJobState(ctx, params.RunnerConfig.Id, jobId).UpdateJobState(apiclient.UpdateJobState{
				State:        apiclient.JobState(state),
				ErrorMessage: jobErr,
			}).Execute()
			return err
		},
		SetRunnerMetadata: func(ctx context.Context, runnerId string, metadata models.RunnerMetadata) error {
			var providers []apiclient.ProviderInfo

			for _, provider := range metadata.Providers {
				providerInfoDto, err := conversion.Convert[models.ProviderInfo, apiclient.ProviderInfo](&provider)
				if err != nil {
					return err
				}
				if providerInfoDto == nil {
					continue
				}

				providers = append(providers, *providerInfoDto)
			}

			setRunnerMetadata := apiclient.SetRunnerMetadata{
				Uptime:    int32(metadata.Uptime),
				Providers: providers,
			}

			if metadata.RunningJobs != nil {
				setRunnerMetadata.RunningJobs = util.Pointer(int32(*metadata.RunningJobs))
			}

			_, err := params.ApiClient.RunnerAPI.SetRunnerMetadata(ctx, runnerId).SetMetadata(setRunnerMetadata).Execute()
			return err
		},
		WorkspaceJobFactory: workspaceJobFactory,
		TargetJobFactory:    targetJobFactory,
		BuildJobFactory:     buildJobFactory,
		RunnerJobFactory:    runnerJobFactory,
	}), nil
}

func getRemoteProviderManager(params RemoteRunnerParams, logger *log.Logger) providermanager.IProviderManager {
	headscaleUrl := util.GetFrpcHeadscaleUrl(params.ServerConfig.Frps.Protocol, params.ServerConfig.Id, params.ServerConfig.Frps.Domain)
	binaryUrl, _ := url.JoinPath(params.RunnerConfig.ServerApiUrl, "binary", "script")

	return providermanager.GetProviderManager(&providermanager.ProviderManagerConfig{
		WorkspaceLogsDir:   filepath.Join(params.ConfigDir, "workspaces", "logs"),
		TargetLogsDir:      filepath.Join(params.ConfigDir, "targets", "logs"),
		ApiUrl:             util.GetFrpcApiUrl(params.ServerConfig.Frps.Protocol, params.ServerConfig.Id, params.ServerConfig.Frps.Domain),
		ApiKey:             &params.RunnerConfig.ServerApiKey,
		RunnerId:           params.RunnerConfig.Id,
		RunnerName:         params.RunnerConfig.Name,
		Logger:             logger,
		DaytonaDownloadUrl: binaryUrl,
		ServerUrl:          headscaleUrl,
		BaseDir:            params.RunnerConfig.ProvidersDir,
		CreateProviderNetworkKey: func(ctx context.Context, providerName string) (string, error) {
			key, _, err := params.ApiClient.ServerAPI.GenerateNetworkKey(ctx).Execute()
			if err != nil {
				return "", err
			}

			return key.Key, nil
		},
		ServerPort: uint32(params.ServerConfig.HeadscalePort),
		ApiPort:    uint32(params.ServerConfig.ApiPort),
		GetTargetConfigMap: func(ctx context.Context) (map[string]*models.TargetConfig, error) {
			list, _, err := params.ApiClient.TargetConfigAPI.ListTargetConfigs(ctx).Execute()
			if err != nil {
				return nil, err
			}

			targetConfigs := make(map[string]*models.TargetConfig)
			for _, targetConfig := range list {
				tc, err := conversion.Convert[apiclient.TargetConfig, models.TargetConfig](&targetConfig)
				if err != nil {
					return nil, err
				}
				if tc == nil {
					continue
				}

				if tc.ProviderInfo.RunnerId != params.RunnerConfig.Id {
					continue
				}

				targetConfigs[targetConfig.Name] = tc
			}

			return targetConfigs, nil
		},
		CreateTargetConfig: func(ctx context.Context, name, options string, providerInfo models.ProviderInfo) error {
			providerInfoDto, err := conversion.Convert[models.ProviderInfo, apiclient.ProviderInfo](&providerInfo)
			if err != nil {
				return err
			}
			if providerInfoDto == nil {
				return errors.New("invalid provider info")
			}

			_, _, err = params.ApiClient.TargetConfigAPI.AddTargetConfig(ctx).TargetConfig(apiclient.AddTargetConfigDTO{
				Name:         fmt.Sprintf("%s-runner-%s", name, params.RunnerConfig.Id),
				Options:      options,
				ProviderInfo: *providerInfoDto,
			}).Execute()
			return err
		},
	})
}

func getRemoteWorkspaceJobFactory(params RemoteJobFactoryParams) (workspace.IWorkspaceJobFactory, error) {
	logsDir := filepath.Join(params.ConfigDir, "workspaces", "logs")
	loggerFactory := logs.NewLoggerFactory(logs.LoggerFactoryConfig{
		LogsDir:     logsDir,
		ApiUrl:      &params.RunnerConfig.ServerApiUrl,
		ApiKey:      &params.RunnerConfig.ServerApiKey,
		ApiBasePath: &logs.ApiBasePathWorkspace,
	})

	return workspace.NewWorkspaceJobFactory(workspace.WorkspaceJobFactoryConfig{
		FindWorkspace: func(ctx context.Context, workspaceId string) (*models.Workspace, error) {
			workspaceDto, _, err := params.ApiClient.WorkspaceAPI.GetWorkspace(ctx, workspaceId).Execute()
			if err != nil {
				return nil, err
			}
			return conversion.Convert[apiclient.WorkspaceDTO, models.Workspace](workspaceDto)
		},
		FindTarget: func(ctx context.Context, targetId string) (*models.Target, error) {
			targetDto, _, err := params.ApiClient.TargetAPI.GetTarget(ctx, targetId).Execute()
			if err != nil {
				return nil, err
			}
			return conversion.Convert[apiclient.TargetDTO, models.Target](targetDto)
		},
		UpdateWorkspaceProviderMetadata: func(ctx context.Context, workspaceId, metadata string) error {
			_, err := params.ApiClient.WorkspaceAPI.UpdateWorkspaceProviderMetadata(ctx, workspaceId).Metadata(apiclient.UpdateWorkspaceProviderMetadataDTO{
				Metadata: metadata,
			}).Execute()
			return err
		},
		FindGitProviderConfig: func(ctx context.Context, id string) (*models.GitProviderConfig, error) {
			gp, _, err := params.ApiClient.GitProviderAPI.GetGitProvider(ctx, id).Execute()
			if err != nil {
				return nil, err
			}

			return conversion.Convert[apiclient.GitProvider, models.GitProviderConfig](gp)
		},
		GetWorkspaceEnvironmentVariables: func(ctx context.Context, w *models.Workspace) (map[string]string, error) {
			envVars, _, err := params.ApiClient.EnvVarAPI.ListEnvironmentVariables(ctx).Execute()
			if err != nil {
				return nil, err
			}

			envVarsMap := make(map[string]string)
			for _, envVar := range envVars {
				envVarsMap[envVar.Key] = envVar.Value
			}

			return util.MergeEnvVars(envVarsMap, w.EnvVars), nil
		},
		TrackTelemetryEvent: func(event telemetry.ServerEvent, clientId string, props map[string]interface{}) error {
			return params.TelemetryService.TrackServerEvent(event, clientId, props)
		},
		LoggerFactory:   loggerFactory,
		ProviderManager: params.ProviderManager,
		BuilderImage:    params.ServerConfig.BuilderImage,
	}), nil
}

func getRemoteTargetJobFactory(params RemoteJobFactoryParams) (target.ITargetJobFactory, error) {
	logsDir := filepath.Join(params.ConfigDir, "targets", "logs")
	loggerFactory := logs.NewLoggerFactory(logs.LoggerFactoryConfig{
		LogsDir:     logsDir,
		ApiUrl:      &params.RunnerConfig.ServerApiUrl,
		ApiKey:      &params.RunnerConfig.ServerApiKey,
		ApiBasePath: &logs.ApiBasePathTarget,
	})

	return target.NewTargetJobFactory(target.TargetJobFactoryConfig{
		FindTarget: func(ctx context.Context, targetId string) (*models.Target, error) {
			targetDto, _, err := params.ApiClient.TargetAPI.GetTarget(ctx, targetId).Execute()
			if err != nil {
				return nil, err
			}

			return conversion.Convert[apiclient.TargetDTO, models.Target](targetDto)
		},
		HandleSuccessfulCreation: func(ctx context.Context, targetId string) error {
			_, err := params.ApiClient.TargetAPI.HandleSuccessfulCreation(ctx, targetId).Execute()
			return err
		},
		UpdateTargetProviderMetadata: func(ctx context.Context, targetId, metadata string) error {
			_, err := params.ApiClient.TargetAPI.UpdateTargetProviderMetadata(ctx, targetId).Metadata(apiclient.UpdateTargetProviderMetadataDTO{
				Metadata: metadata,
			}).Execute()
			return err
		},
		TrackTelemetryEvent: func(event telemetry.ServerEvent, clientId string, props map[string]interface{}) error {
			return params.TelemetryService.TrackServerEvent(event, clientId, props)
		},
		LoggerFactory:   loggerFactory,
		ProviderManager: params.ProviderManager,
	}), nil
}

func getRemoteBuildJobFactory(params RemoteJobFactoryParams) (jobs_build.IBuildJobFactory, error) {
	loggerFactory := logs.NewLoggerFactory(logs.LoggerFactoryConfig{
		LogsDir:     filepath.Join(params.ConfigDir, "builds", "logs"),
		ApiUrl:      &params.RunnerConfig.ServerApiUrl,
		ApiKey:      &params.RunnerConfig.ServerApiKey,
		ApiBasePath: &logs.ApiBasePathBuild,
	})

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	dockerClient := docker.NewDockerClient(docker.DockerClientConfig{
		ApiClient: cli,
	})

	var buildImageNamespace string

	if params.ServerConfig.BuildImageNamespace != nil {
		buildImageNamespace = *params.ServerConfig.BuildImageNamespace
		if buildImageNamespace != "" {
			buildImageNamespace = fmt.Sprintf("/%s", buildImageNamespace)
			buildImageNamespace = strings.TrimSuffix(buildImageNamespace, "/")
		}
	}

	var builderRegistry *models.ContainerRegistry

	envVars, _, err := params.ApiClient.EnvVarAPI.ListEnvironmentVariables(context.Background()).Execute()
	if err != nil {
		builderRegistry = &models.ContainerRegistry{
			Server: params.ServerConfig.BuilderRegistryServer,
		}
	}

	envVarsMap := make(services.EnvironmentVariables)
	for _, envVar := range envVars {
		envVarsMap[envVar.Key] = envVar.Value
	}

	if len(envVarsMap) > 0 {
		builderRegistry = envVarsMap.FindContainerRegistry(params.ServerConfig.BuilderRegistryServer)
	}

	if builderRegistry == nil {
		builderRegistry = &models.ContainerRegistry{
			Server: util.GetFrpcRegistryDomain(params.ServerConfig.Id, params.ServerConfig.Frps.Domain),
		}
	}

	_, containerRegistries := common.ExtractContainerRegistryFromEnvVars(envVarsMap)

	return jobs_build.NewBuildJobFactory(jobs_build.BuildJobFactoryConfig{
		FindBuild: func(ctx context.Context, buildId string) (*services.BuildDTO, error) {
			build, _, err := params.ApiClient.BuildAPI.GetBuild(ctx, buildId).Execute()
			if err != nil {
				return nil, err
			}

			return conversion.Convert[apiclient.BuildDTO, services.BuildDTO](build)
		},
		ListSuccessfulBuilds: func(ctx context.Context, repoUrl string) ([]*models.Build, error) {
			apiclientBuildDtos, _, err := params.ApiClient.BuildAPI.ListSuccessfulBuilds(ctx, url.QueryEscape(repoUrl)).Execute()
			if err != nil {
				return nil, err
			}

			var builds []*models.Build
			for _, apiclientBuildDto := range apiclientBuildDtos {
				buildDto, err := conversion.Convert[apiclient.BuildDTO, services.BuildDTO](&apiclientBuildDto)
				if err != nil {
					return nil, err
				}
				if buildDto == nil {
					continue
				}
				builds = append(builds, &buildDto.Build)
			}
			return builds, nil
		},
		ListConfigsForUrl: func(ctx context.Context, repoUrl string) ([]*models.GitProviderConfig, error) {
			gitProviders, _, err := params.ApiClient.GitProviderAPI.ListGitProvidersForUrl(ctx, url.QueryEscape(repoUrl)).Execute()
			if err != nil {
				return nil, err
			}

			var gitProviderConfigs []*models.GitProviderConfig
			for _, gitProvider := range gitProviders {
				gitProviderConfigDto, err := conversion.Convert[apiclient.GitProvider, models.GitProviderConfig](&gitProvider)
				if err != nil {
					return nil, err
				}
				if gitProviderConfigDto == nil {
					continue
				}
				gitProviderConfigs = append(gitProviderConfigs, gitProviderConfigDto)
			}

			return gitProviderConfigs, nil
		},
		CheckImageExists: func(ctx context.Context, image string) bool {
			_, _, err = cli.ImageInspectWithRaw(ctx, image)
			return err == nil
		},
		DeleteImage: func(ctx context.Context, image string, force bool) error {
			return dockerClient.DeleteImage(image, force, nil)
		},
		TrackTelemetryEvent: func(event telemetry.BuildRunnerEvent, clientId string, props map[string]interface{}) error {
			return params.TelemetryService.TrackBuildRunnerEvent(event, clientId, props)
		},
		LoggerFactory: loggerFactory,
		BuilderFactory: build.NewBuilderFactory(build.BuilderFactoryConfig{
			Image:                       params.ServerConfig.BuilderImage,
			ContainerRegistries:         containerRegistries,
			BuildImageContainerRegistry: builderRegistry,
			BuildImageNamespace:         buildImageNamespace,
			LoggerFactory:               loggerFactory,
			DefaultWorkspaceImage:       params.ServerConfig.DefaultWorkspaceImage,
			DefaultWorkspaceUser:        params.ServerConfig.DefaultWorkspaceUser,
		}),
		BasePath: filepath.Join(params.ConfigDir, "builds"),
	}), nil
}
