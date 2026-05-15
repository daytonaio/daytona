// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/cli/apiclient"
	"github.com/daytonaio/daytona/cli/cmd/common"
	"github.com/daytonaio/daytona/cli/config"
	"github.com/daytonaio/daytona/cli/views/sandbox"
	"github.com/spf13/cobra"
)

var (
	cursorFlag string
	limitFlag  int
)

var ListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List sandboxes",
	Args:    cobra.NoArgs,
	Aliases: common.GetAliases("list"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		apiClient, err := apiclient.GetApiClient(nil, nil)
		if err != nil {
			return err
		}

		limit := float32(100.0)

		if cmd.Flags().Changed("limit") {
			limit = float32(limitFlag)
		}

		request := apiClient.SandboxAPI.ListSandboxes(ctx).Limit(limit)
		if cmd.Flags().Changed("cursor") {
			request = request.Cursor(cursorFlag)
		}

		sandboxList, res, err := request.Execute()
		if err != nil {
			return apiclient.HandleErrorResponse(res, err)
		}

		sandbox.SortSandboxes(&sandboxList.Items)

		if common.FormatFlag != "" {
			formattedData := common.NewFormatter(sandboxList)
			formattedData.Print()
			return nil
		}

		var activeOrganizationName *string

		if !config.IsApiKeyAuth() {
			name, err := common.GetActiveOrganizationName(apiClient, ctx)
			if err != nil {
				return err
			}
			activeOrganizationName = &name
		}

		sandbox.ListSandboxes(sandboxList.Items, activeOrganizationName)

		if sandboxList.NextCursor.IsSet() && sandboxList.NextCursor.Get() != nil {
			fmt.Printf("\nNext cursor: %s\n", *sandboxList.NextCursor.Get())
		}

		return nil
	},
}

func init() {
	ListCmd.Flags().StringVarP(&cursorFlag, "cursor", "c", "", "Cursor for pagination")
	ListCmd.Flags().IntVarP(&limitFlag, "limit", "l", 100, "Maximum number of items per page")
	common.RegisterFormatFlag(ListCmd)
}
