// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"context"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/daytonaio/daytona/cmd/daytona/config"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	workspace_cmd "github.com/daytonaio/daytona/pkg/cmd/workspace"
	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/target"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

var yesFlag bool

var targetRemoveCmd = &cobra.Command{
	Use:     "remove [TARGET_NAME]",
	Short:   "Remove target",
	Args:    cobra.RangeArgs(0, 1),
	Aliases: []string{"rm", "delete"},
	Run: func(cmd *cobra.Command, args []string) {
		var selectedTargetName string

		ctx := context.Background()
		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		c, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		if len(args) == 0 {
			activeProfile, err := c.GetActiveProfile()
			if err != nil {
				log.Fatal(err)
			}

			targetList, res, err := apiClient.TargetAPI.ListTargets(ctx).Execute()
			if err != nil {
				log.Fatal(apiclient_util.HandleErrorResponse(res, err))
			}

			selectedTarget, err := target.GetTargetFromPrompt(targetList, activeProfile.Name, false)
			if err != nil {
				if common.IsCtrlCAbort(err) {
					return
				} else {
					log.Fatal(err)
				}
			}

			selectedTargetName = selectedTarget.Name
		} else {
			selectedTargetName = args[0]
		}

		if yesFlag {
			fmt.Println("Deleting all workspaces.")
			err := RemoveTargetWorkspaces(ctx, apiClient, selectedTargetName)

			if err != nil {
				log.Fatal(err)
			}
		} else {
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						Title(fmt.Sprintf("Delete all workspaces within %s?", selectedTargetName)).
						Description("You might not be able to easily remove these workspaces later.").
						Value(&yesFlag),
				),
			).WithTheme(views.GetCustomTheme())

			err := form.Run()
			if err != nil {
				log.Fatal(err)
			}

			if yesFlag {
				err := RemoveTargetWorkspaces(ctx, apiClient, selectedTargetName)
				if err != nil {
					log.Fatal(err)
				}
			} else {
				fmt.Println("Proceeding with target removal without deleting workspaces.")
			}
		}

		res, err := apiClient.TargetAPI.RemoveTarget(ctx, selectedTargetName).Execute()
		if err != nil {
			log.Fatal(apiclient_util.HandleErrorResponse(res, err))
		}

		views.RenderInfoMessageBold(fmt.Sprintf("Target %s removed successfully", selectedTargetName))
	},
}

func init() {
	targetRemoveCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Confirm deletion of all workspaces without prompt")
}

func RemoveTargetWorkspaces(ctx context.Context, client *apiclient.APIClient, target string) error {
	workspaceList, res, err := client.WorkspaceAPI.ListWorkspaces(ctx).Execute()
	if err != nil {
		return apiclient_util.HandleErrorResponse(res, err)
	}

	for _, workspace := range workspaceList {
		if workspace.Target != target {
			continue
		}
		err := workspace_cmd.RemoveWorkspace(ctx, client, &workspace, false)
		if err != nil {
			log.Errorf("Failed to delete workspace %s: %v", workspace.Name, err)
			continue
		}

		views.RenderInfoMessage(fmt.Sprintf("- Workspace '%s' successfully deleted", workspace.Name))
	}

	return nil
}
