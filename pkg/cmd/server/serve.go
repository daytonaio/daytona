// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal"
	"github.com/daytonaio/daytona/internal/constants"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/api"
	"github.com/daytonaio/daytona/pkg/build"
	"github.com/daytonaio/daytona/pkg/db"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/posthogservice"
	"github.com/daytonaio/daytona/pkg/provider/manager"
	"github.com/daytonaio/daytona/pkg/provisioner"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/server/apikeys"
	"github.com/daytonaio/daytona/pkg/server/builds"
	"github.com/daytonaio/daytona/pkg/server/containerregistries"
	"github.com/daytonaio/daytona/pkg/server/gitproviders"
	"github.com/daytonaio/daytona/pkg/server/headscale"
	"github.com/daytonaio/daytona/pkg/server/profiledata"
	"github.com/daytonaio/daytona/pkg/server/registry"
	"github.com/daytonaio/daytona/pkg/server/targetconfigs"
	"github.com/daytonaio/daytona/pkg/server/targets"
	"github.com/daytonaio/daytona/pkg/server/workspaceconfigs"
	"github.com/daytonaio/daytona/pkg/server/workspaces"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/daytonaio/daytona/pkg/telemetry"
	"github.com/daytonaio/daytona/pkg/views"
	started_view "github.com/daytonaio/daytona/pkg/views/server/started"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var ServeCmd = &cobra.Command{
	Use:     "serve",
	Short:   "Run the server process in the current terminal session",
	GroupID: util.SERVER_GROUP,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if os.Getenv("USER") == "root" {
			views.RenderInfoMessageBold("Running the server as root is not recommended because\nDaytona will not be able to remap workspace directory ownership.\nPlease run the server as a non-root user.")
		}

		if log.GetLevel() < log.InfoLevel {
			//	for now, force the log level to info when running the server
			log.SetLevel(log.InfoLevel)
		}

		configDir, err := server.GetConfigDir()
		if err != nil {
			return err
		}

		c, err := server.GetConfig()
		if err != nil {
			return err
		}

		telemetryService := posthogservice.NewTelemetryService(posthogservice.PosthogServiceConfig{
			ApiKey:   internal.PosthogApiKey,
			Endpoint: internal.PosthogEndpoint,
			Version:  internal.Version,
		})

		apiServer := api.NewApiServer(api.ApiServerConfig{
			ApiPort:          int(c.ApiPort),
			TelemetryService: telemetryService,
			Version:          internal.Version,
			ServerId:         c.Id,
			Frps:             c.Frps,
		})

		server, err := GetInstance(c, configDir, internal.Version, telemetryService)
		if err != nil {
			return err
		}

		buildRunnerConfig, err := build.GetConfig()
		if err != nil {
			return err
		}

		buildRunner, err := GetBuildRunner(c, buildRunnerConfig, telemetryService)
		if err != nil {
			return err
		}

		err = buildRunner.Start()
		if err != nil {
			return err
		}

		apiServerErrChan := make(chan error)

		go func() {
			log.Infof("Starting api server on port %d", c.ApiPort)
			apiServerErrChan <- apiServer.Start()
		}()

		headscaleServerStartedChan := make(chan struct{})
		headscaleServerErrChan := make(chan error)

		go func() {
			log.Info("Starting headscale server...")
			err := server.TailscaleServer.Start(headscaleServerErrChan)
			if err != nil {
				headscaleServerErrChan <- err
				return
			}
			headscaleServerStartedChan <- struct{}{}
		}()

		localContainerRegistryErrChan := make(chan error)

		go func() {
			if server.LocalContainerRegistry != nil {
				log.Info("Starting local container registry...")
				localContainerRegistryErrChan <- server.LocalContainerRegistry.Start()
			} else {
				localContainerRegistryErrChan <- registry.RemoveRegistryContainer()
			}
		}()

		select {
		case <-headscaleServerStartedChan:
			log.Info("Headscale server started")
			go func() {
				headscaleServerErrChan <- server.TailscaleServer.Connect()
			}()
		case err := <-headscaleServerErrChan:
			return err
		}

		err = server.Start()
		if err != nil {
			return err
		}

		err = waitForApiServerToStart(apiServer)
		if err != nil {
			return err
		}

		err = <-localContainerRegistryErrChan
		if err != nil {
			return err
		}

		printServerStartedMessage(c, false)

		err = ensureDefaultProfile(server, c.ApiPort)
		if err != nil {
			return err
		}

		interruptChannel := make(chan os.Signal, 1)
		signal.Notify(interruptChannel, os.Interrupt)

		select {
		case err := <-apiServerErrChan:
			return err
		case err := <-headscaleServerErrChan:
			return err
		case <-interruptChannel:
			log.Info("Shutting down")

			return server.TailscaleServer.Stop()
		}
	},
}

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
	profileDataStore, err := db.NewProfileDataStore(dbConnection)
	if err != nil {
		return nil, err
	}
	workspaceStore, err := db.NewWorkspaceStore(dbConnection)
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

	providerManager := manager.NewProviderManager(manager.ProviderManagerConfig{
		LogsDir:            targetLogsDir,
		ApiUrl:             util.GetFrpcApiUrl(c.Frps.Protocol, c.Id, c.Frps.Domain),
		DaytonaDownloadUrl: getDaytonaScriptUrl(c),
		ServerUrl:          headscaleUrl,
		ServerVersion:      version,
		RegistryUrl:        c.RegistryUrl,
		BaseDir:            c.ProvidersDir,
		CreateProviderNetworkKey: func(ctx context.Context, providerName string) (string, error) {
			return headscaleServer.CreateAuthKey()
		},
		ServerPort: c.HeadscalePort,
		ApiPort:    c.ApiPort,
		GetTargetConfigMap: func(ctx context.Context) (map[string]*models.TargetConfig, error) {
			return targetConfigService.Map()
		},
		CreateTargetConfig: func(ctx context.Context, targetConfig *models.TargetConfig) error {
			return targetConfigService.Save(targetConfig)
		},
	})

	provisioner := provisioner.NewProvisioner(provisioner.ProvisionerConfig{
		ProviderManager: providerManager,
	})

	targetService := targets.NewTargetService(targets.TargetServiceConfig{
		TargetStore: targetStore,
		FindTargetConfig: func(ctx context.Context, name string) (*models.TargetConfig, error) {
			return targetConfigService.Find(&stores.TargetConfigFilter{Name: &name})
		},
		GenerateApiKey: func(ctx context.Context, name string) (string, error) {
			return apiKeyService.Generate(models.ApiKeyTypeTarget, name)
		},
		RevokeApiKey: func(ctx context.Context, name string) error {
			return apiKeyService.Revoke(name)
		},
		ServerApiUrl:     util.GetFrpcApiUrl(c.Frps.Protocol, c.Id, c.Frps.Domain),
		ServerVersion:    version,
		ServerUrl:        headscaleUrl,
		Provisioner:      provisioner,
		LoggerFactory:    loggerFactory,
		TelemetryService: telemetryService,
	})

	workspaceService := workspaces.NewWorkspaceService(workspaces.WorkspaceServiceConfig{
		WorkspaceStore: workspaceStore,
		FindTarget: func(ctx context.Context, targetId string) (*models.Target, error) {
			t, err := targetService.GetTarget(ctx, &stores.TargetFilter{IdOrName: &targetId}, false)
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
		TelemetryService:         telemetryService,
	})

	return s, s.Initialize()
}

