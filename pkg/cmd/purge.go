// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal"
	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/bootstrap"
	server_cmd "github.com/daytonaio/daytona/pkg/cmd/server"
	"github.com/daytonaio/daytona/pkg/cmd/workspace/create"
	"github.com/daytonaio/daytona/pkg/posthogservice"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/telemetry"
	"github.com/daytonaio/daytona/pkg/views"
	view "github.com/daytonaio/daytona/pkg/views/purge"
	"github.com/spf13/cobra"
)

var forceFlag bool

var purgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Purges all Daytona data from the current device",
	Long:  "Purges all Daytona data from the current device - including all local runner providers, configuration files and SSH files. This command is irreversible.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var confirmCheck bool
		var serverStoppedCheck bool
		var defaultProfileNoticeConfirm bool

		c, err := config.GetConfig()
		if err != nil {
			return err
		}

		serverConfig, err := server.GetConfig()
		if err != nil {
			return err
		}

		serverConfigDir, err := server.GetConfigDir()
		if err != nil {
			return err
		}

		if c.ActiveProfileId != "default" {
			if !create.YesFlag {
				view.DefaultProfileNoticePrompt(&defaultProfileNoticeConfirm)
				if !defaultProfileNoticeConfirm {
					fmt.Println("Operation cancelled.")
					return nil
				}
			}
		}

		defaultProfile, err := c.GetProfile("default")
		if err != nil {
			return err
		}

		apiClient, err := apiclient.GetApiClient(&defaultProfile)
		if err != nil {
			if !apiclient.IsHealthCheckFailed(err) {
				return err
			}
		} else {
			view.ServerStoppedPrompt(&serverStoppedCheck)
			if serverStoppedCheck {
				_, _, err = apiClient.DefaultAPI.HealthCheck(context.Background()).Execute()
				if err == nil {
					views.RenderInfoMessage("The Daytona Server is still running. Please stop it before continuing.")
					return nil
				}
			} else {
				fmt.Println("Operation cancelled.")
				return nil
			}
		}

		if !create.YesFlag {
			view.ConfirmPrompt(&confirmCheck)
			if !confirmCheck {
				fmt.Println("Operation cancelled.")
				return nil
			}
		}

		telemetryService := posthogservice.NewTelemetryService(posthogservice.PosthogServiceConfig{
			ApiKey:   internal.PosthogApiKey,
			Endpoint: internal.PosthogEndpoint,
			Version:  internal.Version,
			Source:   telemetry.CLI_SOURCE,
		})

		defer telemetryService.Close()

		server, err := bootstrap.GetInstance(serverConfig, serverConfigDir, internal.Version, telemetryService)
		if err != nil {
			return err
		}

		ctx := context.Background()
		ctx = context.WithValue(ctx, telemetry.CLIENT_ID_CONTEXT_KEY, config.GetClientId())
		ctx = context.WithValue(ctx, telemetry.ENABLED_CONTEXT_KEY, c.TelemetryEnabled)

		// Get all targets, workspaces and builds to prompt user for resource purge
		targets, err := server.TargetService.ListTargets(ctx, nil, services.TargetRetrievalParams{})
		if err != nil {
			if !forceFlag {
				return err
			} else {
				fmt.Printf("Failed to get targets: %v\n", err)
			}
		}

		workspaces, err := server.WorkspaceService.ListWorkspaces(ctx, services.WorkspaceRetrievalParams{})
		if err != nil {
			if !forceFlag {
				return err
			} else {
				fmt.Printf("Failed to get workspaces: %v\n", err)
			}
		}

		builds, err := server.BuildService.List(ctx, nil)
		if err != nil {
			if !forceFlag {
				return err
			} else {
				fmt.Printf("Failed to get builds: %v\n", err)
			}
		}

		if len(targets) != 0 || len(workspaces) != 0 || len(builds) != 0 {
			var continuePurge bool
			commands := view.PurgeResourcesPrompt(&continuePurge, len(targets), len(workspaces), len(builds))
			if err != nil {
				if !forceFlag {
					return err
				} else {
					fmt.Printf("Failed to prompt for resource purge: %v\n", err)
				}
			}
			if !continuePurge {
				fmt.Printf("\nOperation cancelled.\nManually delete leftover resources for a complete purge by starting the server and running the following commands:\n\n%s\n", strings.Join(commands, "\n"))
				return nil
			}
		}

		fmt.Println("Purging the server")

		if server.LocalContainerRegistry != nil {
			fmt.Println("Purging local container registry...")
			err := server.LocalContainerRegistry.Purge()
			if err != nil {
				if !forceFlag {
					return err
				} else {
					fmt.Printf("Failed to purge local container registry: %v\n", err)
				}
			}
		}

		fmt.Println("Purging Tailscale server...")
		err = server.TailscaleServer.Purge()
		if err != nil {
			if !forceFlag {
				return err
			} else {
				fmt.Printf("Failed to purge Tailscale server: %v\n", err)
			}
		}

		localRunnerConfig := server_cmd.GetLocalRunnerConfig(filepath.Join(serverConfigDir, "local-runner"), c.TelemetryEnabled, c.Id)

		params := bootstrap.LocalRunnerParams{
			ServerConfig:     serverConfig,
			RunnerConfig:     localRunnerConfig,
			ConfigDir:        serverConfigDir,
			TelemetryService: telemetryService,
		}

		localRunner, err := bootstrap.GetLocalRunner(params)
		if err != nil {
			return err
		}

		if localRunner != nil {
			fmt.Println("Purging providers...")
			err = localRunner.Purge(ctx)
			if err != nil {
				if !forceFlag {
					return err
				} else {
					fmt.Printf("Failed to purge local runner providers: %v\n", err)
				}
			} else {
				fmt.Println("Providers purged.")
			}
		}

		fmt.Println("Server purged.")

		fmt.Println("\nDeleting the SSH configuration file")
		err = config.UnlinkSshFiles()
		if err != nil {
			return err
		}

		fmt.Println("Deleting autocompletion data")
		err = config.DeleteAutocompletionData()
		if err != nil {
			fmt.Printf("Error deleting autocompletion data: %s\n", err)
		}

		fmt.Println("Deleting the Daytona config directory")
		err = config.DeleteConfigDir()
		if err != nil {
			return err
		}

		binaryMessage := "You may now delete the binary"
		binaryPath, err := os.Executable()
		if err == nil {
			binaryMessage = fmt.Sprintf("You may now delete the binary by running: sudo rm %s", binaryPath)
		}

		views.RenderInfoMessage(fmt.Sprintf("All Daytona data has been successfully cleared from the device.\n%s", binaryMessage))
		return nil
	},
}

func init() {
	purgeCmd.Flags().BoolVarP(&create.YesFlag, "yes", "y", false, "Execute purge without a prompt")
	purgeCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Delete all targets by force")
}
