// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/api"
	"github.com/daytonaio/daytona/pkg/apikey"
	"github.com/daytonaio/daytona/pkg/db"
	"github.com/daytonaio/daytona/pkg/logger"
	"github.com/daytonaio/daytona/pkg/provider/manager"
	"github.com/daytonaio/daytona/pkg/provisioner"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/server/apikeys"
	"github.com/daytonaio/daytona/pkg/server/containerregistries"
	"github.com/daytonaio/daytona/pkg/server/gitproviders"
	"github.com/daytonaio/daytona/pkg/server/headscale"
	"github.com/daytonaio/daytona/pkg/server/providertargets"
	"github.com/daytonaio/daytona/pkg/server/workspaces"
	started_view "github.com/daytonaio/daytona/pkg/views/server/started"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var ServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the server process in the current terminal session",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if log.GetLevel() < log.InfoLevel {
			//	for now, force the log level to info when running the server
			log.SetLevel(log.InfoLevel)
		}

		c, err := server.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		apiServer := api.NewApiServer(api.ApiServerConfig{
			ApiPort: int(c.ApiPort),
		})

		logsDir, err := server.GetWorkspaceLogsDir()
		if err != nil {
			log.Fatal(err)
		}
		loggerFactory := logger.NewLoggerFactory(logsDir)

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

		headscaleServer := headscale.NewHeadscaleServer(&headscale.HeadscaleServerConfig{
			ServerId:      c.Id,
			FrpsDomain:    c.Frps.Domain,
			FrpsProtocol:  c.Frps.Protocol,
			HeadscalePort: c.HeadscalePort,
		})
		err = headscaleServer.Init()
		if err != nil {
			log.Fatal(err)
		}

		containerRegistryService := containerregistries.NewContainerRegistryService(containerregistries.ContainerRegistryServiceConfig{
			Store: containerRegistryStore,
		})

		providerTargetService := providertargets.NewProviderTargetService(providertargets.ProviderTargetServiceConfig{
			TargetStore: providerTargetStore,
		})
		apiKeyService := apikeys.NewApiKeyService(apikeys.ApiKeyServiceConfig{
			ApiKeyStore: apiKeyStore,
		})
		providerManager := manager.NewProviderManager(manager.ProviderManagerConfig{
			LogsDir:               logsDir,
			ProviderTargetService: providerTargetService,
			ServerApiUrl:          util.GetFrpcApiUrl(c.Frps.Protocol, c.Id, c.Frps.Domain),
			ServerDownloadUrl:     getDaytonaScriptUrl(c),
			ServerUrl:             util.GetFrpcServerUrl(c.Frps.Protocol, c.Id, c.Frps.Domain),
			RegistryUrl:           c.RegistryUrl,
			BaseDir:               c.ProvidersDir,
		})
		provisioner := provisioner.NewProvisioner(provisioner.ProvisionerConfig{
			ProviderManager: providerManager,
		})

		workspaceService := workspaces.NewWorkspaceService(workspaces.WorkspaceServiceConfig{
			WorkspaceStore:                  workspaceStore,
			TargetStore:                     providerTargetStore,
			ApiKeyService:                   apiKeyService,
			ContainerRegistryStore:          containerRegistryStore,
			ServerApiUrl:                    util.GetFrpcApiUrl(c.Frps.Protocol, c.Id, c.Frps.Domain),
			ServerUrl:                       util.GetFrpcServerUrl(c.Frps.Protocol, c.Id, c.Frps.Domain),
			DefaultProjectImage:             c.DefaultProjectImage,
			DefaultProjectUser:              c.DefaultProjectUser,
			DefaultProjectPostStartCommands: c.DefaultProjectPostStartCommands,
			Provisioner:                     provisioner,
			LoggerFactory:                   loggerFactory,
		})
		gitProviderService := gitproviders.NewGitProviderService(gitproviders.GitProviderServiceConfig{
			ConfigStore: gitProviderConfigStore,
		})

		server := server.GetInstance(&server.ServerInstanceConfig{
			Config:                   *c,
			TailscaleServer:          headscaleServer,
			ProviderTargetService:    providerTargetService,
			ContainerRegistryService: containerRegistryService,
			ApiKeyService:            apiKeyService,
			WorkspaceService:         workspaceService,
			GitProviderService:       gitProviderService,
			ProviderManager:          providerManager,
		})

		errCh := make(chan error)

		err = server.Start(errCh)
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
				log.Fatal(err)
			}
		}()

		err = waitForServerToStart(apiServer)

		if err != nil {
			log.Fatal(err)
		}

		printServerStartedMessage(c, false)

		err = setDefaultConfig(server)
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
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(userConfigDir, "daytona")

	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "db"), nil
}

func setDefaultConfig(server *server.Server) error {
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
				Url: "http://localhost:3000",
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
					Url: "http://localhost:3000",
					Key: apiKey,
				},
			},
		},
	}

	return config.Save()
}
