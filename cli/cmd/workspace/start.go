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

var startProjectFlag string

var StartCmd = &cobra.Command{
	Use:   "start [WORKSPACE_NAME]",
	Short: "Start the workspace",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		var workspaceName string

		apiClient := api.GetServerApiClient("http://localhost:3000", "")

		if len(args) == 0 {
			workspaceList, _, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Execute()
			if err != nil {
				log.Fatal(err)
			}

			workspaceName = select_prompt.GetWorkspaceNameFromPrompt(workspaceList, "start")
		} else {
			workspaceName = args[0]
		}

		wsName, wsMode := os.LookupEnv("DAYTONA_WS_NAME")
		if wsMode {
			workspaceName = wsName
		}

		if startProjectFlag == "" {
			_, err := apiClient.WorkspaceAPI.StartWorkspace(ctx, workspaceName).Execute()
			if err != nil {
				log.Fatal(err)
			}
		} else {
			_, err := apiClient.WorkspaceAPI.StartProject(ctx, workspaceName, startProjectFlag).Execute()
			if err != nil {
				log.Fatal(err)
			}
		}

		views_util.RenderInfoMessage(fmt.Sprintf("Workspace %s successfully started", workspaceName))
	},
}

func init() {
	_, exists := os.LookupEnv("DAYTONA_WS_DIR")
	if exists {
		StartCmd.Use = "start"
		StartCmd.Args = cobra.ExactArgs(0)
	}

	StartCmd.PersistentFlags().StringVarP(&startProjectFlag, "project", "p", "", "Start the single project in the workspace (project name)")
}
