// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"
	"errors"
	"fmt"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/jetbrains"
	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/server"
	"github.com/daytonaio/daytona/pkg/ide"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
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
	Run: func(cmd *cobra.Command, args []string) {
		c, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		ctx := context.Background()
		var workspaceId string
		var projectName string
		var ideId string
		var workspace *serverapiclient.WorkspaceDTO

		activeProfile, err := c.GetActiveProfile()
		if err != nil {
			log.Fatal(err)
		}

		ideId = c.DefaultIdeId

		apiClient, err := server.GetApiClient(&activeProfile)
		if err != nil {
			log.Fatal(err)
		}

		if len(args) == 0 {
			workspaceList, res, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Execute()
			if err != nil {
				log.Fatal(apiclient.HandleErrorResponse(res, err))
			}

			workspace = selection.GetWorkspaceFromPrompt(workspaceList, "Open")
			if workspace == nil {
				return
			}
			workspaceId = *workspace.Id
		} else {
			workspace, err = server.GetWorkspace(args[0])
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
			projectName = *selectedProject
		}

		if len(args) == 2 {
			projectName = args[1]
		}

		if ideFlag != "" {
			ideId = ideFlag
		}

		views.RenderInfoMessage(fmt.Sprintf("Opening the project '%s' from workspace '%s' in your preferred IDE.", projectName, *workspace.Name))

		err = openIDE(ideId, activeProfile, workspaceId, projectName)
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

func selectWorkspaceProject(workspaceId string, profile *config.Profile) (*string, error) {
	ctx := context.Background()

	apiClient, err := server.GetApiClient(profile)
	if err != nil {
		return nil, err
	}

	wsInfo, res, err := apiClient.WorkspaceAPI.GetWorkspace(ctx, workspaceId).Execute()
	if err != nil {
		return nil, apiclient.HandleErrorResponse(res, err)
	}

	if len(wsInfo.Projects) > 1 {
		selectedProject := selection.GetProjectFromPrompt(wsInfo.Projects, "Open")
		if selectedProject == nil {
			return nil, nil
		}
		return selectedProject.Name, nil
	} else if len(wsInfo.Projects) == 1 {
		return wsInfo.Projects[0].Name, nil
	}

	return nil, errors.New("no projects found in workspace")
}

func openIDE(ideId string, activeProfile config.Profile, workspaceId string, projectName string) error {
	switch ideId {
	case "vscode":
		return ide.OpenVSCode(activeProfile, workspaceId, projectName)
	case "ssh":
		return ide.OpenTerminalSsh(activeProfile, workspaceId, projectName)
	case "browser":
		return ide.OpenBrowserIDE(activeProfile, workspaceId, projectName)
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
