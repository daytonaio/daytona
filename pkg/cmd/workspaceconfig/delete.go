// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaceconfig

import (
	"context"
	"fmt"

	"github.com/charmbracelet/huh"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/selection"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var allFlag bool
var yesFlag bool
var forceFlag bool

var workspaceConfigDeleteCmd = &cobra.Command{
	Use:     "delete",
	Aliases: []string{"remove", "rm"},
	Short:   "Delete a workspace config",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var selectedWorkspaceConfig *apiclient.WorkspaceConfig
		var selectedWorkspaceConfigName string

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		if allFlag {
			if !yesFlag {
				form := huh.NewForm(
					huh.NewGroup(
						huh.NewConfirm().
							Title("Delete all workspace configs?").
							Description("Are you sure you want to delete all workspace configs?").
							Value(&yesFlag),
					),
				).WithTheme(views.GetCustomTheme())

				err := form.Run()
				if err != nil {
					return err
				}

				if !yesFlag {
					fmt.Println("Operation canceled.")
					return nil
				}
			}

			workspaceConfigs, res, err := apiClient.WorkspaceConfigAPI.ListWorkspaceConfigs(context.Background()).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			if len(workspaceConfigs) == 0 {
				views_util.NotifyEmptyWorkspaceConfigList(false)
				return nil
			}

			for _, workspaceConfig := range workspaceConfigs {
				selectedWorkspaceConfigName = workspaceConfig.Name
				res, err := apiClient.WorkspaceConfigAPI.DeleteWorkspaceConfig(context.Background(), selectedWorkspaceConfigName).Execute()
				if err != nil {
					log.Error(apiclient_util.HandleErrorResponse(res, err))
					continue
				}
				views.RenderInfoMessage("Deleted workspace config: " + selectedWorkspaceConfigName)
			}
			return nil
		}

		if len(args) == 0 {
			workspaceConfigs, res, err := apiClient.WorkspaceConfigAPI.ListWorkspaceConfigs(context.Background()).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			if len(workspaceConfigs) == 0 {
				views.RenderInfoMessage("No workspace configs found")
				return nil
			}

			selectedWorkspaceConfig = selection.GetWorkspaceConfigFromPrompt(workspaceConfigs, 0, false, false, "Delete")
			if selectedWorkspaceConfig == nil {
				return nil
			}
			selectedWorkspaceConfigName = selectedWorkspaceConfig.Name
		} else {
			selectedWorkspaceConfigName = args[0]
		}

		res, err := apiClient.WorkspaceConfigAPI.DeleteWorkspaceConfig(context.Background(), selectedWorkspaceConfigName).Force(forceFlag).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		views.RenderInfoMessage("Workspace config deleted successfully")
		return nil
	},
}

func init() {
	workspaceConfigDeleteCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Delete all workspace configs")
	workspaceConfigDeleteCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Confirm deletion without prompt")
	workspaceConfigDeleteCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Force delete prebuild")
}
