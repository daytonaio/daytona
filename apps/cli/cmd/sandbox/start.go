// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/cli/apiclient"
	view_common "github.com/daytonaio/daytona/cli/views/common"
	"github.com/spf13/cobra"
)

var StartCmd = &cobra.Command{
	Use:   "start [SANDBOX_ID]",
	Short: "Start a sandbox",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		apiClient, err := apiclient.GetApiClient(nil, nil)
		if err != nil {
			return err
		}

		sandboxList, res, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Execute()
		if err != nil {
			return apiclient.HandleErrorResponse(res, err)
		}

		if len(args) == 0 {
			if allFlag {
				var startedCount int

				for _, w := range sandboxList {
					res, err := apiClient.WorkspaceAPI.StartWorkspace(ctx, w.Id).Execute()
					if err != nil {
						fmt.Printf("Failed to start sandbox %s: %s\n", w.Id, apiclient.HandleErrorResponse(res, err))
					} else {
						startedCount++
					}
				}

				view_common.RenderInfoMessageBold(fmt.Sprintf("Started %d sandboxes", startedCount))
				return nil
			}
			return cmd.Help()
		}

		startArg := args[0]
		var sandboxCount int

		for _, w := range sandboxList {
			if w.Id == args[0] {
				startArg = w.Id
				sandboxCount++
			}
		}

		switch sandboxCount {
		case 0:
			return fmt.Errorf("sandbox %s not found", args[0])
		case 1:
			res, err := apiClient.WorkspaceAPI.StartWorkspace(ctx, startArg).Execute()
			if err != nil {
				return apiclient.HandleErrorResponse(res, err)
			}

			view_common.RenderInfoMessageBold(fmt.Sprintf("Sandbox %s started", args[0]))
		default:
			return fmt.Errorf("multiple sandboxes with name %s found - please use the sandbox ID instead", args[0])
		}

		return nil
	},
}

var allFlag bool

func init() {
	StartCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Start all sandboxes")
}
