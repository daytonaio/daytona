// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/api"
	"github.com/daytonaio/daytona/pkg/cmd/server/daemon"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/views"
	view "github.com/daytonaio/daytona/pkg/views/server"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var yesFlag bool

var ServerCmd = &cobra.Command{
	Use:     "server",
	Short:   "Start the server process in daemon mode",
	GroupID: util.SERVER_GROUP,
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		confirmCheck := true

		if !yesFlag {
			view.ConfirmPrompt(&confirmCheck)
			if !confirmCheck {
				views.RenderInfoMessage("Operation cancelled.")
				return
			}
		}

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

		views.RenderInfoMessageBold("Starting the Daytona Server daemon...")
		err = daemon.Start(c.LogFilePath)
		if err != nil {
			log.Fatal(err)
		}
		err = waitForServerToStart(apiServer)
		if err != nil {
			log.Fatal(err)
		}
		printServerStartedMessage(c, true)
	},
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the Daytona Server daemon",
	Run:   ServerCmd.Run,
}

func init() {
	ServerCmd.AddCommand(configureCmd)
	ServerCmd.AddCommand(configCmd)
	ServerCmd.AddCommand(logsCmd)
	ServerCmd.AddCommand(startCmd)
	ServerCmd.AddCommand(stopCmd)
	ServerCmd.AddCommand(restartCmd)
	ServerCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Skip the confirmation prompt")
}
