// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/server"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var stopProjectFlag string

var StopCmd = &cobra.Command{
	Use:   "stop [WORKSPACE]",
	Short: "Stop a workspace",
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		var workspaceId string
		var message string

		if allFlag {
			err := stopAllWorkspaces()
			if err != nil {
				log.Fatal(err)
			}
			return
		}

		ctx := context.Background()

		apiClient, err := server.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		if len(args) == 0 {
			if stopProjectFlag != "" {
				err := cmd.Help()
				if err != nil {
					log.Fatal(err)
				}
				return
			}
			workspaceList, res, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Execute()
			if err != nil {
				log.Fatal(apiclient.HandleErrorResponse(res, err))
			}

			workspace := selection.GetWorkspaceFromPrompt(workspaceList, "Stop")
			if workspace == nil {
				return
			}
			workspaceId = *workspace.Name
		} else {
			workspaceId = args[0]
		}

		if stopProjectFlag == "" {
			message = fmt.Sprintf("Workspace '%s' stop request submitted", workspaceId)
			res, err := apiClient.WorkspaceAPI.StopWorkspace(ctx, workspaceId).Execute()
			if err != nil {
				log.Fatal(apiclient.HandleErrorResponse(res, err))
			}
		} else {
			message = fmt.Sprintf("Project '%s' from workspace '%s' stop request submitted", stopProjectFlag, workspaceId)
			res, err := apiClient.WorkspaceAPI.StopProject(ctx, workspaceId, stopProjectFlag).Execute()
			if err != nil {
				log.Fatal(apiclient.HandleErrorResponse(res, err))
			}
		}

		views.RenderInfoMessage(message)
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
	apiClient, err := server.GetApiClient(nil)
	if err != nil {
		return err
	}

	workspaceList, res, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Execute()
	if err != nil {
		return apiclient.HandleErrorResponse(res, err)
	}

	for _, workspace := range workspaceList {
		res, err := apiClient.WorkspaceAPI.StopWorkspace(ctx, *workspace.Id).Execute()
		if err != nil {
			log.Errorf("Failed to stop workspace %s: %v", *workspace.Name, apiclient.HandleErrorResponse(res, err))
			continue
		}
		fmt.Printf("Workspace '%s' stop request submitted\n", *workspace.Name)
	}
	return nil
}
