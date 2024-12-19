// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/api"
	"github.com/daytonaio/daytona/pkg/cmd/bootstrap"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/posthogservice"
	"github.com/daytonaio/daytona/pkg/runner"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/server/registry"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
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

		server, err := bootstrap.GetInstance(c, configDir, internal.Version, telemetryService)
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

		localRunnerErrChan := make(chan error)

		go func() {
			if c.LocalRunnerDisabled != nil && *c.LocalRunnerDisabled {
				err = handleDisabledLocalRunner()
				if err != nil {
					localRunnerErrChan <- err
				}
				return
			}

			localRunnerConfig := getLocalRunnerConfig(filepath.Join(configDir, "local-runner"))

			localRunnerErrChan <- startLocalRunner(bootstrap.LocalRunnerParams{
				ServerConfig:     c,
				RunnerConfig:     localRunnerConfig,
				ConfigDir:        configDir,
				TelemetryService: telemetryService,
			})
		}()

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
		case err := <-localRunnerErrChan:
			return err
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

func printServerStartedMessage(c *server.Config, runAsDaemon bool) {
	started_view.Render(c.ApiPort, util.GetFrpcApiUrl(c.Frps.Protocol, c.Id, c.Frps.Domain), runAsDaemon)
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

	apiKey, err := server.ApiKeyService.Generate(context.Background(), models.ApiKeyTypeClient, "default")
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

func startLocalRunner(params bootstrap.LocalRunnerParams) error {
	runnerService := server.GetInstance(nil).RunnerService

	_, err := runnerService.GetRunner(context.Background(), bootstrap.LOCAL_RUNNER_ID)
	if err != nil {
		if stores.IsRunnerNotFound(err) {
			_, err := runnerService.RegisterRunner(context.Background(), services.RegisterRunnerDTO{
				Id:   bootstrap.LOCAL_RUNNER_ID,
				Name: bootstrap.LOCAL_RUNNER_ID,
			})
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	runner, err := bootstrap.GetLocalRunner(params)
	if err != nil {
		return err
	}

	return runner.Start(context.Background())
}

func getLocalRunnerConfig(configDir string) *runner.Config {
	providersDir := filepath.Join(configDir, "providers")
	logFilePath := filepath.Join(configDir, "runner.log")

	return &runner.Config{
		Id:           bootstrap.LOCAL_RUNNER_ID,
		Name:         bootstrap.LOCAL_RUNNER_ID,
		ProvidersDir: providersDir,
		LogFile:      logs.GetDefaultLogFileConfig(logFilePath),
	}
}

func handleDisabledLocalRunner() error {
	runnerService := server.GetInstance(nil).RunnerService

	_, err := runnerService.GetRunner(context.Background(), bootstrap.LOCAL_RUNNER_ID)
	if err != nil {
		if stores.IsRunnerNotFound(err) {
			return nil
		}
	}

	return runnerService.RemoveRunner(context.Background(), bootstrap.LOCAL_RUNNER_ID)
}
