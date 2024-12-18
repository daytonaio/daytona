// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runner

import (
	"context"
	"io"
	"os"

	"github.com/daytonaio/daytona/internal"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/bootstrap"
	"github.com/daytonaio/daytona/pkg/posthogservice"
	"github.com/daytonaio/daytona/pkg/runner"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var daemonServeCmd = &cobra.Command{
	Use:    "daemon-serve",
	Short:  "Used by the daemon to start the Daytona Runner",
	Args:   cobra.NoArgs,
	Hidden: true,
	RunE:   serveCmd.RunE,
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts the runner in the foreground",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		if log.GetLevel() < log.InfoLevel {
			log.SetLevel(log.InfoLevel)
		}

		runnerConfig, err := runner.GetConfig()
		if err != nil {
			return err
		}

		runnerConfigDir, err := runner.GetConfigDir()
		if err != nil {
			return err
		}

		telemetryService := posthogservice.NewTelemetryService(posthogservice.PosthogServiceConfig{
			ApiKey:   internal.PosthogApiKey,
			Endpoint: internal.PosthogEndpoint,
			Version:  internal.Version,
		})

		apiClient, err := apiclient_util.GetRunnerApiClient(runnerConfig.ServerApiUrl, runnerConfig.ServerApiKey, runnerConfig.ClientId, runnerConfig.TelemetryEnabled)
		if err != nil {
			return err
		}

		serverConfig, _, err := apiClient.ServerAPI.GetConfig(ctx).Execute()
		if err != nil {
			return err
		}

		err = bootstrap.InitRemoteProviderManager(apiClient, serverConfig, runnerConfig, runnerConfigDir)
		if err != nil {
			return err
		}

		var runnerLogWriter io.Writer

		if runnerConfig.LogFile != nil {
			logFile, err := os.OpenFile(runnerConfig.LogFile.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return err
			}
			defer logFile.Close()
			runnerLogWriter = logFile
		}

		runner, err := bootstrap.GetRemoteRunner(bootstrap.RemoteRunnerParams{
			ApiClient:        apiClient,
			ServerConfig:     serverConfig,
			RunnerConfig:     runnerConfig,
			ConfigDir:        runnerConfigDir,
			LogWriter:        runnerLogWriter,
			TelemetryService: telemetryService,
		})
		if err != nil {
			return err
		}

		return runner.Start(ctx)
	},
}
