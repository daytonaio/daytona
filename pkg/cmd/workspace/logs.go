// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/common"
	"github.com/daytonaio/daytona/pkg/cmd/format"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
	"github.com/spf13/cobra"
)

var followFlag bool

var LogsCmd = &cobra.Command{
	Use:     "logs [WORKSPACE]",
	Short:   "View the logs of a workspace",
	Args:    cobra.RangeArgs(0, 2),
	GroupID: util.TARGET_GROUP,
	Aliases: []string{"lg", "log"},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		c, err := config.GetConfig()
		if err != nil {
			return err
		}

		activeProfile, err := c.GetActiveProfile()
		if err != nil {
			return err
		}

		var ws *apiclient.WorkspaceDTO

		if len(args) == 0 {
			workspaceList, res, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Verbose(true).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			if len(workspaceList) == 0 {
				views_util.NotifyEmptyWorkspaceList(true)
				return nil
			}

			if format.FormatFlag != "" {
				format.UnblockStdOut()
			}

			ws = selection.GetWorkspaceFromPrompt(workspaceList, "View Logs For")
			if format.FormatFlag != "" {
				format.BlockStdOut()
			}

		} else {
			ws, _, err = apiclient_util.GetWorkspace(args[0], true)
			if err != nil {
				return err
			}
		}

		if ws == nil {
			return nil
		}

		apiclient_util.ReadWorkspaceLogs(ctx, apiclient_util.ReadLogParams{
			Id:            ws.Id,
			Label:         &ws.Name,
			ActiveProfile: activeProfile,
			Index:         util.Pointer(0),
			Follow:        &followFlag,
		})
		return nil
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return common.GetWorkspaceNameCompletions()
	},
}

func init() {
	LogsCmd.Flags().BoolVarP(&followFlag, "follow", "f", false, "Follow logs")
}
