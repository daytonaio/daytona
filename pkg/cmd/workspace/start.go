// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	workspace_util "github.com/daytonaio/daytona/pkg/cmd/workspace/util"
	"github.com/daytonaio/daytona/pkg/views"
	ide_views "github.com/daytonaio/daytona/pkg/views/ide"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
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
	Use:     "start [WORKSPACE]",
	Short:   "Start a workspace",
	Args:    cobra.RangeArgs(0, 1),
	GroupID: util.WORKSPACE_GROUP,
	RunE: func(cmd *cobra.Command, args []string) error {
		var workspaceIdOrName string
		var activeProfile config.Profile
		var ideId string
		var workspaceId string
		var repoUrl string
		var ideList []config.Ide
		projectProviderMetadata := ""

		if allFlag {
			return startAllWorkspaces()
		}

		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		if len(args) == 0 {
			if startProjectFlag != "" {
				return cmd.Help()
			}
			workspaceList, res, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			workspace := selection.GetWorkspaceFromPrompt(workspaceList, "Start")
			if workspace == nil {
				return nil
			}
			workspaceIdOrName = workspace.Name
		} else {
			workspaceIdOrName = args[0]
		}

		if codeFlag {
			c, err := config.GetConfig()
			if err != nil {
				return err
			}

			ideList = config.GetIdeList()

			activeProfile, err = c.GetActiveProfile()
			if err != nil {
				return err
			}
			ideId = c.DefaultIdeId

			wsInfo, res, err := apiClient.WorkspaceAPI.GetWorkspace(ctx, workspaceIdOrName).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}
			workspaceId = wsInfo.Id
			if startProjectFlag == "" {
				startProjectFlag = wsInfo.Projects[0].Name
				repoUrl = wsInfo.Projects[0].Repository.Url
			} else {
				for _, project := range wsInfo.Projects {
					if project.Name == startProjectFlag {
						repoUrl = project.Repository.Url
						break
					}
				}
			}
			if ideId != "ssh" {
				projectProviderMetadata, err = workspace_util.GetProjectProviderMetadata(wsInfo, wsInfo.Projects[0].Name)
				if err != nil {
					return err
				}
			}
		}

		err = StartWorkspace(apiClient, workspaceIdOrName, startProjectFlag)
		if err != nil {
			return err
		}

		gpgKey, err := GetGitProviderGpgKey(apiClient, ctx, repoUrl)
		if err != nil {
			log.Warn(err)
		}

		if startProjectFlag == "" {
			views.RenderInfoMessage(fmt.Sprintf("Workspace '%s' started successfully", workspaceIdOrName))
		} else {
			views.RenderInfoMessage(fmt.Sprintf("Project '%s' from workspace '%s' started successfully", startProjectFlag, workspaceIdOrName))

			if codeFlag {
				ide_views.RenderIdeOpeningMessage(workspaceIdOrName, startProjectFlag, ideId, ideList)
				err = openIDE(ideId, activeProfile, workspaceId, startProjectFlag, projectProviderMetadata, yesFlag, gpgKey)
				if err != nil {
					return err
				}
			}
		}
		return nil
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
	StartCmd.PersistentFlags().BoolVarP(&codeFlag, "code", "c", false, "Open the workspace in the IDE after workspace start")
	StartCmd.PersistentFlags().BoolVarP(&yesFlag, "yes", "y", false, "Automatically confirm any prompts")

	err := StartCmd.RegisterFlagCompletionFunc("project", getProjectNameCompletions)
	if err != nil {
		log.Error("failed to register completion function: ", err)
	}
}

func startAllWorkspaces() error {
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
		err := StartWorkspace(apiClient, workspace.Name, "")
		if err != nil {
			log.Errorf("Failed to start workspace %s: %v\n\n", workspace.Name, err)
			continue
		}

		views.RenderInfoMessage(fmt.Sprintf("- Workspace '%s' started successfully", workspace.Name))
	}
	return nil
}

func getProjectNameCompletions(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	ctx := context.Background()
	apiClient, err := apiclient_util.GetApiClient(nil)
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
	apiClient, err := apiclient_util.GetApiClient(nil)
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
	apiClient, err := apiclient_util.GetApiClient(nil)
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

func StartWorkspace(apiClient *apiclient.APIClient, workspaceId, projectName string) error {
	ctx := context.Background()
	var message string
	var startFunc func() error

	if projectName == "" {
		message = fmt.Sprintf("Workspace '%s' is starting", workspaceId)
		startFunc = func() error {
			res, err := apiClient.WorkspaceAPI.StartWorkspace(ctx, workspaceId).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}
			return nil
		}
	} else {
		message = fmt.Sprintf("Project '%s' from workspace '%s' is starting", projectName, workspaceId)
		startFunc = func() error {
			res, err := apiClient.WorkspaceAPI.StartProject(ctx, workspaceId, projectName).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}
			return nil
		}
	}

	err := views_util.WithInlineSpinner(message, startFunc)
	if err != nil {
		return err
	}

	return nil
}
