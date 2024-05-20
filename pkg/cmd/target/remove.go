// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util/apiclient"
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

		ctx := context.Background()
		client, err := apiclient.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		res, err := client.TargetAPI.RemoveTarget(ctx, selectedTargetName).Execute()
		if err != nil {
			log.Fatal(apiclient.HandleErrorResponse(res, err))
		}

		workspaceList, res, err := client.WorkspaceAPI.ListWorkspaces(ctx).Execute()
		if err != nil {
			log.Error(apiclient.HandleErrorResponse(res, err))
		}

		if len(workspaceList) > 0 {
			views.RenderInfoMessage(fmt.Sprintln("Deleting workspaces within target..."))
			for _, workspace := range workspaceList {
				if *workspace.Target != selectedTargetName {
					continue
				}

				res, err := client.WorkspaceAPI.RemoveWorkspace(ctx, *workspace.Id).Execute()
				if err != nil {
					log.Errorf("Failed to delete workspace %s: %v", *workspace.Name, apiclient.HandleErrorResponse(res, err))
					continue
				}

				views.RenderLine(fmt.Sprintf("- Workspace %s successfully deleted\n", *workspace.Name))

			}
		}

		views.RenderInfoMessageBold(fmt.Sprintf("Target %s removed successfully", selectedTargetName))
	},
}
