// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_workspace

import (
	"context"
	"fmt"
	"os"

	"github.com/daytonaio/daytona/cli/api"
	views_util "github.com/daytonaio/daytona/cli/cmd/views/util"
	select_prompt "github.com/daytonaio/daytona/cli/cmd/views/workspace/select_prompt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var stopProjectFlag string

var StopCmd = &cobra.Command{
	Use:   "stop [WORKSPACE_NAME]",
	Short: "Stop the workspace",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		var workspaceName string

		apiClient, err := api.GetServerApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		if len(args) == 0 {
			workspaceList, _, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Execute()
			if err != nil {
				log.Fatal(err)
			}

			workspaceName = select_prompt.GetWorkspaceNameFromPrompt(workspaceList, "stop")
		} else {
			workspaceName = args[0]
		}

		wsName, wsMode := os.LookupEnv("DAYTONA_WS_NAME")
		if wsMode {
			workspaceName = wsName
		}

		if stopProjectFlag == "" {
			_, err := apiClient.WorkspaceAPI.StopWorkspace(ctx, workspaceName).Execute()
			if err != nil {
				log.Fatal(err)
			}
		} else {
			_, err := apiClient.WorkspaceAPI.StopProject(ctx, workspaceName, stopProjectFlag).Execute()
			if err != nil {
				log.Fatal(err)
			}
		}

		views_util.RenderInfoMessage(fmt.Sprintf("Workspace %s successfully stopped", workspaceName))
	},
}

func init() {
	_, exists := os.LookupEnv("DAYTONA_WS_DIR")
	if exists {
		StopCmd.Use = "stop"
		StopCmd.Args = cobra.ExactArgs(0)
	}

	StopCmd.Flags().StringVarP(&stopProjectFlag, "project", "p", "", "Stop the single project in the workspace (project name)")
}
