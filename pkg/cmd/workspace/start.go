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

type WorkspaceState string

const (
	WORKSPACE_STATUS_RUNNING WorkspaceState = "Running"
	WORKSPACE_STATUS_STOPPED WorkspaceState = "Unavailable"
)

var startProjectFlag string
var allFlag bool

var StartCmd = &cobra.Command{
	Use:   "start [WORKSPACE]",
	Short: "Start a workspace",
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		var workspaceId string
		var message string

		if allFlag {
			err := startAllWorkspaces()
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
			if startProjectFlag != "" {
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

			workspace := selection.GetWorkspaceFromPrompt(workspaceList, "Start")
			if workspace == nil {
				return
			}
			workspaceId = *workspace.Name
		} else {
			workspaceId = args[0]
		}

		if startProjectFlag == "" {
			message = fmt.Sprintf("Workspace '%s' is starting", workspaceId)
			res, err := apiClient.WorkspaceAPI.StartWorkspace(ctx, workspaceId).Execute()
			if err != nil {
				log.Fatal(apiclient.HandleErrorResponse(res, err))
			}
		} else {
			message = fmt.Sprintf("Project '%s' from workspace '%s' is starting", startProjectFlag, workspaceId)
			res, err := apiClient.WorkspaceAPI.StartProject(ctx, workspaceId, startProjectFlag).Execute()
			if err != nil {
				log.Fatal(apiclient.HandleErrorResponse(res, err))
			}
		}

		views.RenderInfoMessage(message)
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
			log.Errorf("Failed to start workspace %s: %v", *workspace.Name, apiclient.HandleErrorResponse(res, err))
			continue
		}
		fmt.Printf("Workspace '%s' is starting\n", *workspace.Name)
	}
	return nil
}

func getProjectNameCompletions(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	ctx := context.Background()
	apiClient, err := server.GetApiClient(nil)
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
		choices = append(choices, *project.Name)
	}
	return choices, cobra.ShellCompDirectiveDefault
}

func getWorkspaceNameCompletions() ([]string, cobra.ShellCompDirective) {
	ctx := context.Background()
	apiClient, err := server.GetApiClient(nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	workspaceList, _, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Execute()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var choices []string
	for _, v := range workspaceList {
		choices = append(choices, *v.Name)
	}

	return choices, cobra.ShellCompDirectiveNoFileComp
}

func getAllWorkspacesByState(state WorkspaceState) ([]string, cobra.ShellCompDirective) {
	ctx := context.Background()
	apiClient, err := server.GetApiClient(nil)
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
			if state == WORKSPACE_STATUS_RUNNING && *project.IsRunning {
				choices = append(choices, *workspace.Name)
				break
			}
			if state == WORKSPACE_STATUS_STOPPED && !*project.IsRunning {
				choices = append(choices, *workspace.Name)
				break
			}
		}
	}

	return choices, cobra.ShellCompDirectiveNoFileComp
}
