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
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var stopProjectFlag string

var StopCmd = &cobra.Command{
	Use:     "stop [WORKSPACE]",
	Short:   "Stop a workspace",
	GroupID: util.WORKSPACE_GROUP,
	Args:    cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var workspaceId string

		if allFlag {
			return stopAllWorkspaces()
		}

		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		if len(args) == 0 {
			if stopProjectFlag != "" {
				return cmd.Help()
			}
			workspaceList, res, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			workspace := selection.GetWorkspaceFromPrompt(workspaceList, "Stop")
			if workspace == nil {
				return nil
			}
			workspaceId = workspace.Name
		} else {
			workspaceId = args[0]
		}

		err = StopWorkspace(apiClient, workspaceId, stopProjectFlag)
		if err != nil {
			return err
		}
		if stopProjectFlag != "" {
			views.RenderInfoMessage(fmt.Sprintf("Project '%s' from workspace '%s' successfully stopped", stopProjectFlag, workspaceId))
		} else {
			views.RenderInfoMessage(fmt.Sprintf("Workspace '%s' successfully stopped", workspaceId))
		}
		return nil
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) >= 1 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		return getAllWorkspacesByState(WORKSPACE_STATUS_RUNNING)
	},
}

func init() {
	StopCmd.Flags().StringVarP(&stopProjectFlag, "project", "p", "", "Stop a single project in the workspace (project name)")
	StopCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Stop all workspaces")
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
		err := StopWorkspace(apiClient, workspace.Name, "")
		if err != nil {
			log.Errorf("Failed to stop workspace %s: %v\n\n", workspace.Name, err)
			continue
		}
		views.RenderInfoMessage(fmt.Sprintf("- Workspace '%s' successfully stopped", workspace.Name))
	}
	return nil
}

func StopWorkspace(apiClient *apiclient.APIClient, workspaceId, projectName string) error {
	ctx := context.Background()
	var message string
	var stopFunc func() error

	if projectName == "" {
		message = fmt.Sprintf("Workspace '%s' is stopping", workspaceId)
		stopFunc = func() error {
			res, err := apiClient.WorkspaceAPI.StopWorkspace(ctx, workspaceId).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}
			return nil
		}
	} else {
		message = fmt.Sprintf("Project '%s' from workspace '%s' is stopping", projectName, workspaceId)
		stopFunc = func() error {
			res, err := apiClient.WorkspaceAPI.StopProject(ctx, workspaceId, projectName).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}
			return nil
		}
	}

	err := views_util.WithInlineSpinner(message, stopFunc)
	if err != nil {
		return err
	}

	return nil
}
