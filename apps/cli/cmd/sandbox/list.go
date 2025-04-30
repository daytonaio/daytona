// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"context"

	"github.com/daytonaio/daytona-ai-saas/cli/apiclient"
	"github.com/daytonaio/daytona-ai-saas/cli/cmd/common"
	"github.com/daytonaio/daytona-ai-saas/cli/config"
	"github.com/daytonaio/daytona-ai-saas/cli/views/sandbox"
	"github.com/spf13/cobra"
)

func paginate[T any](items []T, page, limit int) []T {
	start := (page - 1) * limit
	if start > len(items) {
		return []T{}
	}

	end := start + limit
	if end > len(items) {
		end = len(items)
	}

	return items[start:end]
}

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

		sandboxList, res, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Verbose(verboseFlag).Execute()
		if err != nil {
			return apiclient.HandleErrorResponse(res, err)
		}

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
		paginatedSandBox := paginate(sandboxList, pageFlag, limitFlag)

		sandbox.ListSandboxes(paginatedSandBox, activeOrganizationName)
		return nil
	},
}

var verboseFlag bool
var pageFlag int
var limitFlag int

func init() {
	ListCmd.Flags().BoolVarP(&verboseFlag, "verbose", "v", false, "Include verbose output")
	ListCmd.Flags().IntVarP(&limitFlag, "limit", "l", 10, "Number of images to list per page")
	ListCmd.Flags().IntVarP(&pageFlag, "page", "p", 1, "Page number to retrive the data")
	common.RegisterFormatFlag(ListCmd)
}