func GetBuildRunner(c *server.Config, buildRunnerConfig *build.Config, telemetryService telemetry.TelemetryService) (*build.BuildRunner, error) {
	logsDir, err := build.GetBuildLogsDir()
	if err != nil {
		return nil, err
	}
	loggerFactory := logs.NewLoggerFactory(nil, &logsDir)

	dbPath, err := getDbPath()
	if err != nil {
		return nil, err
	}

	dbConnection := db.GetSQLiteConnection(dbPath)

	gitProviderConfigStore, err := db.NewGitProviderConfigStore(dbConnection)
	if err != nil {
		return nil, err
	}

	gitProviderService := gitproviders.NewGitProviderService(gitproviders.GitProviderServiceConfig{
		ConfigStore: gitProviderConfigStore,
	})

	buildStore, err := db.NewBuildStore(dbConnection)
	if err != nil {
		return nil, err
	}

	buildImageNamespace := c.BuildImageNamespace
	if buildImageNamespace != "" {
		buildImageNamespace = fmt.Sprintf("/%s", buildImageNamespace)
	}
	buildImageNamespace = strings.TrimSuffix(buildImageNamespace, "/")

	containerRegistryStore, err := db.NewContainerRegistryStore(dbConnection)
	if err != nil {
		return nil, err
	}

	containerRegistryService := containerregistries.NewContainerRegistryService(containerregistries.ContainerRegistryServiceConfig{
		Store: containerRegistryStore,
	})

	var builderRegistry *models.ContainerRegistry

	if c.BuilderRegistryServer != "local" {
		builderRegistry, err = containerRegistryService.Find(c.BuilderRegistryServer)
		if err != nil {
			builderRegistry = &models.ContainerRegistry{
				Server: c.BuilderRegistryServer,
			}
		}
	}

	configDir, err := config.GetConfigDir()
	if err != nil {
		return nil, err
	}

	builderFactory := build.NewBuilderFactory(build.BuilderFactoryConfig{
		Image:                 c.BuilderImage,
		ContainerRegistry:     builderRegistry,
		BuildStore:            buildStore,
		BuildImageNamespace:   buildImageNamespace,
		LoggerFactory:         loggerFactory,
		DefaultWorkspaceImage: c.DefaultWorkspaceImage,
		DefaultWorkspaceUser:  c.DefaultWorkspaceUser,
	})

	return build.NewBuildRunner(build.BuildRunnerInstanceConfig{
		Interval:          buildRunnerConfig.Interval,
		Scheduler:         build.NewCronScheduler(),
		BuildRunnerId:     buildRunnerConfig.Id,
		ContainerRegistry: builderRegistry,
		GitProviderStore:  gitProviderService,
		BuildStore:        buildStore,
		BuilderFactory:    builderFactory,
		LoggerFactory:     loggerFactory,
		BasePath:          filepath.Join(configDir, "builds"),
		TelemetryService:  telemetryService,
	}), nil
}

