// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/common"
	cmd_common "github.com/daytonaio/daytona/pkg/cmd/common"
	"github.com/daytonaio/daytona/pkg/views"
	ide_views "github.com/daytonaio/daytona/pkg/views/ide"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var allFlag bool
var codeFlag bool

var StartCmd = &cobra.Command{
	Use:     "start [WORKSPACE]...",
	Short:   "Start a workspace",
	GroupID: util.TARGET_GROUP,
	RunE: func(cmd *cobra.Command, args []string) error {
		var selectedWorkspaces []*apiclient.WorkspaceDTO
		var activeProfile config.Profile
		var ideId string
		var ideList []config.Ide
		var providerConfigId *string
		workspaceProviderMetadata := ""

		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		if allFlag {
			return startAllWorkspaces()
		}

		if len(args) == 0 {
			workspaceList, res, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			if len(workspaceList) == 0 {
				views_util.NotifyEmptyWorkspaceList(true)
				return nil
			}

			selectedWorkspaces = selection.GetWorkspacesFromPrompt(workspaceList, selection.StartActionVerb)
		} else {
			for _, arg := range args {
				workspace, _, err := apiclient_util.GetWorkspace(arg)
				if err != nil {
					log.Error(fmt.Sprintf("[ %s ] : %v", arg, err))
					continue
				}
				selectedWorkspaces = append(selectedWorkspaces, workspace)
			}
		}

		if len(selectedWorkspaces) == 1 {
			var ws *apiclient.WorkspaceDTO
			var res *http.Response
			workspace := selectedWorkspaces[0]
			if codeFlag {
				c, err := config.GetConfig()
				if err != nil {
					return err
				}

				activeProfile, err = c.GetActiveProfile()
				if err != nil {
					return err
				}

				ideList = config.GetIdeList()
				ideId = c.DefaultIdeId

				ws, res, err = apiClient.WorkspaceAPI.GetWorkspace(ctx, workspace.Id).Execute()
				if err != nil {
					return apiclient_util.HandleErrorResponse(res, err)
				}
				if ideId != "ssh" {
					workspaceProviderMetadata = *ws.ProviderMetadata
				}
			}

			err = StartWorkspace(apiClient, *workspace)
			if err != nil {
				return err
			}
			gpgKey, err := common.GetGitProviderGpgKey(apiClient, ctx, providerConfigId)
			if err != nil {
				log.Warn(err)
			}

			views.RenderInfoMessage(fmt.Sprintf("Workspace '%s' started successfully", workspace.Name))

			if codeFlag {
				ide_views.RenderIdeOpeningMessage(ws.TargetId, ws.Name, ideId, ideList)
				err = common.OpenIDE(ideId, activeProfile, ws.Id, workspaceProviderMetadata, yesFlag, gpgKey)
				if err != nil {
					return err
				}
			}
		} else {
			for _, ws := range selectedWorkspaces {
				err := StartWorkspace(apiClient, *ws)
				if err != nil {
					log.Errorf("Failed to start workspace %s: %v\n\n", ws.Name, err)
					continue
				}
				views.RenderInfoMessage(fmt.Sprintf("- Workspace '%s' started successfully", ws.Name))
			}
		}
		return nil
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return common.GetAllWorkspacesByState(apiclient.ResourceStateNameStopped)
	},
}

func init() {
	StartCmd.PersistentFlags().BoolVarP(&allFlag, "all", "a", false, "Start all targets")
	StartCmd.PersistentFlags().BoolVarP(&codeFlag, "code", "c", false, "Open the target in the IDE after target start")
	StartCmd.PersistentFlags().BoolVarP(&yesFlag, "yes", "y", false, "Automatically confirm any prompts")
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
		err := StartWorkspace(apiClient, workspace)
		if err != nil {
			log.Errorf("Failed to start workspace %s: %v\n\n", workspace.Name, err)
			continue
		}

		views.RenderInfoMessage(fmt.Sprintf("- Workspace '%s' started successfully", workspace.Name))
	}
	return nil
}

func StartWorkspace(apiClient *apiclient.APIClient, workspace apiclient.WorkspaceDTO) error {
	ctx := context.Background()
	timeFormat := time.Now().Format("2006-01-02 15:04:05")
	from, err := time.Parse("2006-01-02 15:04:05", timeFormat)
	if err != nil {
		return err
	}

	c, err := config.GetConfig()
	if err != nil {
		return err
	}

	activeProfile, err := c.GetActiveProfile()
	if err != nil {
		return err
	}

	logsContext, stopLogs := context.WithCancel(context.Background())
	go cmd_common.ReadWorkspaceLogs(logsContext, cmd_common.ReadLogParams{
		Id:        workspace.Id,
		Label:     &workspace.Name,
		ServerUrl: activeProfile.Api.Url,
		ApiKey:    activeProfile.Api.Key,
		Index:     util.Pointer(0),
		Follow:    util.Pointer(true),
		From:      &from,
	})

	res, err := apiClient.WorkspaceAPI.StartWorkspace(ctx, workspace.Id).Execute()
	if err != nil {
		stopLogs()
		return apiclient_util.HandleErrorResponse(res, err)
	}

	err = common.AwaitWorkspaceState(workspace.Id, apiclient.ResourceStateNameStarted)
	if err != nil {
		stopLogs()
		return err
	}

	time.Sleep(100 * time.Millisecond)

	stopLogs()
	return nil
}
