// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal"
	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/bootstrap"
	"github.com/daytonaio/daytona/pkg/cmd/workspace/create"
	"github.com/daytonaio/daytona/pkg/posthogservice"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/telemetry"
	"github.com/daytonaio/daytona/pkg/views"
	view "github.com/daytonaio/daytona/pkg/views/purge"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var forceFlag bool

var purgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Purges all Daytona data from the current device",
	Long:  "Purges all Daytona data from the current device - including all targets, configuration files, and SSH files. This command is irreversible.",
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

		// FIXME: TODO
		// runnerConfig, err := runner.GetConfig()
		// if err != nil {
		// 	return err
		// }

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
		})

		defer telemetryService.Close()

		fmt.Println("Purging the server")
		server, err := bootstrap.GetInstance(serverConfig, serverConfigDir, internal.Version, telemetryService)
		if err != nil {
			return err
		}

		ctx := context.Background()
		ctx = context.WithValue(ctx, telemetry.CLIENT_ID_CONTEXT_KEY, config.GetClientId())
		ctx = context.WithValue(ctx, telemetry.ENABLED_CONTEXT_KEY, c.TelemetryEnabled)

		errCh := make(chan error)

		go func() {
			err := <-errCh
			if err != nil {
				if !forceFlag {
					log.Fatal(err)
				}
			}
		}()

		headscaleServerStartedChan := make(chan struct{})
		headscaleServerErrChan := make(chan error)

		go func() {
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
				localContainerRegistryErrChan <- server.LocalContainerRegistry.Start()
			} else {
				localContainerRegistryErrChan <- nil
			}
		}()

		select {
		case <-headscaleServerStartedChan:
			go func() {
				headscaleServerErrChan <- server.TailscaleServer.Connect()
			}()
		case err := <-headscaleServerErrChan:
			return err
		}

		err = <-localContainerRegistryErrChan
		if err != nil {
			return err
		}

		errs := server.Purge(ctx, forceFlag)
		if len(errs) > 0 {
			errMessage := ""
			for _, err := range errs {
				errMessage += fmt.Sprintf("Failed to purge: %v\n", err)
			}

			return errors.New(errMessage)
		}

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
	purgeCmd.Flags().BoolVarP(&create.YesFlag, "yes", "y", false, "Execute purge without prompt")
	purgeCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Delete all targets by force")
}
