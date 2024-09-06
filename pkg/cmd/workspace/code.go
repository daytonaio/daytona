// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/jetbrains"
	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	workspace_util "github.com/daytonaio/daytona/pkg/cmd/workspace/util"
	"github.com/daytonaio/daytona/pkg/ide"
	"github.com/daytonaio/daytona/pkg/server/workspaces"
	"github.com/daytonaio/daytona/pkg/views"
	ide_views "github.com/daytonaio/daytona/pkg/views/ide"
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
			workspaceId = workspace.Id
		} else {
			workspace, err = apiclient_util.GetWorkspace(url.PathEscape(args[0]))
			if err != nil {
				if strings.Contains(err.Error(), workspaces.ErrWorkspaceNotFound.Error()) {
					log.Debug(err)
					log.Fatal("Workspace not found. You can see all workspace names by running the command `daytona list`")
				}
				log.Fatal(err)
			}
			workspaceId = workspace.Id
		}

		if len(args) == 0 || len(args) == 1 {
			selectedProject, err := selectWorkspaceProject(workspaceId, &activeProfile)
			if err != nil {
				log.Fatal(err)
			}
			if selectedProject == nil {
				return
			}
			projectName = selectedProject.Name
		}

		if len(args) == 2 {
			projectName = args[1]
		}

		if ideFlag != "" {
			ideId = ideFlag
		}

		if !workspace_util.IsProjectRunning(workspace, projectName) {
			wsRunningStatus, err := AutoStartWorkspace(autoStartFlag, workspace.Name, projectName)
			if err != nil {
				log.Fatal(err)
			}
			if !wsRunningStatus {
				return
			}
		}

		providerMetadata := ""
		if ideId != "ssh" {
			providerMetadata, err = workspace_util.GetProjectProviderMetadata(workspace, projectName)
			if err != nil {
				log.Fatal(err)
			}
		}

		ideList := config.GetIdeList()
		ide_views.RenderIdeOpeningMessage(workspace.Name, projectName, ideId, ideList)
		if err := openIDE(ideId, activeProfile, workspaceId, projectName, providerMetadata); err != nil {
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
	case "cursor":
		return ide.OpenCursor(activeProfile, workspaceId, projectName, projectProviderMetadata)
	default:
		_, ok := jetbrains.GetIdes()[jetbrains.Id(ideId)]
		if ok {
			return ide.OpenJetbrainsIDE(activeProfile, ideId, workspaceId, projectName)
		}
	}

	return errors.New("invalid IDE. Please choose one by running `daytona ide`")
}

var ideFlag string
var autoStartFlag bool

func init() {
	ideList := config.GetIdeList()
	ids := make([]string, len(ideList))
	for i, ide := range ideList {
		ids[i] = ide.Id
	}
	ideListStr := strings.Join(ids, ", ")
	CodeCmd.Flags().StringVarP(&ideFlag, "ide", "i", "", fmt.Sprintf("Specify the IDE (%s)", ideListStr))
	CodeCmd.Flags().BoolVarP(&autoStartFlag, "auto-start", "a", false, "Automatically start the project if it is not running")
}

func AutoStartWorkspace(autoStartFlag bool, workspaceId string, projectName string) (bool, error) {
	if !autoStartFlag {
		if !ide_views.RunStartWorkspaceForm(workspaceId) {
			return false, nil
		}
	}

	apiClient, err := apiclient_util.GetApiClient(nil)
	if err != nil {
		return false, err
	}

	views.RenderInfoMessage(fmt.Sprintf("Project '%s' from workspace '%s' is starting", projectName, workspaceId))

	ctx := context.Background()
	res, err := apiClient.WorkspaceAPI.StartProject(ctx, workspaceId, projectName).Execute()
	if err != nil {
		return false, apiclient_util.HandleErrorResponse(res, err)
	}

	return true, nil
}
