// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/common"
	"github.com/daytonaio/daytona/pkg/views"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var RestartCmd = &cobra.Command{
	Use:     "restart [WORKSPACE]...",
	Short:   "Restart a workspace",
	GroupID: util.TARGET_GROUP,
	RunE: func(cmd *cobra.Command, args []string) error {
		var selectedWorkspaces []*apiclient.WorkspaceDTO

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
			selectedWorkspaces = selection.GetWorkspacesFromPrompt(workspaceList, selection.RestartActionVerb)
			if selectedWorkspaces == nil {
				return nil
			}
		} else {
			for _, arg := range args {
				workspace, _, err := apiclient_util.GetWorkspace(arg)
				if err != nil {
					log.Error(fmt.Sprintf("[ %s ] : %v", arg, err))
					continue
				}
				selectedWorkspaces = append(selectedWorkspaces, workspace)
			}
		}

		if len(selectedWorkspaces) == 1 {
			workspace := selectedWorkspaces[0]

			err = RestartWorkspace(apiClient, *workspace)
			if err != nil {
				return err
			}

			views.RenderInfoMessage(fmt.Sprintf("Workspace '%s' restarted successfully", workspace.Name))
		} else {
			for _, ws := range selectedWorkspaces {
				err := RestartWorkspace(apiClient, *ws)
				if err != nil {
					log.Errorf("Failed to restart workspace %s: %v\n\n", ws.Name, err)
					continue
				}
				views.RenderInfoMessage(fmt.Sprintf("- Workspace '%s' restarted successfully", ws.Name))
			}
		}

		return nil
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return common.GetAllWorkspacesByState(apiclient.ResourceStateNameStarted)
	},
}

func RestartWorkspace(apiClient *apiclient.APIClient, workspace apiclient.WorkspaceDTO) error {
	err := StopWorkspace(apiClient, workspace)
	if err != nil {
		return err
	}
	return StartWorkspace(apiClient, workspace)
}
