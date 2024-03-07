// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"
	"fmt"
	"os"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/server"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/views/util"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	allFlag bool
	yesFlag bool
)


var DeleteCmd = &cobra.Command{
	Use:     "delete [WORKSPACE]",
	Short:   "Delete a workspace",
	Aliases: []string{"remove", "rm"},
	Run: func(cmd *cobra.Command, args []string) {
		if allFlag {
			if yesFlag {
				deleteAllWorkspaces()
			} else {
				confirmation := util.ConfirmationPrompt("Are you sure you want to delete all workspaces?")
				if !confirmation {
					fmt.Println("Operation canceled.")
					return
				}
				deleteAllWorkspaces()
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
		var workspace *serverapiclient.Workspace

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

		res, err := apiClient.WorkspaceAPI.RemoveWorkspace(ctx, *workspace.Id).Execute()
		if err != nil {
			log.Fatal(apiclient.HandleErrorResponse(res, err))
		}

		config.RemoveWorkspaceSshEntries(activeProfile.Id, *workspace.Id)

		util.RenderInfoMessage(fmt.Sprintf("Workspace %s successfully deleted", *workspace.Name))
	},
}

func init() {
	DeleteCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Delete all workspaces")
	DeleteCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Confirm deletion without prompt")
}

func deleteAllWorkspaces() {
	ctx := context.Background()
	apiClient, err := server.GetApiClient(nil)
	if err != nil {
		log.Fatal(err)
	}

	workspaceList, res, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Execute()
	if err != nil {
		log.Fatal(apiclient.HandleErrorResponse(res, err))
	}

	for _, workspace := range workspaceList {
		res, err := apiClient.WorkspaceAPI.RemoveWorkspace(ctx, *workspace.Id).Execute()
		if err != nil {
			log.Errorf("Failed to delete workspace %s: %v", *workspace.Id, apiclient.HandleErrorResponse(res, err))
			continue
		}
		fmt.Printf("Workspace %s successfully deleted\n", *workspace.Name)
	}
}
