// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_workspace

import (
	"context"

	"github.com/daytonaio/daytona/cli/api"
	"github.com/daytonaio/daytona/cli/cmd/output"
	views_util "github.com/daytonaio/daytona/cli/cmd/views/util"
	"github.com/daytonaio/daytona/cli/cmd/views/workspace/list_view"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var ListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List workspaces",
	Args:    cobra.ExactArgs(0),
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		apiClient, err := api.GetServerApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		workspaceList, _, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Execute()
		if err != nil {
			log.Fatal(err)
		}

		if len(workspaceList) == 0 {
			views_util.RenderInfoMessage("The workspace list is empty. Start off by running 'daytona create'.")
			return
		}

		if output.FormatFlag != "" {
			output.Output = workspaceList
			return
		}

		list_view.ListWorkspaces(workspaceList)
	},
}
