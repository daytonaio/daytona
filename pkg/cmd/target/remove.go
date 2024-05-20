// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util/apiclient"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/target"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

var targetRemoveCmd = &cobra.Command{
	Use:     "remove [TARGET_NAME]",
	Short:   "Remove target",
	Args:    cobra.RangeArgs(0, 1),
	Aliases: []string{"rm", "delete"},
	Run: func(cmd *cobra.Command, args []string) {
		var selectedTargetName string

		c, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		if len(args) == 0 {
			activeProfile, err := c.GetActiveProfile()
			if err != nil {
				log.Fatal(err)
			}

			targets, err := apiclient.GetTargetList()
			if err != nil {
				log.Fatal(err)
			}

			selectedTarget, err := target.GetTargetFromPrompt(targets, activeProfile.Name, false)
			if err != nil {
				log.Fatal(err)
			}

			selectedTargetName = *selectedTarget.Name
		} else {
			selectedTargetName = args[0]
		}

		client, err := apiclient.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		res, err := client.TargetAPI.RemoveTarget(context.Background(), selectedTargetName).Execute()
		if err != nil {
			log.Fatal(apiclient.HandleErrorResponse(res, err))
		}

		workspaceClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		workspaceList, res, err := workspaceClient.WorkspaceAPI.ListWorkspaces(context.Background()).Execute()
		if err != nil {
			log.Fatal(apiclient_util.HandleErrorResponse(res, err))
		}

		for _, workspace := range workspaceList {
			res, err := workspaceClient.WorkspaceAPI.RemoveWorkspace(context.Background(), *workspace.Id).Execute()

			if err != nil {
				log.Fatal(apiclient_util.HandleErrorResponse(res, err))
			}

			views.RenderLine(fmt.Sprintf("- Workspace %s successfully deleted\n", *workspace.Name))
		}

		views.RenderInfoMessageBold(fmt.Sprintf("Target %s removed successfully", selectedTargetName))
	},
}
