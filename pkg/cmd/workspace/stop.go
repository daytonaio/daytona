// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"
	"fmt"
	"os"

	internal_util "github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/server"
	"github.com/daytonaio/daytona/pkg/views/util"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var stopProjectFlag string

var StopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop a workspace",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		if allFlag {
			err := stopAllWorkspaces()
			if err != nil {
				log.Fatal(err)
			}
			return
		}

		ctx := context.Background()
		var workspaceId string

		apiClient, err := server.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		if internal_util.WorkspaceMode() {
			workspaceId = os.Getenv("DAYTONA_WS_ID")
		} else if len(args) == 0 {
			workspaceList, res, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Execute()
			if err != nil {
				log.Fatal(apiclient.HandleErrorResponse(res, err))
			}

			workspace := selection.GetWorkspaceFromPrompt(workspaceList, "stop")
			if workspace == nil {
				return
			}
			workspaceId = *workspace.Name
		} else {
			workspaceId = args[0]
		}

		if stopProjectFlag == "" {
			res, err := apiClient.WorkspaceAPI.StopWorkspace(ctx, workspaceId).Execute()
			if err != nil {
				log.Fatal(apiclient.HandleErrorResponse(res, err))
			}
		} else {
			res, err := apiClient.WorkspaceAPI.StopProject(ctx, workspaceId, stopProjectFlag).Execute()
			if err != nil {
				log.Fatal(apiclient.HandleErrorResponse(res, err))
			}
		}

		util.RenderInfoMessage(fmt.Sprintf("Workspace %s successfully stopped", workspaceId))
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) >= 1 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		return getAllWorkspacesByState(WORKSPACE_STATUS_RUNNING)
	},
}

func init() {
	if !internal_util.WorkspaceMode() {
		StopCmd.Use += " [WORKSPACE]"
		StopCmd.Args = cobra.MaximumNArgs(0)
	}

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
			log.Errorf("Failed to stop workspace %s: %v", *workspace.Id, apiclient.HandleErrorResponse(res, err))
			continue
		}
		fmt.Printf("Workspace %s successfully stopped\n", *workspace.Id)
	}
	return nil
}
