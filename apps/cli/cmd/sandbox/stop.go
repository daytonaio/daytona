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

var StopCmd = &cobra.Command{
	Use:   "stop [SANDBOX_ID]",
	Short: "Stop a sandbox",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		apiClient, err := apiclient.GetApiClient(nil, nil)
		if err != nil {
			return err
		}

		sandboxList, res, err := apiClient.SandboxAPI.ListSandboxes(ctx).Execute()
		if err != nil {
			return apiclient.HandleErrorResponse(res, err)
		}

		if len(args) == 0 {
			if allFlag {
				var stoppedCount int

				for _, s := range sandboxList {
					res, err := apiClient.SandboxAPI.StopSandbox(ctx, s.Id).Execute()
					if err != nil {
						fmt.Printf("Failed to stop sandbox %s: %s\n", s.Id, apiclient.HandleErrorResponse(res, err))
					} else {
						stoppedCount++
					}
				}

				view_common.RenderInfoMessageBold(fmt.Sprintf("Stopped %d sandboxes", stoppedCount))
				return nil
			}
			return cmd.Help()
		}

		stopArg := args[0]
		var sandboxCount int

		for _, s := range sandboxList {
			if s.Id == args[0] {
				stopArg = s.Id
				sandboxCount++
			}
		}

		switch sandboxCount {
		case 0:
			return fmt.Errorf("sandbox %s not found", args[0])
		case 1:
			res, err := apiClient.SandboxAPI.StopSandbox(ctx, stopArg).Execute()
			if err != nil {
				return apiclient.HandleErrorResponse(res, err)
			}

			view_common.RenderInfoMessageBold(fmt.Sprintf("Sandbox %s stopped", args[0]))
		default:
			return fmt.Errorf("multiple sandboxes with name %s found - please use the sandbox ID instead", args[0])
		}

		return nil
	},
}

func init() {
	StopCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Stop all sandboxes")
}
