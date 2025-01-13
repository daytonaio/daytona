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
	"github.com/daytonaio/daytona/pkg/telemetry"
	ide_views "github.com/daytonaio/daytona/pkg/views/ide"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := config.GetConfig()
		if err != nil {
			return err
		}

		ctx := context.Background()
		var workspaceId string
		var projectName string
		var providerConfigId *string
		var ideId string
		var workspace *apiclient.WorkspaceDTO

		activeProfile, err := c.GetActiveProfile()
		if err != nil {
			return err
		}

		ideId = c.DefaultIdeId

		apiClient, err := apiclient_util.GetApiClient(&activeProfile)
		if err != nil {
			return err
		}

		if len(args) == 0 {
			workspaceList, res, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Verbose(true).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			if len(workspaceList) == 0 {
				views_util.NotifyEmptyWorkspaceList(true)
				return nil
			}

			workspace = selection.GetWorkspaceFromPrompt(workspaceList, "Open")
			if workspace == nil {
				return nil
			}
			workspaceId = workspace.Id
		} else {
			workspace, err = apiclient_util.GetWorkspace(url.PathEscape(args[0]), true)
			if err != nil {
				if strings.Contains(err.Error(), workspaces.ErrWorkspaceNotFound.Error()) {
					log.Debug(err)
					return errors.New("workspace not found. You can see all workspace names by running the command `daytona list`")
				}
				return err
			}
			workspaceId = workspace.Id
		}

		if len(args) == 0 || len(args) == 1 {
			selectedProject, err := selectWorkspaceProject(workspaceId, &activeProfile)
			if err != nil {
				return err
			}
			if selectedProject == nil {
				return nil
			}

			projectName = selectedProject.Name
			providerConfigId = selectedProject.GitProviderConfigId
		}

		if len(args) == 2 {
			projectName = args[1]
			for _, project := range workspace.Projects {
				if project.Name == projectName {
					providerConfigId = project.GitProviderConfigId
					break
				}
			}
		}

		if ideFlag != "" {
			ideId = ideFlag
		}

		if !workspace_util.IsProjectRunning(workspace, projectName) {
			wsRunningStatus, err := AutoStartWorkspace(workspace.Name, projectName)
			if err != nil {
				return err
			}
			if !wsRunningStatus {
				return nil
			}
		}

		providerMetadata := ""
		if ideId != "ssh" {
			providerMetadata, err = workspace_util.GetProjectProviderMetadata(workspace, projectName)
			if err != nil {
				return err
			}
		}

		gpgKey, err := GetGitProviderGpgKey(apiClient, ctx, providerConfigId)
		if err != nil {
			log.Warn(err)
		}

		yesFlag, _ := cmd.Flags().GetBool("yes")
		ideList := config.GetIdeList()
		ide_views.RenderIdeOpeningMessage(workspace.Name, projectName, ideId, ideList)
		return openIDE(ideId, activeProfile, workspaceId, projectName, providerMetadata, yesFlag, gpgKey)
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
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

func openIDE(ideId string, activeProfile config.Profile, workspaceId string, projectName string, projectProviderMetadata string, yesFlag bool, gpgKey string) error {
	telemetry.AdditionalData["ide"] = ideId

	switch ideId {
	case "vscode":
		return ide.OpenVSCode(activeProfile, workspaceId, projectName, projectProviderMetadata, gpgKey)
	case "code-insiders":
		return ide.OpenVSCodeInsiders(activeProfile, workspaceId, projectName, projectProviderMetadata, gpgKey)
	case "ssh":
		return ide.OpenTerminalSsh(activeProfile, workspaceId, projectName, gpgKey, nil)
	case "browser":
		return ide.OpenBrowserIDE(activeProfile, workspaceId, projectName, projectProviderMetadata, gpgKey)
	case "codium":
		return ide.OpenVScodium(activeProfile, workspaceId, projectName, projectProviderMetadata, gpgKey)
	case "cursor":
		return ide.OpenCursor(activeProfile, workspaceId, projectName, projectProviderMetadata, gpgKey)
	case "jupyter":
		return ide.OpenJupyterIDE(activeProfile, workspaceId, projectName, projectProviderMetadata, yesFlag, gpgKey)
	case "fleet":
		return ide.OpenFleet(activeProfile, workspaceId, projectName, gpgKey)
	case "positron":
		return ide.OpenPositron(activeProfile, workspaceId, projectName, projectProviderMetadata, gpgKey)
	case "zed":
		return ide.OpenZed(activeProfile, workspaceId, projectName, gpgKey)
	case "windsurf":
		return ide.OpenWindsurf(activeProfile, workspaceId, projectName, projectProviderMetadata, gpgKey)
	default:
		_, ok := jetbrains.GetIdes()[jetbrains.Id(ideId)]
		if ok {
			return ide.OpenJetbrainsIDE(activeProfile, ideId, workspaceId, projectName, gpgKey)
		}
	}

	return errors.New("invalid IDE. Please choose one by running `daytona ide`")
}

var ideFlag string

func init() {
	ideList := config.GetIdeList()
	ids := make([]string, len(ideList))
	for i, ide := range ideList {
		ids[i] = ide.Id
	}
	ideListStr := strings.Join(ids, ", ")
	CodeCmd.Flags().StringVarP(&ideFlag, "ide", "i", "", fmt.Sprintf("Specify the IDE (%s)", ideListStr))

	CodeCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Automatically confirm any prompts")

}

func AutoStartWorkspace(workspaceId string, projectName string) (bool, error) {
	if !yesFlag {
		if !ide_views.RunStartWorkspaceForm(workspaceId) {
			return false, nil
		}
	}

	apiClient, err := apiclient_util.GetApiClient(nil)
	if err != nil {
		return false, err
	}

	err = StartWorkspace(apiClient, workspaceId, projectName)
	if err != nil {
		return false, err
	}

	return true, nil
}
