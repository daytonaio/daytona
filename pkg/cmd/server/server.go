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
	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/views"
	view "github.com/daytonaio/daytona/pkg/views/server"
	"github.com/fatih/color"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var yesFlag bool
var allRequirementsMet bool

type RequirementProvider struct {
	provider.Provider
}

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
		infoColor := color.New(color.FgHiBlue).SprintFunc()
		warningColor := color.New(color.FgRed).SprintFunc()
		p := RequirementProvider{}

		requirements, err := p.CheckRequirements()
		if err != nil {
			return err
		}

		for _, req := range requirements {
			if req.Met {
				allRequirementsMet = true
				fmt.Printf("%s[0000]     Requirement met: %s\n", infoColor("INFO"), req.Reason)
			} else {
				allRequirementsMet = false
				fmt.Printf("%s[0000]  Requirement not met: %s\n", warningColor("WARNING"), req.Reason)
			}
			if !allRequirementsMet {
				return fmt.Errorf("    Daytona server startup aborted, one or more requirement not met")
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
		err = daemon.Start(c.LogFilePath)
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
