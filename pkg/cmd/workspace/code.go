// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"
	"errors"
	"fmt"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/jetbrains"
	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	workspace_util "github.com/daytonaio/daytona/pkg/cmd/workspace/util"
	"github.com/daytonaio/daytona/pkg/ide"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var CodeCmd = &cobra.Command{
	Use:     "code [WORKSPACE] [PROJECT]",
	Short:   "Open a workspace in your preferred IDE",
	Args:    cobra.RangeArgs(0, 2),
	Aliases: []string{"open"},
	GroupID: util.WORKSPACE_GROUP,
	Run: func(cmd *cobra.Command, args []string) {
		c, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		ctx := context.Background()
		var workspaceId string
		var projectName string
		var ideId string
		var workspace *apiclient.WorkspaceDTO

		activeProfile, err := c.GetActiveProfile()
		if err != nil {
			log.Fatal(err)
		}

		ideId = c.DefaultIdeId

		apiClient, err := apiclient_util.GetApiClient(&activeProfile)
		if err != nil {
			log.Fatal(err)
		}

		if len(args) == 0 {
			workspaceList, res, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Verbose(true).Execute()
			if err != nil {
				log.Fatal(apiclient_util.HandleErrorResponse(res, err))
			}

			workspace = selection.GetWorkspaceFromPrompt(workspaceList, "Open")
			if workspace == nil {
				return
			}
			workspaceId = *workspace.Id
		} else {
			workspace, err = apiclient_util.GetWorkspace(args[0])
			if err != nil {
				log.Fatal(err)
			}
			workspaceId = *workspace.Id
		}

		if len(args) == 0 || len(args) == 1 {
			selectedProject, err := selectWorkspaceProject(workspaceId, &activeProfile)
			if err != nil {
				log.Fatal(err)
			}
			if selectedProject == nil {
				return
			}
			projectName = *selectedProject.Name
		}

		if len(args) == 2 {
			projectName = args[1]
		}

		if ideFlag != "" {
			ideId = ideFlag
		}

		ideList := config.GetIdeList()
		ideName := ""
		for _, ide := range ideList {
			if ide.Id == ideId {
				ideName = ide.Name
				break
			}
		}

		providerMetadata := ""
		if workspace.Info != nil {
			for _, project := range workspace.Info.Projects {
				if *project.Name == projectName {
					if !project.GetIsRunning() {
						views.RenderInfoMessage(fmt.Sprintf("Project '%s' from workspace '%s' is not in running state", projectName, *workspace.Name))
						return
					}
					if project.ProviderMetadata == nil {
						log.Fatal(errors.New("project provider metadata is missing"))
					}
					providerMetadata = *project.ProviderMetadata
					break
				}
			}
		}

		if !workspace_util.IsProjectRunning(workspace, projectName) {
			views.RenderInfoMessage(fmt.Sprintf("Project '%s' from workspace '%s' is not in running state", projectName, *workspace.Name))
			return
		}

		views.RenderInfoMessage(fmt.Sprintf("Opening the project '%s' from workspace '%s' in %s", projectName, *workspace.Name, ideName))

		err = openIDE(ideId, activeProfile, workspaceId, projectName, providerMetadata)
		if err != nil {
			log.Fatal(err)
		}
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) >= 2 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		if len(args) == 1 {
			return getProjectNameCompletions(cmd, args, toComplete)
		}

		return getWorkspaceNameCompletions()
	},
}

func selectWorkspaceProject(workspaceId string, profile *config.Profile) (*apiclient.Project, error) {
	ctx := context.Background()

	apiClient, err := apiclient_util.GetApiClient(profile)
	if err != nil {
		return nil, err
	}

	wsInfo, res, err := apiClient.WorkspaceAPI.GetWorkspace(ctx, workspaceId).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	if len(wsInfo.Projects) > 1 {
		selectedProject := selection.GetProjectFromPrompt(wsInfo.Projects, "Open")
		if selectedProject == nil {
			return nil, nil
		}
		return selectedProject, nil
	} else if len(wsInfo.Projects) == 1 {
		return &wsInfo.Projects[0], nil
	}

	return nil, errors.New("no projects found in workspace")
}

func openIDE(ideId string, activeProfile config.Profile, workspaceId string, projectName string, projectProviderMetadata string) error {
	switch ideId {
	case "vscode":
		return ide.OpenVSCode(activeProfile, workspaceId, projectName, projectProviderMetadata)
	case "ssh":
		return ide.OpenTerminalSsh(activeProfile, workspaceId, projectName)
	case "browser":
		return ide.OpenBrowserIDE(activeProfile, workspaceId, projectName, projectProviderMetadata)
	default:
		_, ok := jetbrains.GetIdes()[jetbrains.Id(ideId)]
		if ok {
			return ide.OpenJetbrainsIDE(activeProfile, ideId, workspaceId, projectName)
		}
	}

	return errors.New("invalid IDE. Please choose one by running `daytona ide`")
}

var ideFlag string

func init() {
	CodeCmd.Flags().StringVarP(&ideFlag, "ide", "i", "", "Specify the IDE ('vscode' or 'browser')")
}
