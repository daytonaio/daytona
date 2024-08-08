// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/api"
	"github.com/daytonaio/daytona/pkg/apikey"
	"github.com/daytonaio/daytona/pkg/build"
	"github.com/daytonaio/daytona/pkg/db"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/posthogservice"
	"github.com/daytonaio/daytona/pkg/provider/manager"
	"github.com/daytonaio/daytona/pkg/provisioner"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/server/apikeys"
	"github.com/daytonaio/daytona/pkg/server/containerregistries"
	"github.com/daytonaio/daytona/pkg/server/gitproviders"
	"github.com/daytonaio/daytona/pkg/server/headscale"
	"github.com/daytonaio/daytona/pkg/server/profiledata"
	"github.com/daytonaio/daytona/pkg/server/projectconfig"
	"github.com/daytonaio/daytona/pkg/server/providertargets"
	"github.com/daytonaio/daytona/pkg/server/registry"
	"github.com/daytonaio/daytona/pkg/server/workspaces"
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
	Run: func(cmd *cobra.Command, args []string) {
		if os.Getenv("USER") == "root" {
			views.RenderInfoMessageBold("Running the server as root is not recommended because\nDaytona will not be able to remap project directory ownership.\nPlease run the server as a non-root user.")
		}

		if log.GetLevel() < log.InfoLevel {
			//	for now, force the log level to info when running the server
			log.SetLevel(log.InfoLevel)
		}

		configDir, err := server.GetConfigDir()
		if err != nil {
			log.Fatal(err)
		}

		c, err := server.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		telemetryService := posthogservice.NewTelemetryService(posthogservice.PosthogServiceConfig{
			ApiKey:   internal.PosthogApiKey,
			Endpoint: internal.PosthogEndpoint,
		})
		go func() {
			interruptChannel := make(chan os.Signal, 1)
			signal.Notify(interruptChannel, os.Interrupt)

			for range interruptChannel {
				log.Info("Shutting down")
				telemetryService.Close()
			}
		}()

		apiServer := api.NewApiServer(api.ApiServerConfig{
			ApiPort:          int(c.ApiPort),
			TelemetryService: telemetryService,
		})

		logsDir, err := server.GetWorkspaceLogsDir()
		if err != nil {
			log.Fatal(err)
		}
		loggerFactory := logs.NewLoggerFactory(logsDir)

		dbPath, err := getDbPath()
		if err != nil {
			log.Fatal(err)
		}

		dbConnection := db.GetSQLiteConnection(dbPath)
		apiKeyStore, err := db.NewApiKeyStore(dbConnection)
		if err != nil {
			log.Fatal(err)
		}
		containerRegistryStore, err := db.NewContainerRegistryStore(dbConnection)
		if err != nil {
			log.Fatal(err)
		}
		projectConfigStore, err := db.NewProjectConfigStore(dbConnection)
		if err != nil {
			log.Fatal(err)
		}
		gitProviderConfigStore, err := db.NewGitProviderConfigStore(dbConnection)
		if err != nil {
			log.Fatal(err)
		}
		providerTargetStore, err := db.NewProviderTargetStore(dbConnection)
		if err != nil {
			log.Fatal(err)
		}
		workspaceStore, err := db.NewWorkspaceStore(dbConnection)
		if err != nil {
			log.Fatal(err)
		}
		profileDataStore, err := db.NewProfileDataStore(dbConnection)
		if err != nil {
			log.Fatal(err)
		}
		buildStore, err := db.NewBuildStore(dbConnection)
		if err != nil {
			log.Fatal(err)
		}

		headscaleServer := headscale.NewHeadscaleServer(&headscale.HeadscaleServerConfig{
			ServerId:      c.Id,
			FrpsDomain:    c.Frps.Domain,
			FrpsProtocol:  c.Frps.Protocol,
			HeadscalePort: c.HeadscalePort,
			ConfigDir:     filepath.Join(configDir, "headscale"),
		})
		err = headscaleServer.Init()
		if err != nil {
			log.Fatal(err)
		}

		containerRegistryService := containerregistries.NewContainerRegistryService(containerregistries.ContainerRegistryServiceConfig{
			Store: containerRegistryStore,
		})

		projectConfigService := projectconfig.NewConfigService(projectconfig.ProjectConfigServiceConfig{
			ConfigStore: projectConfigStore,
		})

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
			})
			c.BuilderRegistryServer = util.GetFrpcRegistryDomain(c.Id, c.Frps.Domain)
		}

		providerTargetService := providertargets.NewProviderTargetService(providertargets.ProviderTargetServiceConfig{
			TargetStore: providerTargetStore,
		})

		apiKeyService := apikeys.NewApiKeyService(apikeys.ApiKeyServiceConfig{
			ApiKeyStore: apiKeyStore,
		})

		headscaleUrl := util.GetFrpcHeadscaleUrl(c.Frps.Protocol, c.Id, c.Frps.Domain)

		providerManager := manager.NewProviderManager(manager.ProviderManagerConfig{
			LogsDir:               logsDir,
			ProviderTargetService: providerTargetService,
			ApiUrl:                util.GetFrpcApiUrl(c.Frps.Protocol, c.Id, c.Frps.Domain),
			DaytonaDownloadUrl:    getDaytonaScriptUrl(c),
			ServerUrl:             headscaleUrl,
			RegistryUrl:           c.RegistryUrl,
			BaseDir:               c.ProvidersDir,
			CreateProviderNetworkKey: func(providerName string) (string, error) {
				return headscaleServer.CreateAuthKey()
			},
			ServerPort: c.HeadscalePort,
			ApiPort:    c.ApiPort,
		})

		buildImageNamespace := c.BuildImageNamespace
		if buildImageNamespace != "" {
			buildImageNamespace = fmt.Sprintf("/%s", buildImageNamespace)
		}
		buildImageNamespace = strings.TrimSuffix(buildImageNamespace, "/")

		builderFactory := build.NewBuilderFactory(build.BuilderFactoryConfig{
			ContainerRegistryServer:  c.BuilderRegistryServer,
			BuildImageNamespace:      buildImageNamespace,
			BuildStore:               buildStore,
			LoggerFactory:            loggerFactory,
			DefaultProjectImage:      c.DefaultProjectImage,
			DefaultProjectUser:       c.DefaultProjectUser,
			Image:                    c.BuilderImage,
			ContainerRegistryService: containerRegistryService,
		})

		provisioner := provisioner.NewProvisioner(provisioner.ProvisionerConfig{
			ProviderManager: providerManager,
		})

		gitProviderService := gitproviders.NewGitProviderService(gitproviders.GitProviderServiceConfig{
			ConfigStore: gitProviderConfigStore,
		})

		workspaceService := workspaces.NewWorkspaceService(workspaces.WorkspaceServiceConfig{
			WorkspaceStore:           workspaceStore,
			TargetStore:              providerTargetStore,
			ApiKeyService:            apiKeyService,
			GitProviderService:       gitProviderService,
			ContainerRegistryService: containerRegistryService,
			ProjectConfigService:     projectConfigService,
			ServerApiUrl:             util.GetFrpcApiUrl(c.Frps.Protocol, c.Id, c.Frps.Domain),
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

		buildRunnerConfig, err := build.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		buildRunner := build.NewBuildRunner(build.BuildRunnerInstanceConfig{
			Interval:         buildRunnerConfig.Interval,
			Scheduler:        build.NewCronScheduler(),
			BuildRunnerId:    buildRunnerConfig.Id,
			BuildStore:       buildStore,
			BuilderFactory:   builderFactory,
			LoggerFactory:    loggerFactory,
			TelemetryService: telemetryService,
		})

		server := server.GetInstance(&server.ServerInstanceConfig{
			Config:                   *c,
			TailscaleServer:          headscaleServer,
			ProviderTargetService:    providerTargetService,
			ContainerRegistryService: containerRegistryService,
			ProjectConfigService:     projectConfigService,
			LocalContainerRegistry:   localContainerRegistry,
			ApiKeyService:            apiKeyService,
			WorkspaceService:         workspaceService,
			GitProviderService:       gitProviderService,
			ProviderManager:          providerManager,
			ProfileDataService:       profileDataService,
			TelemetryService:         telemetryService,
		})

		errCh := make(chan error)

		err = server.Start(errCh)
		if err != nil {
			log.Fatal(err)
		}

		err = buildRunner.Start()
		if err != nil {
			log.Fatal(err)
		}

		go func() {
			err := apiServer.Start()
			if err != nil {
				log.Fatal(err)
			}
		}()

		go func() {
			err := <-errCh
			if err != nil {
				buildRunner.Stop()
				log.Fatal(err)
			}
		}()

		err = waitForServerToStart(apiServer)

		if err != nil {
			log.Fatal(err)
		}

		printServerStartedMessage(c, false)

		err = setDefaultConfig(server, c.ApiPort)
		if err != nil {
			log.Fatal(err)
		}

		err = <-errCh
		if err != nil {
			log.Fatal(err)
		}
	},
}

