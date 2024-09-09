// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"

	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/format"
	"github.com/daytonaio/daytona/pkg/views/workspace/info"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var InfoCmd = &cobra.Command{
	Use:     "info [WORKSPACE]",
	Short:   "Show workspace info",
	Aliases: []string{"view", "inspect"},
	Args:    cobra.RangeArgs(0, 1),
	GroupID: util.WORKSPACE_GROUP,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		var workspace *apiclient.WorkspaceDTO

		if len(args) == 0 {
			workspaceList, res, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Verbose(true).Execute()
			if err != nil {
				log.Fatal(apiclient_util.HandleErrorResponse(res, err))
			}

			if format.FormatFlag != "" {
				format.UnblockStdOut()
			}

			workspace = selection.GetWorkspaceFromPrompt(workspaceList, "View")
			if format.FormatFlag != "" {
				format.BlockStdOut()
			}

		} else {
			workspace, err = apiclient_util.GetWorkspace(args[0], true)
			if err != nil {
				log.Fatal(err)
			}
		}

		if workspace == nil {
			return
		}

		if format.FormatFlag != "" {
			formattedData := format.NewFormatter(workspace)
			formattedData.Print()
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

func init() {
	format.RegisterFormatFlag(InfoCmd)
}
