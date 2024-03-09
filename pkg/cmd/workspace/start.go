// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/server"
	"github.com/daytonaio/daytona/pkg/views/util"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var startProjectFlag string
var allFlag bool

var StartCmd = &cobra.Command{
	Use:   "start [WORKSPACE]",
	Short: "Start a workspace",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if allFlag {
			err := startAllWorkspaces()
			if err != nil {
				log.Fatal(err)
			}
		}

		ctx := context.Background()
		var workspaceId string

		apiClient, err := server.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		if len(args) == 0 {
			workspaceList, res, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Execute()
			if err != nil {
				log.Fatal(apiclient.HandleErrorResponse(res, err))
			}

			workspace := selection.GetWorkspaceFromPrompt(workspaceList, "start")
			if workspace == nil {
				return
			}
			workspaceId = *workspace.Id
		} else {
			workspaceId = args[0]
		}

		if startProjectFlag == "" {
			res, err := apiClient.WorkspaceAPI.StartWorkspace(ctx, workspaceId).Execute()
			if err != nil {
				log.Fatal(apiclient.HandleErrorResponse(res, err))
			}
		} else {
			res, err := apiClient.WorkspaceAPI.StartProject(ctx, workspaceId, startProjectFlag).Execute()
			if err != nil {
				log.Fatal(apiclient.HandleErrorResponse(res, err))
			}
		}

		util.RenderInfoMessage(fmt.Sprintf("Workspace %s successfully started", workspaceId))
	},
}

func init() {
	StartCmd.PersistentFlags().StringVarP(&startProjectFlag, "project", "p", "", "Start a single project in the workspace (project name)")
	StartCmd.PersistentFlags().BoolVarP(&allFlag, "all", "a", false, "Start all workspaces")
}

func startAllWorkspaces() error {
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
		res, err := apiClient.WorkspaceAPI.StartWorkspace(ctx, *workspace.Id).Execute()
		if err != nil {
			log.Errorf("Failed to start workspace %s: %v", *workspace.Id, apiclient.HandleErrorResponse(res, err))
			continue
		}
		fmt.Printf("Workspace %s successfully started\n", *workspace.Id)
	}
	return nil
}
