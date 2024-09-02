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

type WorkspaceState string

const (
	WORKSPACE_STATUS_RUNNING WorkspaceState = "Running"
	WORKSPACE_STATUS_STOPPED WorkspaceState = "Unavailable"
)

var startProjectFlag string
var allFlag bool

var StartCmd = &cobra.Command{
	Use:     "start [WORKSPACE]",
	Short:   "Start a workspace",
	Args:    cobra.RangeArgs(0, 1),
	GroupID: util.WORKSPACE_GROUP,
	Run: func(cmd *cobra.Command, args []string) {
		s := spinner.New("Processing...")
		s.Start()
		defer s.Success()

		var workspaceId string
		var message string

		if allFlag {
			s.UpdateMessage("Starting all workspaces...")
			err := startAllWorkspaces(s)
			if err != nil {
				s.Error("Failed to start all workspaces.")
				log.Fatal(err)
			}
			s.UpdateMessage("All workspaces started successfully.")
			return
		}

		ctx := context.Background()

		apiClient, err := apiclient.GetApiClient(nil)
		if err != nil {
			s.Error("Failed to create API client.")
			log.Fatal(err)
		}

		if len(args) == 0 {
			if startProjectFlag != "" {
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

			workspace := selection.GetWorkspaceFromPrompt(workspaceList, "Start")
			if workspace == nil {
				s.Error("No workspace selected.")
				return
			}
			workspaceId = workspace.Name
		} else {
			workspaceId = args[0]
		}

		if startProjectFlag == "" {
			s.UpdateMessage(fmt.Sprintf("Starting workspace '%s'...", workspaceId))
			res, err := apiClient.WorkspaceAPI.StartWorkspace(ctx, workspaceId).Execute()
			if err != nil {
				s.Error(fmt.Sprintf("Failed to start workspace '%s'.", workspaceId))
				log.Fatal(apiclient.HandleErrorResponse(res, err))
			}
			message = fmt.Sprintf("Workspace '%s' started successfully.", workspaceId)
			s.UpdateMessage(message)
		} else {
			s.UpdateMessage(fmt.Sprintf("Starting project '%s' in workspace '%s'...", startProjectFlag, workspaceId))
			res, err := apiClient.WorkspaceAPI.StartProject(ctx, workspaceId, startProjectFlag).Execute()
			if err != nil {
				s.Error(fmt.Sprintf("Failed to start project '%s' in workspace '%s'.", startProjectFlag, workspaceId))
				log.Fatal(apiclient.HandleErrorResponse(res, err))
			}
			message = fmt.Sprintf("Project '%s' in workspace '%s' started successfully.", startProjectFlag, workspaceId)
			s.UpdateMessage(message)
		}

	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		return getAllWorkspacesByState(WORKSPACE_STATUS_STOPPED)
	},
}

func init() {
	StartCmd.PersistentFlags().StringVarP(&startProjectFlag, "project", "p", "", "Start a single project in the workspace (project name)")
	StartCmd.PersistentFlags().BoolVarP(&allFlag, "all", "a", false, "Start all workspaces")

	err := StartCmd.RegisterFlagCompletionFunc("project", getProjectNameCompletions)
	if err != nil {
		log.Error("failed to register completion function: ", err)
	}
}

func startAllWorkspaces(s *spinner.Spinner) error {
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
		s.UpdateMessage(fmt.Sprintf("Starting workspace '%s'...", workspace.Name))
		s.Successf(fmt.Sprintf("Workspace '%s' started successfully", workspace.Name))
		s.Start()
		res, err := apiClient.WorkspaceAPI.StartWorkspace(ctx, workspace.Id).Execute()
		if err != nil {
			log.Errorf("Failed to start workspace %s: %v", workspace.Name, apiclient.HandleErrorResponse(res, err))
			continue
		}
	}
	return nil
}

func getProjectNameCompletions(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	ctx := context.Background()
	apiClient, err := apiclient.GetApiClient(nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveDefault
	}

	workspaceId := args[0]
	workspace, _, err := apiClient.WorkspaceAPI.GetWorkspace(ctx, workspaceId).Execute()
	if err != nil {
		return nil, cobra.ShellCompDirectiveDefault
	}

	var choices []string
	for _, project := range workspace.Projects {
		choices = append(choices, project.Name)
	}
	return choices, cobra.ShellCompDirectiveDefault
}

func getWorkspaceNameCompletions() ([]string, cobra.ShellCompDirective) {
	ctx := context.Background()
	apiClient, err := apiclient.GetApiClient(nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	workspaceList, _, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Execute()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var choices []string
	for _, v := range workspaceList {
		choices = append(choices, v.Name)
	}

	return choices, cobra.ShellCompDirectiveNoFileComp
}

func getAllWorkspacesByState(state WorkspaceState) ([]string, cobra.ShellCompDirective) {
	ctx := context.Background()
	apiClient, err := apiclient.GetApiClient(nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	workspaceList, _, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Execute()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var choices []string
	for _, workspace := range workspaceList {
		for _, project := range workspace.Info.Projects {
			if state == WORKSPACE_STATUS_RUNNING && project.IsRunning {
				choices = append(choices, workspace.Name)
				break
			}
			if state == WORKSPACE_STATUS_STOPPED && !project.IsRunning {
				choices = append(choices, workspace.Name)
				break
			}
		}
	}

	return choices, cobra.ShellCompDirectiveNoFileComp
}
