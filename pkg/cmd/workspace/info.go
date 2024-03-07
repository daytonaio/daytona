// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"
	"os"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/server"
	"github.com/daytonaio/daytona/pkg/cmd/output"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/views/workspace/info"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var InfoCmd = &cobra.Command{
	Use:     "info",
	Short:   "Show workspace info",
	Aliases: []string{"view"},
	Args:    cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		apiClient, err := server.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		var workspace *serverapiclient.Workspace

		if util.WorkspaceMode() {
			workspace, err = server.GetWorkspace(os.Getenv("DAYTONA_WS_ID"))
			if err != nil {
				log.Fatal(err)
			}
		} else if len(args) == 0 {
			workspaceList, res, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Execute()
			if err != nil {
				log.Fatal(apiclient.HandleErrorResponse(res, err))
			}

			workspace = selection.GetWorkspaceFromPrompt(workspaceList, "view")
		} else {
			workspace, err = server.GetWorkspace(args[0])
			if err != nil {
				log.Fatal(err)
			}
		}

		if output.FormatFlag != "" {
			output.Output = workspace
			return
		}

		info.Render(workspace)
	},
}

func init() {
	if !util.WorkspaceMode() {
		InfoCmd.Use += " [WORKSPACE]"
		InfoCmd.Args = cobra.RangeArgs(0, 1)
	}
}