func waitForApiServerToStart(apiServer *api.ApiServer) error {
	var err error
	for i := 0; i < 30; i++ {
		time.Sleep(1 * time.Second)
		err = apiServer.HealthCheck()
		if err != nil {
			continue
		}

		return nil
	}

	return err
}

func getDaytonaScriptUrl(config *server.Config) string {
	url, _ := url.JoinPath(util.GetFrpcApiUrl(config.Frps.Protocol, config.Id, config.Frps.Domain), "binary", "script")
	return url
}

func printServerStartedMessage(c *server.Config, runAsDaemon bool) {
	started_view.Render(c.ApiPort, util.GetFrpcApiUrl(c.Frps.Protocol, c.Id, c.Frps.Domain), runAsDaemon)
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

func ensureDefaultProfile(server *server.Server, apiPort uint32) error {
	existingConfig, err := config.GetConfig()
	if err != nil {
		return err
	}

	if existingConfig == nil {
		return errors.New("config does not exist")
	}

	for _, profile := range existingConfig.Profiles {
		if profile.Id == "default" {
			return nil
		}
	}

	apiKey, err := server.ApiKeyService.Generate(models.ApiKeyTypeClient, "default")
	if err != nil {
		return err
	}

	return existingConfig.AddProfile(config.Profile{
		Id:   "default",
		Name: "default",
		Api: config.ServerApi{
			Url: fmt.Sprintf("http://localhost:%d", apiPort),
			Key: apiKey,
		},
	})
}
