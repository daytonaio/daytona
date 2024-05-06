// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/daytonaio/daytona/cmd/daytona/config"
	workspace "github.com/daytonaio/daytona/pkg/cmd/workspace"
	"github.com/daytonaio/daytona/pkg/views"
	profile_view "github.com/daytonaio/daytona/pkg/views/profile"
	view "github.com/daytonaio/daytona/pkg/views/purge"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var yesFlag bool

var purgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Purges all Daytona data from the current device",
	Long:  "Purges all Daytona data from the current device - including all workspaces, configuration files, and SSH files. This command is irreversible.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		var confirmCheck bool
		var serverStoppedCheck bool
		var switchProfileCheck bool

		c, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		if c.ActiveProfileId != "default" {
			profile_view.SwitchToDefaultPrompt(&switchProfileCheck)
			if !switchProfileCheck {
				views.RenderInfoMessage("Operation cancelled.")
				return
			}
			c.ActiveProfileId = "default"
			err = c.Save()
			if err != nil {
				log.Fatal(err)
			}
		}

		if !yesFlag {
			view.ConfirmPrompt(&confirmCheck)
			if !confirmCheck {
				views.RenderInfoMessage("Operation cancelled.")
				return
			}
		}

		views.RenderLine("\nDeleting all workspaces")
		err = workspace.DeleteAllWorkspaces()
		if err != nil {
			log.Fatal(err)
		}

		views.RenderLine("Deleting the SSH configuration file")
		err = config.UnlinkSshFiles()
		if err != nil {
			log.Fatal(err)
		}

		views.RenderLine("Deleting autocompletion data")
		err = config.DeleteAutocompletionData()
		if err != nil {
			log.Fatal(err)
		}

		view.ServerStoppedPrompt(&serverStoppedCheck)
		if !serverStoppedCheck {
			views.RenderInfoMessage("Operation cancelled.")
			return
		}

		views.RenderLine("Deleting the Daytona config directory")
		err = config.DeleteConfigDir()
		if err != nil {
			log.Fatal(err)
		}

		views.RenderInfoMessage("All Daytona data has been successfully cleared from the device.\nYou may now delete the binary.")
	},
}

func init() {
	purgeCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Execute purge without prompt")
}
