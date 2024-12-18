// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/constants"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/db"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/server/apikeys"
	"github.com/daytonaio/daytona/pkg/server/builds"
	"github.com/daytonaio/daytona/pkg/server/env"
	"github.com/daytonaio/daytona/pkg/server/gitproviders"
	"github.com/daytonaio/daytona/pkg/server/headscale"
	"github.com/daytonaio/daytona/pkg/server/jobs"
	"github.com/daytonaio/daytona/pkg/server/registry"
	"github.com/daytonaio/daytona/pkg/server/runners"
	"github.com/daytonaio/daytona/pkg/server/targetconfigs"
	"github.com/daytonaio/daytona/pkg/server/targets"
	"github.com/daytonaio/daytona/pkg/server/workspaces"
	"github.com/daytonaio/daytona/pkg/server/workspacetemplates"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/daytonaio/daytona/pkg/telemetry"

	log "github.com/sirupsen/logrus"
)

func GetInstance(c *server.Config, configDir string, version string, telemetryService telemetry.TelemetryService) (*server.Server, error) {
	dbPath, err := getDbPath()
	if err != nil {
		return nil, err
	}

	dbConnection := db.GetSQLiteConnection(dbPath)

	store := db.NewStore(dbConnection)

	apiKeyStore, err := db.NewApiKeyStore(store)
	if err != nil {
		return nil, err
	}
	buildStore, err := db.NewBuildStore(store)
	if err != nil {
		return nil, err
	}
	workspaceTemplateStore, err := db.NewWorkspaceTemplateStore(store)
	if err != nil {
		return nil, err
	}
	gitProviderConfigStore, err := db.NewGitProviderConfigStore(store)
	if err != nil {
		return nil, err
	}
	targetConfigStore, err := db.NewTargetConfigStore(store)
	if err != nil {
		return nil, err
	}
	targetStore, err := db.NewTargetStore(store)
	if err != nil {
		return nil, err
	}
	targetMetadataStore, err := db.NewTargetMetadataStore(store)
	if err != nil {
		return nil, err
	}
	envVarStore, err := db.NewEnvironmentVariableStore(store)
	if err != nil {
		return nil, err
	}
	workspaceStore, err := db.NewWorkspaceStore(store)
	if err != nil {
		return nil, err
	}
	workspaceMetadataStore, err := db.NewWorkspaceMetadataStore(store)
	if err != nil {
		return nil, err
	}
	jobStore, err := db.NewJobStore(store)
	if err != nil {
		return nil, err
	}
	runnerStore, err := db.NewRunnerStore(store)
	if err != nil {
		return nil, err
	}
	runnerMetadataStore, err := db.NewRunnerMetadataStore(store)
	if err != nil {
		return nil, err
	}

	headscaleServer := headscale.NewHeadscaleServer(&headscale.HeadscaleServerConfig{
		ServerId:      c.Id,
		FrpsDomain:    c.Frps.Domain,
		FrpsProtocol:  c.Frps.Protocol,
		HeadscalePort: c.HeadscalePort,
		ConfigDir:     filepath.Join(configDir, "headscale"),
		Frps:          c.Frps,
	})
	err = headscaleServer.Init()
	if err != nil {
		return nil, err
	}

	gitProviderService := gitproviders.NewGitProviderService(gitproviders.GitProviderServiceConfig{
		ConfigStore: gitProviderConfigStore,
		DetachWorkspaceTemplates: func(ctx context.Context, gitProviderConfigId string) error {
			workspaceTemplates, err := workspaceTemplateStore.List(ctx, &stores.WorkspaceTemplateFilter{
				GitProviderConfigId: &gitProviderConfigId,
			})

			if err != nil {
				return err
			}

			for _, workspaceTemplate := range workspaceTemplates {
				workspaceTemplate.GitProviderConfigId = nil
				err = workspaceTemplateStore.Save(ctx, workspaceTemplate)
				if err != nil {
					return err
				}
			}

			return nil
		},
	})

	jobService := jobs.NewJobService(jobs.JobServiceConfig{
		JobStore: jobStore,
	})

	buildService := builds.NewBuildService(builds.BuildServiceConfig{
		BuildStore: buildStore,
		FindWorkspaceTemplate: func(ctx context.Context, name string) (*models.WorkspaceTemplate, error) {
			return workspaceTemplateStore.Find(ctx, &stores.WorkspaceTemplateFilter{
				Name: &name,
			})
		},
		GetRepositoryContext: func(ctx context.Context, url, branch string) (*gitprovider.GitRepository, error) {
			gitProvider, _, err := gitProviderService.GetGitProviderForUrl(ctx, url)
			if err != nil {
				return nil, err
			}

			repo, err := gitProvider.GetRepositoryContext(gitprovider.GetRepositoryContext{
				Url: url,
			})

			return repo, err
		},
		CreateJob: func(ctx context.Context, workspaceId string, action models.JobAction) error {
			return jobService.Create(ctx, &models.Job{
				ResourceId:   workspaceId,
				ResourceType: models.ResourceTypeBuild,
				Action:       action,
				State:        models.JobStatePending,
			})
		},
		LoggerFactory: logs.NewLoggerFactory(server.GetBuildLogsDir(configDir)),
	})

	prebuildWebhookEndpoint := fmt.Sprintf("%s%s", util.GetFrpcApiUrl(c.Frps.Protocol, c.Id, c.Frps.Domain), constants.WEBHOOK_EVENT_ROUTE)

	workspaceTemplateService := workspacetemplates.NewWorkspaceTemplateService(workspacetemplates.WorkspaceTemplateServiceConfig{
		PrebuildWebhookEndpoint: prebuildWebhookEndpoint,
		ConfigStore:             workspaceTemplateStore,
		FindNewestBuild: func(ctx context.Context, prebuildId string) (*services.BuildDTO, error) {
			return buildService.Find(ctx, &services.BuildFilter{
				StoreFilter: stores.BuildFilter{
					PrebuildIds: &[]string{prebuildId},
					GetNewest:   util.Pointer(true),
				},
			})
		},
		ListSuccessfulBuilds: func(ctx context.Context) ([]*services.BuildDTO, error) {
			return buildService.List(ctx, &services.BuildFilter{
				StateNames: &[]models.ResourceStateName{models.ResourceStateNameRunSuccessful},
			})
		},
		CreateBuild: func(ctx context.Context, workspaceTemplate *models.WorkspaceTemplate, repo *gitprovider.GitRepository, prebuildId string) error {
			createBuildDto := services.CreateBuildDTO{
				WorkspaceTemplateName: workspaceTemplate.Name,
				Branch:                repo.Branch,
				PrebuildId:            &prebuildId,
				EnvVars:               workspaceTemplate.EnvVars,
			}

			_, err := buildService.Create(ctx, createBuildDto)
			return err
		},
		DeleteBuilds: func(ctx context.Context, id, prebuildId *string, force bool) []error {
			var prebuildIds *[]string
			if prebuildId != nil {
				prebuildIds = &[]string{*prebuildId}
			}

			return buildService.Delete(ctx, &services.BuildFilter{
				StoreFilter: stores.BuildFilter{
					Id:          id,
					PrebuildIds: prebuildIds,
				},
			}, force)
		},
		GetRepositoryContext: func(ctx context.Context, url string) (*gitprovider.GitRepository, string, error) {
			gitProvider, gitProviderId, err := gitProviderService.GetGitProviderForUrl(ctx, url)
			if err != nil {
				return nil, "", err
			}

			repo, err := gitProvider.GetRepositoryContext(gitprovider.GetRepositoryContext{
				Url: url,
			})

			return repo, gitProviderId, err
		},
		FindPrebuildWebhook: func(ctx context.Context, gitProviderId string, repo *gitprovider.GitRepository, endpointUrl string) (*string, error) {
			return gitProviderService.GetPrebuildWebhook(ctx, gitProviderId, repo, endpointUrl)
		},
		UnregisterPrebuildWebhook: func(ctx context.Context, gitProviderId string, repo *gitprovider.GitRepository, id string) error {
			return gitProviderService.UnregisterPrebuildWebhook(ctx, gitProviderId, repo, id)
		},
		RegisterPrebuildWebhook: func(ctx context.Context, gitProviderId string, repo *gitprovider.GitRepository, endpointUrl string) (string, error) {
			return gitProviderService.RegisterPrebuildWebhook(ctx, gitProviderId, repo, endpointUrl)
		},
		GetCommitsRange: func(ctx context.Context, repo *gitprovider.GitRepository, initialSha, currentSha string) (int, error) {
			gitProvider, _, err := gitProviderService.GetGitProviderForUrl(ctx, repo.Url)
			if err != nil {
				return 0, err
			}

			return gitProvider.GetCommitsRange(repo, initialSha, currentSha)
		},
	})

	err = workspaceTemplateService.StartRetentionPoller(context.Background())
	if err != nil {
		return nil, err
	}

	var localContainerRegistry server.ILocalContainerRegistry

	if c.BuilderRegistryServer != "local" {
		envVarService := server.GetInstance(nil).EnvironmentVariableService
		envVars, err := envVarService.Map(context.Background())
		if err != nil || envVars.FindContainerRegistry(c.BuilderRegistryServer) == nil {
			log.Errorf("Failed to find container registry credentials for builder registry server %s\n", c.BuilderRegistryServer)
			log.Errorf("Defaulting to local container registry. To use %s as the builder registry, add credentials for the registry server by adding them as environment variables using `daytona env set` and restart the server\n", c.BuilderRegistryServer)
			c.BuilderRegistryServer = "local"
		}
	}

	if c.BuilderRegistryServer == "local" {
		localContainerRegistry = registry.NewLocalContainerRegistry(&registry.LocalContainerRegistryConfig{
			DataPath: filepath.Join(configDir, "registry"),
			Port:     c.LocalBuilderRegistryPort,
			Image:    c.LocalBuilderRegistryImage,
			Logger:   log.StandardLogger().Writer(),
			Frps:     c.Frps,
			ServerId: c.Id,
		})
		c.BuilderRegistryServer = util.GetFrpcRegistryDomain(c.Id, c.Frps.Domain)
	}

	targetConfigService := targetconfigs.NewTargetConfigService(targetconfigs.TargetConfigServiceConfig{
		TargetConfigStore: targetConfigStore,
	})

	apiKeyService := apikeys.NewApiKeyService(apikeys.ApiKeyServiceConfig{
		ApiKeyStore: apiKeyStore,
	})

	headscaleUrl := util.GetFrpcHeadscaleUrl(c.Frps.Protocol, c.Id, c.Frps.Domain)

	targetService := targets.NewTargetService(targets.TargetServiceConfig{
		TargetStore:         targetStore,
		TargetMetadataStore: targetMetadataStore,
		FindTargetConfig: func(ctx context.Context, name string) (*models.TargetConfig, error) {
			return targetConfigService.Find(ctx, name)
		},
		GenerateApiKey: func(ctx context.Context, name string) (string, error) {
			return apiKeyService.Generate(ctx, models.ApiKeyTypeTarget, name)
		},
		RevokeApiKey: func(ctx context.Context, name string) error {
			return apiKeyService.Revoke(ctx, name)
		},
		CreateJob: func(ctx context.Context, targetId string, runnerId string, action models.JobAction) error {
			return jobService.Create(ctx, &models.Job{
				ResourceId:   targetId,
				RunnerId:     &runnerId,
				ResourceType: models.ResourceTypeTarget,
				Action:       action,
				State:        models.JobStatePending,
			})
		},
		ServerApiUrl:     util.GetFrpcApiUrl(c.Frps.Protocol, c.Id, c.Frps.Domain),
		ServerVersion:    version,
		ServerUrl:        headscaleUrl,
		LoggerFactory:    logs.NewLoggerFactory(server.GetTargetLogsDir(configDir)),
		TelemetryService: telemetryService,
	})

	workspaceService := workspaces.NewWorkspaceService(workspaces.WorkspaceServiceConfig{
		WorkspaceStore:         workspaceStore,
		WorkspaceMetadataStore: workspaceMetadataStore,
		FindTarget: func(ctx context.Context, targetId string) (*models.Target, error) {
			t, err := targetService.GetTarget(ctx, &stores.TargetFilter{IdOrName: &targetId}, services.TargetRetrievalParams{})
			if err != nil {
				return nil, err
			}
			return &t.Target, nil
		},
		FindContainerRegistry: func(ctx context.Context, image string, envVars map[string]string) *models.ContainerRegistry {
			return services.EnvironmentVariables(envVars).FindContainerRegistryByImageName(image)
		},
		FindCachedBuild: func(ctx context.Context, w *models.Workspace) (*models.CachedBuild, error) {
			validStates := []models.ResourceStateName{
				models.ResourceStateNameRunSuccessful,
			}

			build, err := buildService.Find(ctx, &services.BuildFilter{
				StateNames: &validStates,
				StoreFilter: stores.BuildFilter{
					RepositoryUrl: &w.Repository.Url,
					Branch:        &w.Repository.Branch,
					EnvVars:       &w.EnvVars,
					BuildConfig:   w.BuildConfig,
					GetNewest:     util.Pointer(true),
				},
			})
			if err != nil {
				return nil, err
			}

			if build.Image == nil || build.User == nil {
				return nil, errors.New("cached build is missing image or user")
			}

			return &models.CachedBuild{
				User:  *build.User,
				Image: *build.Image,
			}, nil
		},
		GenerateApiKey: func(ctx context.Context, name string) (string, error) {
			return apiKeyService.Generate(ctx, models.ApiKeyTypeWorkspace, name)
		},
		RevokeApiKey: func(ctx context.Context, name string) error {
			return apiKeyService.Revoke(ctx, name)
		},
		ListGitProviderConfigs: func(ctx context.Context, repoUrl string) ([]*models.GitProviderConfig, error) {
			return gitProviderService.ListConfigsForUrl(ctx, repoUrl)
		},
		FindGitProviderConfig: func(ctx context.Context, id string) (*models.GitProviderConfig, error) {
			return gitProviderService.GetConfig(ctx, id)
		},
		GetLastCommitSha: func(ctx context.Context, repo *gitprovider.GitRepository) (string, error) {
			return gitProviderService.GetLastCommitSha(ctx, repo)
		},
		CreateJob: func(ctx context.Context, workspaceId string, runnerId string, action models.JobAction) error {
			return jobService.Create(ctx, &models.Job{
				ResourceId:   workspaceId,
				RunnerId:     &runnerId,
				ResourceType: models.ResourceTypeWorkspace,
				Action:       action,
				State:        models.JobStatePending,
			})
		},
		TrackTelemetryEvent:   telemetryService.TrackServerEvent,
		ServerApiUrl:          util.GetFrpcApiUrl(c.Frps.Protocol, c.Id, c.Frps.Domain),
		ServerVersion:         version,
		ServerUrl:             headscaleUrl,
		DefaultWorkspaceImage: c.DefaultWorkspaceImage,
		DefaultWorkspaceUser:  c.DefaultWorkspaceUser,
		LoggerFactory:         logs.NewLoggerFactory(server.GetWorkspaceLogsDir(configDir)),
	})

	envVarService := env.NewEnvironmentVariableService(env.EnvironmentVariableServiceConfig{
		EnvironmentVariableStore: envVarStore,
	})

	runnerService := runners.NewRunnerService(runners.RunnerServiceConfig{
		RunnerStore:         runnerStore,
		RunnerMetadataStore: runnerMetadataStore,
		LoggerFactory:       logs.NewLoggerFactory(server.GetRunnerLogsDir(configDir)),
		CreateJob: func(ctx context.Context, runnerId string, action models.JobAction, metadata string) error {
			return jobService.Create(ctx, &models.Job{
				ResourceId:   runnerId,
				RunnerId:     &runnerId,
				ResourceType: models.ResourceTypeRunner,
				Action:       action,
				State:        models.JobStatePending,
				Metadata:     &metadata,
			})
		},
		ListJobsForRunner: func(ctx context.Context, runnerId string) ([]*models.Job, error) {
			return jobService.List(ctx, &stores.JobFilter{
				RunnerIdOrIsNil: &runnerId,
				States:          &[]models.JobState{models.JobStatePending},
			})
		},
		UpdateJobState: func(ctx context.Context, jobId string, updateJobStateDto services.UpdateJobStateDTO) error {
			return jobService.SetState(ctx, jobId, updateJobStateDto)
		},
		GenerateApiKey: func(ctx context.Context, name string) (string, error) {
			return apiKeyService.Generate(ctx, models.ApiKeyTypeRunner, name)
		},
		RevokeApiKey: apiKeyService.Revoke,
		UnsetDefaultTarget: func(ctx context.Context, runnerId string) error {
			targets, err := targetService.ListTargets(ctx, nil, services.TargetRetrievalParams{})
			if err != nil {
				return err
			}

			for _, t := range targets {
				if t.TargetConfig.ProviderInfo.RunnerId == runnerId && t.IsDefault {
					t.IsDefault = false
					err = targetService.SaveTarget(ctx, &t.Target)
					if err != nil {
						return err
					}
					break
				}
			}
			return nil
		},
		TrackTelemetryEvent: telemetryService.TrackServerEvent,
	})

	s := server.GetInstance(&server.ServerInstanceConfig{
		Config:                     *c,
		Version:                    version,
		TailscaleServer:            headscaleServer,
		TargetConfigService:        targetConfigService,
		BuildService:               buildService,
		WorkspaceTemplateService:   workspaceTemplateService,
		WorkspaceService:           workspaceService,
		LocalContainerRegistry:     localContainerRegistry,
		ApiKeyService:              apiKeyService,
		TargetService:              targetService,
		GitProviderService:         gitProviderService,
		EnvironmentVariableService: envVarService,
		JobService:                 jobService,
		RunnerService:              runnerService,
		TelemetryService:           telemetryService,
	})

	return s, s.Initialize()
}

func getDbPath() (string, error) {
	configDir, err := config.GetConfigDir()
	if err != nil {
		return "", err
	}

	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "db"), nil
}
