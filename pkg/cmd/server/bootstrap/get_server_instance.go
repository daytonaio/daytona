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
	"github.com/daytonaio/daytona/pkg/build"
	"github.com/daytonaio/daytona/pkg/db"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/provider/manager"
	"github.com/daytonaio/daytona/pkg/provisioner"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/server/apikeys"
	"github.com/daytonaio/daytona/pkg/server/builds"
	"github.com/daytonaio/daytona/pkg/server/containerregistries"
	"github.com/daytonaio/daytona/pkg/server/gitproviders"
	"github.com/daytonaio/daytona/pkg/server/headscale"
	"github.com/daytonaio/daytona/pkg/server/jobs"
	"github.com/daytonaio/daytona/pkg/server/profiledata"
	"github.com/daytonaio/daytona/pkg/server/registry"
	"github.com/daytonaio/daytona/pkg/server/targetconfigs"
	"github.com/daytonaio/daytona/pkg/server/targets"
	"github.com/daytonaio/daytona/pkg/server/workspaceconfigs"
	"github.com/daytonaio/daytona/pkg/server/workspaces"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/daytonaio/daytona/pkg/telemetry"

	log "github.com/sirupsen/logrus"
)

func GetInstance(c *server.Config, configDir string, version string, telemetryService telemetry.TelemetryService) (*server.Server, error) {
	targetLogsDir, err := server.GetTargetLogsDir(configDir)
	if err != nil {
		return nil, err
	}
	buildLogsDir, err := build.GetBuildLogsDir()
	if err != nil {
		return nil, err
	}
	loggerFactory := logs.NewLoggerFactory(&targetLogsDir, &buildLogsDir)

	dbPath, err := getDbPath()
	if err != nil {
		return nil, err
	}

	dbConnection := db.GetSQLiteConnection(dbPath)

	apiKeyStore, err := db.NewApiKeyStore(dbConnection)
	if err != nil {
		return nil, err
	}
	containerRegistryStore, err := db.NewContainerRegistryStore(dbConnection)
	if err != nil {
		return nil, err
	}
	buildStore, err := db.NewBuildStore(dbConnection)
	if err != nil {
		return nil, err
	}
	workspaceConfigStore, err := db.NewWorkspaceConfigStore(dbConnection)
	if err != nil {
		return nil, err
	}
	gitProviderConfigStore, err := db.NewGitProviderConfigStore(dbConnection)
	if err != nil {
		return nil, err
	}
	targetConfigStore, err := db.NewTargetConfigStore(dbConnection)
	if err != nil {
		return nil, err
	}
	targetStore, err := db.NewTargetStore(dbConnection)
	if err != nil {
		return nil, err
	}
	targetMetadataStore, err := db.NewTargetMetadataStore(dbConnection)
	if err != nil {
		return nil, err
	}
	profileDataStore, err := db.NewProfileDataStore(dbConnection)
	if err != nil {
		return nil, err
	}
	workspaceStore, err := db.NewWorkspaceStore(dbConnection)
	if err != nil {
		return nil, err
	}
	workspaceMetadataStore, err := db.NewWorkspaceMetadataStore(dbConnection)
	if err != nil {
		return nil, err
	}
	jobStore, err := db.NewJobStore(dbConnection)
	if err != nil {
		return nil, err
	}

	providerManager := manager.GetProviderManager(nil)

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

	containerRegistryService := containerregistries.NewContainerRegistryService(containerregistries.ContainerRegistryServiceConfig{
		Store: containerRegistryStore,
	})

	gitProviderService := gitproviders.NewGitProviderService(gitproviders.GitProviderServiceConfig{
		ConfigStore: gitProviderConfigStore,
		DetachWorkspaceConfigs: func(ctx context.Context, gitProviderConfigId string) error {
			workspaceConfigs, err := workspaceConfigStore.List(&stores.WorkspaceConfigFilter{
				GitProviderConfigId: &gitProviderConfigId,
			})

			if err != nil {
				return err
			}

			for _, workspaceConfig := range workspaceConfigs {
				workspaceConfig.GitProviderConfigId = nil
				err = workspaceConfigStore.Save(workspaceConfig)
				if err != nil {
					return err
				}
			}

			return nil
		},
	})

	buildService := builds.NewBuildService(builds.BuildServiceConfig{
		BuildStore: buildStore,
		FindWorkspaceConfig: func(ctx context.Context, name string) (*models.WorkspaceConfig, error) {
			return workspaceConfigStore.Find(&stores.WorkspaceConfigFilter{
				Name: &name,
			})
		},
		GetRepositoryContext: func(ctx context.Context, url, branch string) (*gitprovider.GitRepository, error) {
			gitProvider, _, err := gitProviderService.GetGitProviderForUrl(url)
			if err != nil {
				return nil, err
			}

			repo, err := gitProvider.GetRepositoryContext(gitprovider.GetRepositoryContext{
				Url: url,
			})

			return repo, err
		},
		LoggerFactory: loggerFactory,
	})

	prebuildWebhookEndpoint := fmt.Sprintf("%s%s", util.GetFrpcApiUrl(c.Frps.Protocol, c.Id, c.Frps.Domain), constants.WEBHOOK_EVENT_ROUTE)

	workspaceConfigService := workspaceconfigs.NewWorkspaceConfigService(workspaceconfigs.WorkspaceConfigServiceConfig{
		PrebuildWebhookEndpoint: prebuildWebhookEndpoint,
		ConfigStore:             workspaceConfigStore,
		FindNewestBuild: func(ctx context.Context, prebuildId string) (*models.Build, error) {
			return buildService.Find(&stores.BuildFilter{
				PrebuildIds: &[]string{prebuildId},
				GetNewest:   util.Pointer(true),
			})
		},
		ListPublishedBuilds: func(ctx context.Context) ([]*models.Build, error) {
			return buildService.List(&stores.BuildFilter{
				States: &[]models.BuildState{models.BuildStatePublished},
			})
		},
		CreateBuild: func(ctx context.Context, workspaceConfig *models.WorkspaceConfig, repo *gitprovider.GitRepository, prebuildId string) error {
			createBuildDto := services.CreateBuildDTO{
				WorkspaceConfigName: workspaceConfig.Name,
				Branch:              repo.Branch,
				PrebuildId:          &prebuildId,
				EnvVars:             workspaceConfig.EnvVars,
			}

			_, err := buildService.Create(createBuildDto)
			return err
		},
		DeleteBuilds: func(ctx context.Context, id, prebuildId *string, force bool) []error {
			var prebuildIds *[]string
			if prebuildId != nil {
				prebuildIds = &[]string{*prebuildId}
			}

			return buildService.MarkForDeletion(&stores.BuildFilter{
				Id:          id,
				PrebuildIds: prebuildIds,
			}, force)
		},
		GetRepositoryContext: func(ctx context.Context, url string) (*gitprovider.GitRepository, string, error) {
			gitProvider, gitProviderId, err := gitProviderService.GetGitProviderForUrl(url)
			if err != nil {
				return nil, "", err
			}

			repo, err := gitProvider.GetRepositoryContext(gitprovider.GetRepositoryContext{
				Url: url,
			})

			return repo, gitProviderId, err
		},
		FindPrebuildWebhook: func(ctx context.Context, gitProviderId string, repo *gitprovider.GitRepository, endpointUrl string) (*string, error) {
			return gitProviderService.GetPrebuildWebhook(gitProviderId, repo, endpointUrl)
		},
		UnregisterPrebuildWebhook: func(ctx context.Context, gitProviderId string, repo *gitprovider.GitRepository, id string) error {
			return gitProviderService.UnregisterPrebuildWebhook(gitProviderId, repo, id)
		},
		RegisterPrebuildWebhook: func(ctx context.Context, gitProviderId string, repo *gitprovider.GitRepository, endpointUrl string) (string, error) {
			return gitProviderService.RegisterPrebuildWebhook(gitProviderId, repo, endpointUrl)
		},
		GetCommitsRange: func(ctx context.Context, repo *gitprovider.GitRepository, initialSha, currentSha string) (int, error) {
			gitProvider, _, err := gitProviderService.GetGitProviderForUrl(repo.Url)
			if err != nil {
				return 0, err
			}

			return gitProvider.GetCommitsRange(repo, initialSha, currentSha)
		},
	})

	err = workspaceConfigService.StartRetentionPoller()
	if err != nil {
		return nil, err
	}

	var localContainerRegistry server.ILocalContainerRegistry

	if c.BuilderRegistryServer != "local" {
		_, err := containerRegistryService.Find(c.BuilderRegistryServer)
		if err != nil {
			log.Errorf("Failed to find container registry credentials for builder registry server %s\n", c.BuilderRegistryServer)
			log.Errorf("Defaulting to local container registry. To use %s as the builder registry, add credentials for the registry server with 'daytona container-registry set' and restart the server\n", c.BuilderRegistryServer)
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

	jobService := jobs.NewJobService(jobs.JobServiceConfig{
		JobStore: jobStore,
	})

	provisioner := provisioner.NewProvisioner(provisioner.ProvisionerConfig{
		ProviderManager: providerManager,
	})

	targetService := targets.NewTargetService(targets.TargetServiceConfig{
		TargetStore:         targetStore,
		TargetMetadataStore: targetMetadataStore,
		FindTargetConfig: func(ctx context.Context, name string) (*models.TargetConfig, error) {
			return targetConfigService.Find(&stores.TargetConfigFilter{Name: &name})
		},
		GenerateApiKey: func(ctx context.Context, name string) (string, error) {
			return apiKeyService.Generate(models.ApiKeyTypeTarget, name)
		},
		RevokeApiKey: func(ctx context.Context, name string) error {
			return apiKeyService.Revoke(name)
		},
		CreateJob: func(ctx context.Context, targetId string, action models.JobAction) error {
			return jobService.Save(&models.Job{
				ResourceId:   targetId,
				ResourceType: models.ResourceTypeTarget,
				Action:       action,
				State:        models.JobStatePending,
			})
		},
		ServerApiUrl:     util.GetFrpcApiUrl(c.Frps.Protocol, c.Id, c.Frps.Domain),
		ServerVersion:    version,
		ServerUrl:        headscaleUrl,
		Provisioner:      provisioner,
		LoggerFactory:    loggerFactory,
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
		FindContainerRegistry: func(ctx context.Context, image string) (*models.ContainerRegistry, error) {
			return containerRegistryService.FindByImageName(image)
		},
		FindCachedBuild: func(ctx context.Context, w *models.Workspace) (*models.CachedBuild, error) {
			validStates := &[]models.BuildState{
				models.BuildStatePublished,
			}

			build, err := buildService.Find(&stores.BuildFilter{
				States:        validStates,
				RepositoryUrl: &w.Repository.Url,
				Branch:        &w.Repository.Branch,
				EnvVars:       &w.EnvVars,
				BuildConfig:   w.BuildConfig,
				GetNewest:     util.Pointer(true),
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
			return apiKeyService.Generate(models.ApiKeyTypeWorkspace, name)
		},
		RevokeApiKey: func(ctx context.Context, name string) error {
			return apiKeyService.Revoke(name)
		},
		ListGitProviderConfigs: func(ctx context.Context, repoUrl string) ([]*models.GitProviderConfig, error) {
			return gitProviderService.ListConfigsForUrl(repoUrl)
		},
		FindGitProviderConfig: func(ctx context.Context, id string) (*models.GitProviderConfig, error) {
			return gitProviderService.GetConfig(id)
		},
		GetLastCommitSha: func(ctx context.Context, repo *gitprovider.GitRepository) (string, error) {
			return gitProviderService.GetLastCommitSha(repo)
		},
		CreateJob: func(ctx context.Context, workspaceId string, action models.JobAction) error {
			return jobService.Save(&models.Job{
				ResourceId:   workspaceId,
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
		Provisioner:           provisioner,
		LoggerFactory:         loggerFactory,
	})

	profileDataService := profiledata.NewProfileDataService(profiledata.ProfileDataServiceConfig{
		ProfileDataStore: profileDataStore,
	})

	s := server.GetInstance(&server.ServerInstanceConfig{
		Config:                   *c,
		Version:                  version,
		TailscaleServer:          headscaleServer,
		TargetConfigService:      targetConfigService,
		ContainerRegistryService: containerRegistryService,
		BuildService:             buildService,
		WorkspaceConfigService:   workspaceConfigService,
		WorkspaceService:         workspaceService,
		LocalContainerRegistry:   localContainerRegistry,
		ApiKeyService:            apiKeyService,
		TargetService:            targetService,
		GitProviderService:       gitProviderService,
		ProviderManager:          providerManager,
		ProfileDataService:       profileDataService,
		JobService:               jobService,
		TelemetryService:         telemetryService,
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
