// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_workspace

import (
	"context"
	"fmt"
	"os"

	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/server"
	"github.com/daytonaio/daytona/pkg/views/util"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
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

		apiClient, err := server.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		if len(args) == 0 {
			workspaceList, res, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Execute()
			if err != nil {
				log.Fatal(apiclient.HandleErrorResponse(res, err))
			}

			workspaceName = selection.GetWorkspaceNameFromPrompt(workspaceList, "start")
		} else {
			workspaceName = args[0]
		}

		wsName, wsMode := os.LookupEnv("DAYTONA_WS_NAME")
		if wsMode {
			workspaceName = wsName
		}

		if startProjectFlag == "" {
			res, err := apiClient.WorkspaceAPI.StartWorkspace(ctx, workspaceName).Execute()
			if err != nil {
				log.Fatal(apiclient.HandleErrorResponse(res, err))
			}
		} else {
			res, err := apiClient.WorkspaceAPI.StartProject(ctx, workspaceName, startProjectFlag).Execute()
			if err != nil {
				log.Fatal(apiclient.HandleErrorResponse(res, err))
			}
		}

		util.RenderInfoMessage(fmt.Sprintf("Workspace %s successfully started", workspaceName))
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
