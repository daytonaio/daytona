// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/common"
	"github.com/daytonaio/daytona/pkg/views"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var yesFlag bool
var forceFlag bool

var DeleteCmd = &cobra.Command{
	Use:     "delete [WORKSPACE]...",
	Short:   "Delete a workspace",
	GroupID: util.TARGET_GROUP,
	Aliases: common.GetAliases("delete"),
	RunE: func(cmd *cobra.Command, args []string) error {
		if allFlag {
			if yesFlag {
				fmt.Println("Deleting all workspaces.")
				err := DeleteAllWorkspaces(forceFlag)
				if err != nil {
					return err
				}
			} else {
				form := huh.NewForm(
					huh.NewGroup(
						huh.NewConfirm().
							Title("Delete all workspaces?").
							Description("Are you sure you want to delete all workspaces?").
							Value(&yesFlag),
					),
				).WithTheme(views.GetCustomTheme())

				err := form.Run()
				if err != nil {
					return err
				}

				if yesFlag {
					err := DeleteAllWorkspaces(forceFlag)
					if err != nil {
						return err
					}
				} else {
					fmt.Println("Operation canceled.")
				}
			}
			return nil
		}

		ctx := context.Background()

		var workspaceDeleteList []*apiclient.WorkspaceDTO
		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		if len(args) == 0 {
			workspaceList, res, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			if len(workspaceList) == 0 {
				views_util.NotifyEmptyWorkspaceList(false)
				return nil
			}

			workspaceDeleteList = selection.GetWorkspacesFromPrompt(workspaceList, selection.DeleteActionVerb)
		} else {
			for _, arg := range args {
				workspace, _, err := apiclient_util.GetWorkspace(arg)
				if err != nil {
					log.Error(fmt.Sprintf("[ %s ] : %v", arg, err))
					continue
				}
				workspaceDeleteList = append(workspaceDeleteList, workspace)
			}
		}

		if len(workspaceDeleteList) == 0 {
			return nil
		}

		wsDeleteListNames := util.ArrayMap(workspaceDeleteList, func(w *apiclient.WorkspaceDTO) string {
			return w.Name
		})

		if !yesFlag {
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						Title(fmt.Sprintf("Delete workspace(s): [%s]?", strings.Join(wsDeleteListNames, ", "))).
						Description(fmt.Sprintf("Are you sure you want to delete the workspace(s): [%s]?", strings.Join(wsDeleteListNames, ", "))).
						Value(&yesFlag),
				),
			).WithTheme(views.GetCustomTheme())

			err := form.Run()
			if err != nil {
				return err
			}
		}

		if !yesFlag {
			fmt.Println("Operation canceled.")
		} else {
			for _, workspace := range workspaceDeleteList {
				err := common.DeleteWorkspace(ctx, apiClient, workspace.Id, workspace.Name, forceFlag)
				if err != nil {
					log.Error(fmt.Sprintf("[ %s ] : %v", workspace.Name, err))
					continue
				}
				views.RenderInfoMessage(fmt.Sprintf("Workspace '%s' successfully deleted", workspace.Name))
			}
		}
		return nil
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return common.GetWorkspaceNameCompletions()
	},
}

func init() {
	DeleteCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Delete all workspaces")
	DeleteCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Confirm deletion without prompt")
	DeleteCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Delete a workspace by force")
}

func DeleteAllWorkspaces(force bool) error {
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
		err := common.DeleteWorkspace(ctx, apiClient, workspace.Id, workspace.Name, force)
		if err != nil {
			log.Errorf("Failed to delete workspace %s: %v", workspace.Name, err)
			continue
		}
		views.RenderInfoMessage(fmt.Sprintf("- Workspace '%s' successfully deleted", workspace.Name))
	}
	return nil
}
