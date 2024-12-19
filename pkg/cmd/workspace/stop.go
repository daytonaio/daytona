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
	"github.com/daytonaio/daytona/pkg/cmd/common"
	cmd_common "github.com/daytonaio/daytona/pkg/cmd/common"
	"github.com/daytonaio/daytona/pkg/views"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var StopCmd = &cobra.Command{
	Use:     "stop [WORKSPACE]...",
	Short:   "Stop a workspace",
	GroupID: util.TARGET_GROUP,
	RunE: func(cmd *cobra.Command, args []string) error {
		var selectedWorkspaces []*apiclient.WorkspaceDTO

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

			selectedWorkspaces = selection.GetWorkspacesFromPrompt(workspaceList, selection.StopActionVerb)
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

			err = StopWorkspace(apiClient, *workspace)
			if err != nil {
				return err
			}

			views.RenderInfoMessage(fmt.Sprintf("Workspace '%s' stopped successfully", workspace.Name))
		} else {
			for _, ws := range selectedWorkspaces {
				err := StopWorkspace(apiClient, *ws)
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
		return common.GetAllWorkspacesByState(apiclient.ResourceStateNameStarted)
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
	go cmd_common.ReadWorkspaceLogs(logsContext, cmd_common.ReadLogParams{
		Id:        workspace.Id,
		Label:     &workspace.Name,
		ServerUrl: activeProfile.Api.Url,
		ApiKey:    activeProfile.Api.Key,
		Index:     util.Pointer(0),
		Follow:    util.Pointer(true),
		From:      &from,
	})

	res, err := apiClient.WorkspaceAPI.StopWorkspace(ctx, workspace.Id).Execute()
	if err != nil {
		stopLogs()
		return apiclient_util.HandleErrorResponse(res, err)
	}

	err = common.AwaitWorkspaceState(workspace.Id, apiclient.ResourceStateNameStopped)
	if err != nil {
		stopLogs()
		return err
	}

	// Ensure reading remaining logs is completed
	time.Sleep(100 * time.Millisecond)

	stopLogs()
	return nil
}
