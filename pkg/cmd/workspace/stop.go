// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
	"github.com/leaanthony/spinner"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var stopProjectFlag string

var StopCmd = &cobra.Command{
	Use:     "stop [WORKSPACE]",
	Short:   "Stop a workspace",
	GroupID: util.WORKSPACE_GROUP,
	Args:    cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		s := spinner.New("Processing...")
		s.Start()
		defer s.Success()

		var workspaceId string
		var message string

		if allFlag {
			s.UpdateMessage("Stopping all workspaces...")
			err := stopAllWorkspaces(s)
			if err != nil {
				s.Error("Failed to stop all workspaces.")
				log.Fatal(err)
				return
			}
			s.UpdateMessage("All workspaces stopped successfully.")
			return
		}

		ctx := context.Background()

		apiClient, err := apiclient.GetApiClient(nil)
		if err != nil {
			s.Error("Failed to create API client.")
			log.Fatal(err)
		}

		if len(args) == 0 {
			if stopProjectFlag != "" {
				err := cmd.Help()
				if err != nil {
					s.Error("Failed to display help.")
					log.Fatal(err)
				}
				return
			}
			s.UpdateMessage("Fetching workspace list...")
			workspaceList, res, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Execute()
			if err != nil {
				s.Error("Failed to retrieve workspaces.")
				log.Fatal(apiclient.HandleErrorResponse(res, err))
			}

			if len(workspaceList) == 0 {
				s.Error("No workspaces available to stop.")
				return
			}

			s.UpdateMessage("Selecting a workspace to stop...")
			workspace := selection.GetWorkspaceFromPrompt(workspaceList, "Stop")
			if workspace == nil {
				s.Error("No workspace selected.")
				return
			}
			workspaceId = workspace.Name
		} else {
			workspaceId = args[0]
		}

		if stopProjectFlag == "" {
			message = fmt.Sprintf("Workspace '%s' is stopping...", workspaceId)
			s.UpdateMessage(message)
			res, err := apiClient.WorkspaceAPI.StopWorkspace(ctx, workspaceId).Execute()
			if err != nil {
				s.Error(fmt.Sprintf("Failed to stop workspace '%s'.", workspaceId))
				log.Fatal(apiclient.HandleErrorResponse(res, err))
			}
			message = fmt.Sprintf("Workspace '%s' stopped successfully.", workspaceId)
			s.UpdateMessage(message)
		} else {
			message = fmt.Sprintf("Project '%s' from workspace '%s' is stopping...", stopProjectFlag, workspaceId)
			s.UpdateMessage(message)
			res, err := apiClient.WorkspaceAPI.StopProject(ctx, workspaceId, stopProjectFlag).Execute()
			if err != nil {
				s.Error(fmt.Sprintf("Failed to stop project '%s' in workspace '%s'.", stopProjectFlag, workspaceId))
				log.Fatal(apiclient.HandleErrorResponse(res, err))
			}
			message = fmt.Sprintf("Project '%s' in workspace '%s' stopped successfully.", stopProjectFlag, workspaceId)
			s.UpdateMessage(message)
		}

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

func stopAllWorkspaces(s *spinner.Spinner) error {
	ctx := context.Background()
	apiClient, err := apiclient.GetApiClient(nil)
	if err != nil {
		return err
	}

	workspaceList, res, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Execute()
	if err != nil {
		return apiclient.HandleErrorResponse(res, err)
	}

	for _, workspace := range workspaceList {
		s.UpdateMessage(fmt.Sprintf("Stopping workspace '%s'...", workspace.Name))
		s.Successf(fmt.Sprintf("Workspace '%s' stopped successfully", workspace.Name))
		s.Start()
		res, err := apiClient.WorkspaceAPI.StopWorkspace(ctx, workspace.Id).Execute()
		if err != nil {
			log.Errorf("Failed to stop workspace %s: %v", workspace.Name, apiclient.HandleErrorResponse(res, err))
			continue
		}

	}
	return nil
}
