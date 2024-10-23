// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
	"github.com/spf13/cobra"
)

var restartProjectFlag string

var RestartCmd = &cobra.Command{
	Use:     "restart [WORKSPACE]",
	Short:   "Restart a workspace",
	Args:    cobra.RangeArgs(0, 1),
	GroupID: util.WORKSPACE_GROUP,
	RunE: func(cmd *cobra.Command, args []string) error {
		var workspaceId string

		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		if len(args) == 0 {
			if restartProjectFlag != "" {
				err := cmd.Help()
				if err != nil {
					return err
				}
				return nil
			}

			workspaceList, res, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			workspace := selection.GetWorkspaceFromPrompt(workspaceList, "Restart")
			if workspace == nil {
				return nil
			}
			workspaceId = workspace.Name
		} else {
			workspaceId = args[0]
		}

		err = RestartWorkspace(apiClient, workspaceId, restartProjectFlag)
		if err != nil {
			return err
		}
		if restartProjectFlag != "" {
			views.RenderInfoMessage(fmt.Sprintf("Project '%s' from workspace '%s' successfully restarted", restartProjectFlag, workspaceId))
		} else {
			views.RenderInfoMessage(fmt.Sprintf("Workspace '%s' successfully restarted", workspaceId))
		}
		return nil
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return getAllWorkspacesByState(WORKSPACE_STATUS_RUNNING)
	},
}

func init() {
	RestartCmd.Flags().StringVarP(&restartProjectFlag, "project", "p", "", "Restart a single project in the workspace (project name)")
}

func RestartWorkspace(apiClient *apiclient.APIClient, workspaceId, projectName string) error {
	err := StopWorkspace(apiClient, workspaceId, projectName)
	if err != nil {
		return err
	}
	return StartWorkspace(apiClient, workspaceId, projectName)
}
