// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/workspace/common"
	"github.com/daytonaio/daytona/pkg/views"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"

	"github.com/spf13/cobra"
)

var RestartCmd = &cobra.Command{
	Use:     "restart [WORKSPACE]",
	Short:   "Restart a workspace",
	Args:    cobra.RangeArgs(0, 1),
	GroupID: util.TARGET_GROUP,
	RunE: func(cmd *cobra.Command, args []string) error {
		var workspace *apiclient.WorkspaceDTO

		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		workspaceList, res, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		if len(workspaceList) == 0 {
			views_util.NotifyEmptyWorkspaceList(true)
			return nil
		}

		if len(args) == 0 {
			workspace := selection.GetWorkspaceFromPrompt(workspaceList, "Restart")
			if workspace == nil {
				return nil
			}
		} else {
			workspace, err = apiclient_util.GetWorkspace(args[0], false)
			if err != nil {
				return err
			}
		}

		err = RestartWorkspace(apiClient, *workspace)
		if err != nil {
			return err
		}

		views.RenderInfoMessage(fmt.Sprintf("Workspace '%s' restarted successfully", workspace.Name))

		return nil
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return common.GetAllWorkspacesByState(common.WORKSPACE_STATE_RUNNING)
	},
}

func RestartWorkspace(apiClient *apiclient.APIClient, workspace apiclient.WorkspaceDTO) error {
	err := StopWorkspace(apiClient, workspace)
	if err != nil {
		return err
	}
	return StartWorkspace(apiClient, workspace)
}
