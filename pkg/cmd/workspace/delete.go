// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var yesFlag bool
var forceFlag bool

var DeleteCmd = &cobra.Command{
	Use:     "delete [WORKSPACE]",
	Short:   "Delete a workspace",
	GroupID: util.WORKSPACE_GROUP,
	Aliases: []string{"remove", "rm"},
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

		var workspaceDeleteList = []*apiclient.WorkspaceDTO{}
		var workspaceDeleteListNames = []string{}
		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		if len(args) == 0 {
			workspaceList, res, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}
			workspaceDeleteList = selection.GetWorkspacesFromPrompt(workspaceList, "Delete")
			for _, workspace := range workspaceDeleteList {
				workspaceDeleteListNames = append(workspaceDeleteListNames, workspace.Name)
			}
		} else {
			for _, arg := range args {
				workspace, err := apiclient_util.GetWorkspace(arg, false)
				if err != nil {
					log.Error(fmt.Sprintf("[ %s ] : %v", arg, err))
					continue
				}
				workspaceDeleteList = append(workspaceDeleteList, workspace)
				workspaceDeleteListNames = append(workspaceDeleteListNames, workspace.Name)
			}
		}

		if len(workspaceDeleteList) == 0 {
			return nil
		}

		if !yesFlag {
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						Title(fmt.Sprintf("Delete workspace(s): [%s]?", strings.Join(workspaceDeleteListNames, ", "))).
						Description(fmt.Sprintf("Are you sure you want to delete the workspace(s): [%s]?", strings.Join(workspaceDeleteListNames, ", "))).
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
				err := RemoveWorkspace(ctx, apiClient, workspace, forceFlag)
				if err != nil {
					log.Error(fmt.Sprintf("[ %s ] : %v", workspace.Name, err))
				}
				views.RenderInfoMessage(fmt.Sprintf("Workspace '%s' successfully deleted", workspace.Name))
			}
		}
		return nil
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		return getWorkspaceNameCompletions()
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
		err := RemoveWorkspace(ctx, apiClient, &workspace, force)
		if err != nil {
			log.Errorf("Failed to delete workspace %s: %v", workspace.Name, err)
			continue
		}
		views.RenderInfoMessage(fmt.Sprintf("- Workspace '%s' successfully deleted", workspace.Name))
	}
	return nil
}

func RemoveWorkspace(ctx context.Context, apiClient *apiclient.APIClient, workspace *apiclient.WorkspaceDTO, force bool) error {
	message := fmt.Sprintf("Deleting workspace %s", workspace.Name)
	err := views_util.WithInlineSpinner(message, func() error {
		res, err := apiClient.WorkspaceAPI.RemoveWorkspace(ctx, workspace.Id).Force(force).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}
		c, err := config.GetConfig()
		if err != nil {
			return err
		}

		activeProfile, err := c.GetActiveProfile()
		if err != nil {
			return err
		}

		err = config.RemoveWorkspaceSshEntries(activeProfile.Id, workspace.Id, "")
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
