// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_workspace

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/cli/api"
	"github.com/daytonaio/daytona/cli/cmd/output"
	list_view "github.com/daytonaio/daytona/cli/cmd/views/workspace/list_view"

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

		if output.FormatFlag != "" {
			output.Output = workspaceList
			return
		}

		list_view.ListWorkspaces(workspaceList, "test")
		fmt.Println()

		// for _, workspaceInfo := range workspaceList {
		// 	fmt.Println(*workspaceInfo.Name)
		// 	fmt.Println(*workspaceInfo.Projects[0].Created)
		// 	fmt.Println(*workspaceInfo.Projects[0].Started)
		// 	fmt.Println(*workspaceInfo.Projects[0].IsRunning)
		// }
	},
}