func waitForServerToStart(apiServer *api.ApiServer) error {
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

func setDefaultConfig(server *server.Server, apiPort uint32) error {
	existingConfig, err := config.GetConfig()
	if err != nil && !config.IsNotExist(err) {
		return err
	}

	if existingConfig != nil {
		for _, profile := range existingConfig.Profiles {
			if profile.Id == "default" {
				return nil
			}
		}
	}

	apiKey, err := server.ApiKeyService.Generate(apikey.ApiKeyTypeClient, "default")
	if err != nil {
		return err
	}

	if existingConfig != nil {
		err := existingConfig.AddProfile(config.Profile{
			Id:   "default",
			Name: "default",
			Api: config.ServerApi{
				Url: fmt.Sprintf("http://localhost:%d", apiPort),
				Key: apiKey,
			},
		})
		if err != nil {
			return err
		}

		return existingConfig.Save()
	}

	config := &config.Config{
		ActiveProfileId: "default",
		DefaultIdeId:    config.DefaultIdeId,
		Profiles: []config.Profile{
			{
				Id:   "default",
				Name: "default",
				Api: config.ServerApi{
					Url: fmt.Sprintf("http://localhost:%d", apiPort),
					Key: apiKey,
				},
			},
		},
		TelemetryEnabled: true,
	}

	return config.Save()
}
