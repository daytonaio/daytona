// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/server"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/output"
	"github.com/daytonaio/daytona/pkg/views/workspace/info"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var InfoCmd = &cobra.Command{
	Use:     "info [WORKSPACE]",
	Short:   "Show workspace info",
	Aliases: []string{"view"},
	Args:    cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		apiClient, err := server.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		var workspace *apiclient.WorkspaceDTO

		if len(args) == 0 {
			workspaceList, res, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Verbose(true).Execute()
			if err != nil {
				log.Fatal(apiclient_util.HandleErrorResponse(res, err))
			}

			workspace = selection.GetWorkspaceFromPrompt(workspaceList, "View")
		} else {
			workspace, err = server.GetWorkspace(args[0])
			if err != nil {
				log.Fatal(err)
			}
		}

		if workspace == nil {
			return
		}

		if output.FormatFlag != "" {
			output.Output = workspace
			return
		}

		info.Render(workspace, "", false)
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		return getWorkspaceNameCompletions()
	},
}
