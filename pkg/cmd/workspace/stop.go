// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"
	"fmt"
	"time"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/workspace/common"
	"github.com/daytonaio/daytona/pkg/views"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var StopCmd = &cobra.Command{
	Use:     "stop [WORKSPACE]",
	Short:   "Stop a workspace",
	Args:    cobra.RangeArgs(0, 1),
	GroupID: util.TARGET_GROUP,
	RunE: func(cmd *cobra.Command, args []string) error {
		var selectedWorkspaces []apiclient.WorkspaceDTO

		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		if allFlag {
			return stopAllWorkspaces()
		}

		if len(args) == 0 {
			workspaceList, res, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			if len(workspaceList) == 0 {
				views_util.NotifyEmptyWorkspaceList(true)
				return nil
			}
			selectedWorkspace := selection.GetWorkspaceFromPrompt(workspaceList, "Stop")
			if selectedWorkspace == nil {
				return nil
			}
			selectedWorkspaces = append(selectedWorkspaces, *selectedWorkspace)
		} else {
			workspace, err := apiclient_util.GetWorkspace(args[0], true)
			if err != nil {
				return err
			}

			selectedWorkspaces = append(selectedWorkspaces, *workspace)
		}

		if len(selectedWorkspaces) == 1 {
			workspace := selectedWorkspaces[0]

			err = StopWorkspace(apiClient, workspace)
			if err != nil {
				return err
			}

			views.RenderInfoMessage(fmt.Sprintf("Workspace '%s' stopped successfully", workspace.Name))
		} else {
			for _, ws := range selectedWorkspaces {
				err := StopWorkspace(apiClient, ws)
				if err != nil {
					log.Errorf("Failed to stop workspace %s: %v\n\n", ws.Name, err)
					continue
				}
				views.RenderInfoMessage(fmt.Sprintf("- Workspace '%s' stopped successfully", ws.Name))
			}
		}
		return nil
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return common.GetAllWorkspacesByState(common.WORKSPACE_STATE_RUNNING)
	},
}

func init() {
	StopCmd.PersistentFlags().BoolVarP(&allFlag, "all", "a", false, "Stop all targets")
	StopCmd.PersistentFlags().BoolVarP(&yesFlag, "yes", "y", false, "Automatically confirm any prompts")
}

func stopAllWorkspaces() error {
	ctx := context.Background()
	apiClient, err := apiclient_util.GetApiClient(nil)
	if err != nil {
		return err
	}

	workspaceList, res, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Execute()
	if err != nil {
		return apiclient_util.HandleErrorResponse(res, err)
	}

	for _, workspace := range workspaceList {
		err := StopWorkspace(apiClient, workspace)
		if err != nil {
			log.Errorf("Failed to stop workspace %s: %v\n\n", workspace.Name, err)
			continue
		}

		views.RenderInfoMessage(fmt.Sprintf("- Workspace '%s' stopped successfully", workspace.Name))
	}
	return nil
}

func StopWorkspace(apiClient *apiclient.APIClient, workspace apiclient.WorkspaceDTO) error {
	ctx := context.Background()
	timeFormat := time.Now().Format("2006-01-02 15:04:05")
	from, err := time.Parse("2006-01-02 15:04:05", timeFormat)
	if err != nil {
		return err
	}

	c, err := config.GetConfig()
	if err != nil {
		return err
	}

	activeProfile, err := c.GetActiveProfile()
	if err != nil {
		return err
	}

	logsContext, stopLogs := context.WithCancel(context.Background())
	go apiclient_util.ReadWorkspaceLogs(logsContext, apiclient_util.ReadLogParams{
		Id:            workspace.Id,
		Label:         &workspace.Name,
		ActiveProfile: activeProfile,
		Index:         util.Pointer(0),
		Follow:        util.Pointer(true),
		From:          &from,
	})

	res, err := apiClient.WorkspaceAPI.StopWorkspace(ctx, workspace.Id).Execute()
	if err != nil {
		stopLogs()
		return apiclient_util.HandleErrorResponse(res, err)
	}
	time.Sleep(100 * time.Millisecond)
	stopLogs()
	return nil
}
