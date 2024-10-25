// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targetconfig

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
	"github.com/daytonaio/daytona/pkg/views/targetconfig"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

var yesFlag bool

var removeCmd = &cobra.Command{
	Use:     "remove [CONFIG_NAME]",
	Short:   "Remove target config",
	Args:    cobra.RangeArgs(0, 1),
	Aliases: []string{"rm", "delete"},
	RunE: func(cmd *cobra.Command, args []string) error {
		var selectedConfigName string

		ctx := context.Background()
		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		c, err := config.GetConfig()
		if err != nil {
			return err
		}

		if len(args) == 0 {
			activeProfile, err := c.GetActiveProfile()
			if err != nil {
				return err
			}

			targetConfigs, res, err := apiClient.TargetConfigAPI.ListTargetConfigs(ctx).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			if len(targetConfigs) == 0 {
				views_util.NotifyEmptyTargetConfigList(false)
				return nil
			}

			selectedTargetConfig, err := targetconfig.GetTargetConfigFromPrompt(targetConfigs, activeProfile.Name, nil, false, "Remove")
			if err != nil {
				if common.IsCtrlCAbort(err) {
					return nil
				} else {
					return err
				}
			}

			selectedConfigName = selectedTargetConfig.Name
		} else {
			selectedConfigName = args[0]
		}

		if yesFlag {
			fmt.Println("Deleting all workspaces.")
			err := RemoveTargetConfigWorkspaces(ctx, apiClient, selectedConfigName)

			if err != nil {
				return err
			}
		} else {
			var configWorkspaceCount int

			workspaceList, res, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			for _, workspace := range workspaceList {
				if workspace.TargetConfig == selectedConfigName {
					configWorkspaceCount++
				}
			}

			if configWorkspaceCount > 0 {
				title := fmt.Sprintf("Delete %d workspaces within %s?", configWorkspaceCount, selectedConfigName)
				description := "You might not be able to easily remove these workspaces later."

				if configWorkspaceCount == 1 {
					title = fmt.Sprintf("Delete %d workspace within %s?", configWorkspaceCount, selectedConfigName)
					description = "You might not be able to easily remove this workspace later."
				}

				form := huh.NewForm(
					huh.NewGroup(
						huh.NewConfirm().
							Title(title).
							Description(description).
							Value(&yesFlag),
					),
				).WithTheme(views.GetCustomTheme())

				err := form.Run()
				if err != nil {
					return err
				}

				if yesFlag {
					err := RemoveTargetConfigWorkspaces(ctx, apiClient, selectedConfigName)
					if err != nil {
						return err
					}
				} else {
					fmt.Println("Proceeding with target config removal without deleting workspaces.")
				}
			}
		}

		res, err := apiClient.TargetConfigAPI.RemoveTargetConfig(ctx, selectedConfigName).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		views.RenderInfoMessageBold(fmt.Sprintf("Target config %s removed successfully", selectedConfigName))
		return nil
	},
}

func init() {
	removeCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Confirm deletion of all workspaces without prompt")
}

func RemoveTargetConfigWorkspaces(ctx context.Context, client *apiclient.APIClient, targetConfig string) error {
	workspaceList, res, err := client.WorkspaceAPI.ListWorkspaces(ctx).Execute()
	if err != nil {
		return apiclient_util.HandleErrorResponse(res, err)
	}

	for _, workspace := range workspaceList {
		if workspace.TargetConfig != targetConfig {
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
