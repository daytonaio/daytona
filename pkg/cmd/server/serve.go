// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
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
	"github.com/daytonaio/daytona/pkg/apikey"
	"github.com/daytonaio/daytona/pkg/build"
	"github.com/daytonaio/daytona/pkg/containerregistry"
	"github.com/daytonaio/daytona/pkg/db"
	"github.com/daytonaio/daytona/pkg/logs"
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
	"github.com/daytonaio/daytona/pkg/server/projectconfig"
	"github.com/daytonaio/daytona/pkg/server/registry"
	"github.com/daytonaio/daytona/pkg/server/targetconfigs"
	"github.com/daytonaio/daytona/pkg/server/targets"
	"github.com/daytonaio/daytona/pkg/telemetry"
	"github.com/daytonaio/daytona/pkg/views"
	started_view "github.com/daytonaio/daytona/pkg/views/server/started"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var DaemonServeCmd = &cobra.Command{
	Use:    "daemon-serve",
	Short:  "Used by the daemon to start the Daytona Server",
	Args:   cobra.NoArgs,
	Hidden: true,
	RunE:   ServeCmd.RunE,
}

var ServeCmd = &cobra.Command{
	Use:     "serve",
	Short:   "Run the server process in the current terminal session",
	GroupID: util.SERVER_GROUP,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if os.Getenv("USER") == "root" {
			views.RenderInfoMessageBold("Running the server as root is not recommended because\nDaytona will not be able to remap project directory ownership.\nPlease run the server as a non-root user.")
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
			log.Errorf("Failed to start local container registry: %v\nBuilds may not work properly.\nRestart the server to restart the registry.", err)
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
	projectConfigStore, err := db.NewProjectConfigStore(dbConnection)
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

	buildService := builds.NewBuildService(builds.BuildServiceConfig{
		BuildStore:    buildStore,
		LoggerFactory: loggerFactory,
	})

	gitProviderService := gitproviders.NewGitProviderService(gitproviders.GitProviderServiceConfig{
		ConfigStore:        gitProviderConfigStore,
		ProjectConfigStore: projectConfigStore,
	})

	prebuildWebhookEndpoint := fmt.Sprintf("%s%s", util.GetFrpcApiUrl(c.Frps.Protocol, c.Id, c.Frps.Domain), constants.WEBHOOK_EVENT_ROUTE)

	projectConfigService := projectconfig.NewProjectConfigService(projectconfig.ProjectConfigServiceConfig{
		PrebuildWebhookEndpoint: prebuildWebhookEndpoint,
		ConfigStore:             projectConfigStore,
		BuildService:            buildService,
		GitProviderService:      gitProviderService,
	})

	err = projectConfigService.StartRetentionPoller()
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
		cr, err := containerRegistryService.FindByImageName(c.LocalBuilderRegistryImage)
		if err != nil && !containerregistry.IsContainerRegistryNotFound(err) {
			return nil, err
		}

		localContainerRegistry = registry.NewLocalContainerRegistry(&registry.LocalContainerRegistryConfig{
			DataPath:          filepath.Join(configDir, "registry"),
			Port:              c.LocalBuilderRegistryPort,
			Image:             c.LocalBuilderRegistryImage,
			ContainerRegistry: cr,
			Logger:            log.StandardLogger().Writer(),
			Frps:              c.Frps,
			ServerId:          c.Id,
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
		CreateProviderNetworkKey: func(providerName string) (string, error) {
			return headscaleServer.CreateAuthKey()
		},
		ServerPort:          c.HeadscalePort,
		ApiPort:             c.ApiPort,
		TargetConfigService: targetConfigService,
	})

	provisioner := provisioner.NewProvisioner(provisioner.ProvisionerConfig{
		ProviderManager: providerManager,
	})

	targetService := targets.NewTargetService(targets.TargetServiceConfig{
		TargetStore:              targetStore,
		TargetConfigStore:        targetConfigStore,
		ApiKeyService:            apiKeyService,
		GitProviderService:       gitProviderService,
		ContainerRegistryService: containerRegistryService,
		BuilderImage:             c.BuilderImage,
		BuildService:             buildService,
		ProjectConfigService:     projectConfigService,
		ServerApiUrl:             util.GetFrpcApiUrl(c.Frps.Protocol, c.Id, c.Frps.Domain),
		ServerVersion:            version,
		ServerUrl:                headscaleUrl,
		DefaultProjectImage:      c.DefaultProjectImage,
		DefaultProjectUser:       c.DefaultProjectUser,
		Provisioner:              provisioner,
		LoggerFactory:            loggerFactory,
		TelemetryService:         telemetryService,
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
		ProjectConfigService:     projectConfigService,
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

	var buildImageCr *containerregistry.ContainerRegistry

	if c.BuilderRegistryServer != "local" {
		buildImageCr, err = containerRegistryService.Find(c.BuilderRegistryServer)
		if err != nil {
			buildImageCr = &containerregistry.ContainerRegistry{
				Server: c.BuilderRegistryServer,
			}
		}
	}

	cr, err := containerRegistryService.FindByImageName(c.BuilderImage)
	if err != nil && !containerregistry.IsContainerRegistryNotFound(err) {
		return nil, err
	}

	configDir, err := config.GetConfigDir()
	if err != nil {
		return nil, err
	}

	builderFactory := build.NewBuilderFactory(build.BuilderFactoryConfig{
		Image:                       c.BuilderImage,
		ContainerRegistry:           cr,
		BuildImageContainerRegistry: buildImageCr,
		BuildStore:                  buildStore,
		BuildImageNamespace:         buildImageNamespace,
		LoggerFactory:               loggerFactory,
		DefaultProjectImage:         c.DefaultProjectImage,
		DefaultProjectUser:          c.DefaultProjectUser,
	})

	return build.NewBuildRunner(build.BuildRunnerInstanceConfig{
		Interval:          buildRunnerConfig.Interval,
		Scheduler:         build.NewCronScheduler(),
		BuildRunnerId:     buildRunnerConfig.Id,
		ContainerRegistry: buildImageCr,
		TelemetryEnabled:  buildRunnerConfig.TelemetryEnabled,
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

	apiKey, err := server.ApiKeyService.Generate(apikey.ApiKeyTypeClient, "default")
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
