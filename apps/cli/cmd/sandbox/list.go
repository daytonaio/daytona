// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"context"

	"github.com/daytonaio/daytona/cli/apiclient"
	"github.com/daytonaio/daytona/cli/cmd/common"
	"github.com/daytonaio/daytona/cli/config"
	"github.com/daytonaio/daytona/cli/views/sandbox"
	"github.com/spf13/cobra"
)

var (
	verboseFlag bool
	pageFlag    int
	limitFlag   int
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

		sandboxList, res, err := apiClient.SandboxAPI.ListSandboxes(ctx).Verbose(verboseFlag).Execute()
		if err != nil {
			return apiclient.HandleErrorResponse(res, err)
		}

		sandbox.SortSandboxes(&sandboxList)

		start := (pageFlag - 1) * limitFlag
		end := start + limitFlag
		if start > len(sandboxList) {
			start = len(sandboxList)
		}
		if end > len(sandboxList) {
			end = len(sandboxList)
		}
		paginatedList := sandboxList[start:end]

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

		sandbox.ListSandboxes(paginatedList, activeOrganizationName)
		return nil
	},
}

func init() {
	ListCmd.Flags().BoolVarP(&verboseFlag, "verbose", "v", false, "Include verbose output")
	ListCmd.Flags().IntVarP(&pageFlag, "page", "p", 1, "Page number for pagination (starting from 1)")
	ListCmd.Flags().IntVarP(&limitFlag, "limit", "l", 100, "Maximum number of items per page")
	common.RegisterFormatFlag(ListCmd)
}
