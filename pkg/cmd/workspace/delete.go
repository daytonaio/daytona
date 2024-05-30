// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/daytonaio/daytona/cmd/daytona/config"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var yesFlag bool
var forceFlag bool

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

		ctx := context.Background()
		var workspace *apiclient.WorkspaceDTO

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		if len(args) == 0 {
			workspaceList, res, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Execute()
			if err != nil {
				log.Fatal(apiclient_util.HandleErrorResponse(res, err))
			}

			workspace = selection.GetWorkspaceFromPrompt(workspaceList, "Delete")
		} else {
			workspace, err = apiclient_util.GetWorkspace(args[0])
			if err != nil {
				log.Fatal(err)
			}
		}
		if workspace == nil {
			return
		}

		if !yesFlag {
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						Title(fmt.Sprintf("Delete workspace %s?", *workspace.Name)).
						Description(fmt.Sprintf("Are you sure you want to delete workspace %s?", *workspace.Name)).
						Value(&yesFlag),
				),
			).WithTheme(views.GetCustomTheme())

			err := form.Run()
			if err != nil {
				log.Fatal(err)
			}
		}

		if yesFlag {
			if !forceFlag {
				form := huh.NewForm(
					huh.NewGroup(
						huh.NewConfirm().
							Title(fmt.Sprintf("Delete workspace %s by force?", *workspace.Name)).
							Description("Provider resources might not be removed.").
							Value(&forceFlag),
					),
				).WithTheme(views.GetCustomTheme())

				err := form.Run()
				if err != nil {
					log.Fatal(err)
				}

				if forceFlag {
					forceRemoveWorkspace(ctx, apiClient, workspace)
				} else {
					err := removeWorkspace(ctx, apiClient, workspace)
					if err != nil {
						log.Fatal(err)
					}
				}
			} else {
				forceRemoveWorkspace(ctx, apiClient, workspace)
			}
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
	DeleteCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Delete a workspace by force")
}

func DeleteAllWorkspaces() error {
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
		res, err := apiClient.WorkspaceAPI.RemoveWorkspace(ctx, *workspace.Id).Execute()
		if err != nil {
			log.Errorf("Failed to delete workspace %s: %v", *workspace.Name, apiclient_util.HandleErrorResponse(res, err))
			continue
		}
		views.RenderLine(fmt.Sprintf("- Workspace %s successfully deleted\n", *workspace.Name))
	}
	return nil
}

func removeWorkspace(ctx context.Context, apiClient *apiclient.APIClient, workspace *apiclient.WorkspaceDTO) error {
	res, err := apiClient.WorkspaceAPI.RemoveWorkspace(ctx, *workspace.Id).Execute()
	if err != nil {
		log.Fatal(apiclient_util.HandleErrorResponse(res, err))
	}

	c, err := config.GetConfig()
	if err != nil {
		log.Fatal(err)
	}

	activeProfile, err := c.GetActiveProfile()
	if err != nil {
		log.Fatal(err)
	}

	err = config.RemoveWorkspaceSshEntries(activeProfile.Id, *workspace.Id)
	if err != nil {
		log.Fatal(err)
	}

	views.RenderInfoMessage(fmt.Sprintf("Workspace %s successfully deleted", *workspace.Name))
	return nil
}

func forceRemoveWorkspace(ctx context.Context, apiClient *apiclient.APIClient, workspace *apiclient.WorkspaceDTO) error {
	apiClient.WorkspaceAPI.RemoveWorkspace(ctx, *workspace.Id).Execute()

	c, _ := config.GetConfig()
	activeProfile, _ := c.GetActiveProfile()
	config.RemoveWorkspaceSshEntries(activeProfile.Id, *workspace.Id)

	views.RenderInfoMessage(fmt.Sprintf("Workspace %s successfully deleted", *workspace.Name))
	return nil
}
