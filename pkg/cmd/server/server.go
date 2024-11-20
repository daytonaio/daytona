// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"fmt"
	"os"
	"runtime"

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
	RunE: func(cmd *cobra.Command, args []string) error {
		confirmCheck := true

		if !yesFlag {
			view.ConfirmPrompt(&confirmCheck)
			if !confirmCheck {
				views.RenderInfoMessage("Operation cancelled.")
				return nil
			}
		}

		if log.GetLevel() < log.InfoLevel {
			//	for now, force the log level to info when running the server
			log.SetLevel(log.InfoLevel)
		}

		c, err := server.GetConfig()
		if err != nil {
			return err
		}

		apiServer := api.NewApiServer(api.ApiServerConfig{
			ApiPort: int(c.ApiPort),
		})

		views.RenderInfoMessageBold("Starting the Daytona Server daemon...")
		err = daemon.Start(c.LogFile.Path)
		if err != nil {
			return err
		}
		err = waitForApiServerToStart(apiServer)
		if err != nil {
			return err
		}
		printServerStartedMessage(c, true)

		switch runtime.GOOS {
		case "linux":
			fmt.Printf("Use `loginctl enable-linger %s` to allow the service to run after logging out.\n", os.Getenv("USER"))
		}
		return nil
	},
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the Daytona Server daemon",
	RunE:  ServerCmd.RunE,
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
