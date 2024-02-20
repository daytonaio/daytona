// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_workspace

import (
	"context"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/daytonaio/daytona/cli/api"
	"github.com/daytonaio/daytona/cli/cmd/output"
	view "github.com/daytonaio/daytona/cli/cmd/views/workspace/info_view"
	select_prompt "github.com/daytonaio/daytona/cli/cmd/views/workspace/select_prompt"
)

var InfoCmd = &cobra.Command{
	Use:     "info [WORKSPACE_NAME]",
	Short:   "Show workspace info",
	Aliases: []string{"view"},
	Args:    cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		var workspaceName string

		apiClient := api.GetServerApiClient("http://localhost:3000", "")

		if len(args) == 0 {
			workspaceList, _, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Execute()
			if err != nil {
				log.Fatal(err)
			}

			workspaceName = select_prompt.GetWorkspaceNameFromPrompt(workspaceList, "view")
		} else {
			workspaceName = args[0]
		}

		wsName, wsMode := os.LookupEnv("DAYTONA_WS_NAME")
		if wsMode {
			workspaceName = wsName
		}

		workspaceInfo, _, err := apiClient.WorkspaceAPI.GetWorkspaceInfo(ctx, workspaceName).Execute()
		if err != nil {
			log.Fatal(err)
		}

		if output.FormatFlag != "" {
			output.Output = workspaceInfo
			return
		}

		view.Render(workspaceInfo)
	},
}

func init() {
	_, exists := os.LookupEnv("DAYTONA_WS_DIR")
	if exists {
		InfoCmd.Use = "info"
		InfoCmd.Args = cobra.ExactArgs(0)
	}
}
