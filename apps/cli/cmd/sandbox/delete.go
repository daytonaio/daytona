// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	apiclient_cli "github.com/daytonaio/daytona/cli/apiclient"
	"github.com/daytonaio/daytona/cli/cmd/common"
	view_common "github.com/daytonaio/daytona/cli/views/common"
	views_util "github.com/daytonaio/daytona/cli/views/util"
	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
	"github.com/spf13/cobra"
)

const spinnerThreshold = 10

var DeleteCmd = &cobra.Command{
	Use:     "delete [SANDBOX_ID] | [SANDBOX_NAME]",
	Short:   "Delete a sandbox",
	Args:    cobra.MaximumNArgs(1),
	Aliases: common.GetAliases("delete"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		apiClient, err := apiclient_cli.GetApiClient(nil, nil)
		if err != nil {
			return err
		}

		// Handle case when no sandbox ID is provided and allFlag is true
		if len(args) == 0 {
			if allFlag {
				page := float32(1.0)
				limit := float32(200.0) // 200 is the maximum limit for the API
				var allSandboxes []apiclient.Sandbox

				for {
					sandboxBatch, res, err := apiClient.SandboxAPI.ListSandboxesPaginated(ctx).Page(page).Limit(limit).Execute()
					if err != nil {
						return apiclient_cli.HandleErrorResponse(res, err)
					}

					allSandboxes = append(allSandboxes, sandboxBatch.Items...)

					if len(sandboxBatch.Items) < int(limit) || page >= float32(sandboxBatch.TotalPages) {
						break
					}
					page++
				}

				if len(allSandboxes) == 0 {
					view_common.RenderInfoMessageBold("No sandboxes to delete")
					return nil
				}

				var deletedCount int64

				deleteFn := func() error {
					var wg sync.WaitGroup
					sem := make(chan struct{}, 10) // limit to 10 concurrent deletes

					for _, sb := range allSandboxes {
						wg.Add(1)
						go func(sb apiclient.Sandbox) {
							defer wg.Done()
							sem <- struct{}{}
							defer func() { <-sem }()

							_, res, err := apiClient.SandboxAPI.DeleteSandbox(ctx, sb.Id).Execute()
							if err != nil {
								fmt.Printf("Failed to delete sandbox %s: %s\n", sb.Id, apiclient_cli.HandleErrorResponse(res, err))
							} else {
								atomic.AddInt64(&deletedCount, 1)
							}
						}(sb)
					}
					wg.Wait()
					return nil
				}

				if len(allSandboxes) > spinnerThreshold {
					err = views_util.WithInlineSpinner("Deleting all sandboxes", deleteFn)
				} else {
					err = deleteFn()
				}
				if err != nil {
					return err
				}

				view_common.RenderInfoMessageBold(fmt.Sprintf("Deleted %d sandboxes", atomic.LoadInt64(&deletedCount)))
				return nil
			}
			return cmd.Help()
		}

		// Handle case when a sandbox ID is provided
		sandboxIdOrNameArg := args[0]

		_, res, err := apiClient.SandboxAPI.DeleteSandbox(ctx, sandboxIdOrNameArg).Execute()
		if err != nil {
			return apiclient_cli.HandleErrorResponse(res, err)
		}

		view_common.RenderInfoMessageBold(fmt.Sprintf("Sandbox %s deleted", sandboxIdOrNameArg))

		return nil
	},
}

var allFlag bool

func init() {
	DeleteCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Delete all sandboxes")
}
