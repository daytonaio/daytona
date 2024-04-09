// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/api"
	"github.com/daytonaio/daytona/pkg/apikey"
	apikeyCmd "github.com/daytonaio/daytona/pkg/cmd/server/apikey"
	"github.com/daytonaio/daytona/pkg/cmd/server/daemon"
	. "github.com/daytonaio/daytona/pkg/cmd/server/provider"
	. "github.com/daytonaio/daytona/pkg/cmd/server/target"
	"github.com/daytonaio/daytona/pkg/db"
	"github.com/daytonaio/daytona/pkg/logger"
	"github.com/daytonaio/daytona/pkg/provider/manager"
	"github.com/daytonaio/daytona/pkg/provisioner"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/server/apikeys"
	"github.com/daytonaio/daytona/pkg/server/gitproviders"
	"github.com/daytonaio/daytona/pkg/server/headscale"
	"github.com/daytonaio/daytona/pkg/server/providertargets"
	"github.com/daytonaio/daytona/pkg/server/workspaces"
	views_util "github.com/daytonaio/daytona/pkg/views/util"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var runAsDaemon bool

var ServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the server process",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if log.GetLevel() < log.InfoLevel {
			//	for now, force the log level to info when running the server
			log.SetLevel(log.InfoLevel)
		}

		if runAsDaemon {
			fmt.Println("Starting the Daytona Server daemon...")
			err := daemon.Start()
			if err != nil {
				log.Fatal(err)
			}
			c, err := server.GetConfig()
			if err != nil {
				log.Fatal(err)
			}
			printServerStartedMessage(c)
			return
		}

		c, err := server.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		logsDir, err := server.GetWorkspaceLogsDir()
		if err != nil {
			log.Fatal(err)
		}

		dbPath, err := getDbPath()
		if err != nil {
			log.Fatal(err)
		}

		dbConnection := db.GetSQLiteConnection(dbPath)
		apiKeyStore, err := db.NewApiKeyStore(dbConnection)
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

		providerTargetService := providertargets.NewProviderTargetService(providertargets.ProviderTargetServiceConfig{
			TargetStore: providerTargetStore,
		})
		apiKeyService := apikeys.NewApiKeyService(apikeys.ApiKeyServiceConfig{
			ApiKeyStore: apiKeyStore,
		})
		providerManager := manager.NewProviderManager(manager.ProviderManagerConfig{
			LogsDir:               logsDir,
			ProviderTargetService: *providerTargetService,
			ServerApiUrl:          util.GetFrpcApiUrl(c.Frps.Protocol, c.Id, c.Frps.Domain),
			ServerDownloadUrl:     getDaytonaScriptUrl(c),
			ServerUrl:             util.GetFrpcServerUrl(c.Frps.Protocol, c.Id, c.Frps.Domain),
			RegistryUrl:           c.RegistryUrl,
			BaseDir:               c.ProvidersDir,
		})
		provisioner := provisioner.NewProvisioner(provisioner.ProvisionerConfig{
			ProviderManager: *providerManager,
		})

		workspaceService := workspaces.NewWorkspaceService(workspaces.WorkspaceServiceConfig{
			WorkspaceStore: workspaceStore,
			TargetStore:    providerTargetStore,
			ApiKeyService:  *apiKeyService,
			ServerApiUrl:   util.GetFrpcApiUrl(c.Frps.Protocol, c.Id, c.Frps.Domain),
			ServerUrl:      util.GetFrpcServerUrl(c.Frps.Protocol, c.Id, c.Frps.Domain),
			Provisioner:    *provisioner,
			NewWorkspaceLogger: func(workspaceId string) logger.Logger {
				return logger.NewWorkspaceLogger(logsDir, workspaceId)
			},
			NewProjectLogger: func(workspaceId, projectName string) logger.Logger {
				return logger.NewProjectLogger(logsDir, workspaceId, projectName)
			},
			NewWorkspaceLogReader: func(workspaceId string) (io.Reader, error) {
				return logger.GetWorkspaceLogReader(logsDir, workspaceId)
			},
		})
		gitProviderService := gitproviders.NewGitProviderService(gitproviders.GitProviderServiceConfig{
			ConfigStore: gitProviderConfigStore,
		})

		server := server.GetInstance(&server.ServerInstanceConfig{
			Config:                *c,
			TailscaleServer:       headscaleServer,
			ProviderTargetService: *providerTargetService,
			ApiKeyService:         *apiKeyService,
			WorkspaceService:      *workspaceService,
			GitProviderService:    *gitProviderService,
			ProviderManager:       *providerManager,
		})

		errCh := make(chan error)

		err = server.Start(errCh)
		if err != nil {
			log.Fatal(err)
		}

		apiServer := api.NewApiServer(api.ApiServerConfig{
			ApiPort: int(c.ApiPort),
		})

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

		for i := 0; i < 3; i++ {
			err = apiServer.HealthCheck()
			if err != nil {
				time.Sleep(3 * time.Second)
				continue
			}

			printServerStartedMessage(c)
			break
		}

		if err != nil {
			log.Fatal(err)
		}

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

func getDaytonaScriptUrl(config *server.Config) string {
	url, _ := url.JoinPath(util.GetFrpcApiUrl(config.Frps.Protocol, config.Id, config.Frps.Domain), "binary", "script")
	return url
}

func printServerStartedMessage(c *server.Config) {
	views_util.RenderBorderedMessage(fmt.Sprintf("Daytona Server running on port: %d.\nYou can now begin developing locally.\n\nIf you want to connect to the server remotely:\n\n1. Create an API key on this machine:\ndaytona server api-key new\n\n2. On the client machine run:\ndaytona profile add -a %s -k API_KEY", c.ApiPort, util.GetFrpcApiUrl(c.Frps.Protocol, c.Id, c.Frps.Domain)))
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

	config := &config.Config{
		ActiveProfileId: "default",
		DefaultIdeId:    "vscode",
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

func init() {
	ServerCmd.PersistentFlags().BoolVarP(&runAsDaemon, "daemon", "d", false, "Run the server as a daemon")
	ServerCmd.AddCommand(configureCmd)
	ServerCmd.AddCommand(configCmd)
	ServerCmd.AddCommand(logsCmd)
	ServerCmd.AddCommand(TargetCmd)
	ServerCmd.AddCommand(ProviderCmd)
	ServerCmd.AddCommand(stopCmd)
	ServerCmd.AddCommand(restartCmd)
	ServerCmd.AddCommand(apikeyCmd.ApiKeyCmd)
}
