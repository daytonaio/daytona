// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/cli/apiclient"
	"github.com/daytonaio/daytona/cli/cmd/common"
	view_common "github.com/daytonaio/daytona/cli/views/common"
	"github.com/spf13/cobra"
)

var DeleteCmd = &cobra.Command{
	Use:     "delete [SANDBOX_ID]",
	Short:   "Delete a sandbox",
	Args:    cobra.MaximumNArgs(1),
	Aliases: common.GetAliases("delete"),
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
				var deletedCount int

				for _, s := range sandboxList {
					res, err := apiClient.SandboxAPI.DeleteSandbox(ctx, s.Id).Force(forceFlag).Execute()
					if err != nil {
						fmt.Printf("Failed to delete sandbox %s: %s\n", s.Id, apiclient.HandleErrorResponse(res, err))
					} else {
						deletedCount++
					}
				}

				view_common.RenderInfoMessageBold(fmt.Sprintf("Deleted %d sandboxes", deletedCount))
				return nil
			}
			return cmd.Help()
		}

		deletionArg := args[0]

		var sandboxCount int

		for _, s := range sandboxList {
			if s.Id == args[0] {
				deletionArg = s.Id
				sandboxCount++
			}
		}

		switch sandboxCount {
		case 0:
			return fmt.Errorf("sandbox %s not found", args[0])
		case 1:
			res, err := apiClient.SandboxAPI.DeleteSandbox(ctx, deletionArg).Force(forceFlag).Execute()
			if err != nil {
				return apiclient.HandleErrorResponse(res, err)
			}

			view_common.RenderInfoMessageBold(fmt.Sprintf("Sandbox %s deleted", args[0]))
		default:
			return fmt.Errorf("multiple sandboxes with name %s found - please use the sandbox ID instead", args[0])
		}

		return nil
	},
}

var forceFlag bool

func init() {
	DeleteCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Delete all sandboxes")
	DeleteCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Force delete")
}
