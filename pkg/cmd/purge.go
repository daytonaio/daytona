// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal"
	"github.com/daytonaio/daytona/internal/util/apiclient"
	server_cmd "github.com/daytonaio/daytona/pkg/cmd/server"
	"github.com/daytonaio/daytona/pkg/posthogservice"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/telemetry"
	"github.com/daytonaio/daytona/pkg/views"
	view "github.com/daytonaio/daytona/pkg/views/purge"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var yesFlag bool
var forceFlag bool

var purgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Purges all Daytona data from the current device",
	Long:  "Purges all Daytona data from the current device - including all workspaces, configuration files, and SSH files. This command is irreversible.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		var confirmCheck bool
		var serverStoppedCheck bool
		var defaultProfileNoticeConfirm bool

		c, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		serverConfig, err := server.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		serverConfigDir, err := server.GetConfigDir()
		if err != nil {
			log.Fatal(err)
		}

		if c.ActiveProfileId != "default" {
			if !yesFlag {
				view.DefaultProfileNoticePrompt(&defaultProfileNoticeConfirm)
				if !defaultProfileNoticeConfirm {
					fmt.Println("Operation cancelled.")
					return
				}
			}
		}

		defaultProfile, err := c.GetProfile("default")
		if err != nil {
			log.Fatal(err)
		}

		_, err = apiclient.GetApiClient(&defaultProfile)
		if err != nil {
			if !apiclient.IsHealthCheckFailed(err) {
				log.Fatal(err)
			}
		} else {
			view.ServerStoppedPrompt(&serverStoppedCheck)
			if !serverStoppedCheck {
				fmt.Println("Operation cancelled.")
				return
			}
		}

		if !yesFlag {
			view.ConfirmPrompt(&confirmCheck)
			if !confirmCheck {
				fmt.Println("Operation cancelled.")
				return
			}
		}

		telemetryService := posthogservice.NewTelemetryService(posthogservice.PosthogServiceConfig{
			ApiKey:   internal.PosthogApiKey,
			Endpoint: internal.PosthogEndpoint,
		})

		defer telemetryService.Close()

		fmt.Println("Purging the server")
		server, err := server_cmd.GetInstance(serverConfig, serverConfigDir, telemetryService)
		if err != nil {
			log.Fatal(err)
		}

		ctx := context.Background()
		ctx = context.WithValue(ctx, telemetry.CLIENT_ID_CONTEXT_KEY, c.Id)
		ctx = context.WithValue(ctx, telemetry.ENABLED_CONTEXT_KEY, c.TelemetryEnabled)

		err = server.Purge(ctx, forceFlag)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Server purged.")

		fmt.Println("\nDeleting the SSH configuration file")
		err = config.UnlinkSshFiles()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Deleting autocompletion data")
		err = config.DeleteAutocompletionData()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Deleting the Daytona config directory")
		err = config.DeleteConfigDir()
		if err != nil {
			log.Fatal(err)
		}

		binaryMessage := "You may now delete the binary"
		binaryPath, err := os.Executable()
		if err == nil {
			binaryMessage = fmt.Sprintf("You may now delete the binary by running: sudo rm %s", binaryPath)
		}

		views.RenderInfoMessage(fmt.Sprintf("All Daytona data has been successfully cleared from the device.\n%s", binaryMessage))
	},
}

func init() {
	purgeCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Execute purge without prompt")
	purgeCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Delete all workspaces by force")
}
