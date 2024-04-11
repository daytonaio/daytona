// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/server"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/util"
	view_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var yesFlag bool

var DeleteCmd = &cobra.Command{
	Use:     "delete [WORKSPACE]",
	Short:   "Delete a workspace",
	Aliases: []string{"remove", "rm"},
	Run: func(cmd *cobra.Command, args []string) {
		if allFlag {
			if yesFlag {
				fmt.Println("Deleting all workspaces.")
				err := DeleteAllWorkspaces()
				if err != nil {
					log.Fatal(err)
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
					log.Fatal(err)
				}

				if yesFlag {
					err := DeleteAllWorkspaces()
					if err != nil {
						log.Fatal(err)
					}
				} else {
					fmt.Println("Operation canceled.")
				}
			}
			return
		}

		c, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		activeProfile, err := c.GetActiveProfile()
		if err != nil {
			log.Fatal(err)
		}

		ctx := context.Background()
		var workspace *serverapiclient.WorkspaceDTO

		apiClient, err := server.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		if len(args) == 0 {
			workspaceList, res, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Execute()
			if err != nil {
				log.Fatal(apiclient.HandleErrorResponse(res, err))
			}

			workspace = selection.GetWorkspaceFromPrompt(workspaceList, "delete")
		} else {
			workspace, err = server.GetWorkspace(args[0])
			if err != nil {
				log.Fatal(err)
			}
		}
		if workspace == nil {
			return
		}

		var confirmDeletionFlag bool

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Delete workspace?").
					Description("Are you sure you want to delete this workspace?").
					Value(&confirmDeletionFlag),
			),
		).WithTheme(views.GetCustomTheme())

		if err := form.Run(); err != nil {
			log.Fatal(err)
		}

		if confirmDeletionFlag {
			res, err := apiClient.WorkspaceAPI.RemoveWorkspace(ctx, *workspace.Id).Execute()
			if err != nil {
				log.Fatal(apiclient.HandleErrorResponse(res, err))
			}

			err = config.RemoveWorkspaceSshEntries(activeProfile.Id, *workspace.Id)
			if err != nil {
				log.Fatal(err)
			}

			util.RenderInfoMessage(fmt.Sprintf("Workspace %s successfully deleted", *workspace.Name))
		} else {
			fmt.Println("Operation canceled.")
		}
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
}

func DeleteAllWorkspaces() error {
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
		res, err := apiClient.WorkspaceAPI.RemoveWorkspace(ctx, *workspace.Id).Execute()
		if err != nil {
			log.Errorf("Failed to delete workspace %s: %v", *workspace.Id, apiclient.HandleErrorResponse(res, err))
			continue
		}
		view_util.RenderLine(fmt.Sprintf("- Workspace %s successfully deleted\n", *workspace.Name))
	}
	return nil
}
