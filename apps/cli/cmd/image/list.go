// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package image

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona-ai-saas/cli/apiclient"
	"github.com/daytonaio/daytona-ai-saas/cli/cmd/common"
	"github.com/daytonaio/daytona-ai-saas/cli/config"
	"github.com/daytonaio/daytona-ai-saas/cli/views/image"
	"github.com/spf13/cobra"
)

var (
	pageFlag  int
	limitFlag int
)

var ListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all images",
	Long:    "List all available Daytona images",
	Aliases: common.GetAliases("list"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		apiClient, err := apiclient.GetApiClient(nil, nil)
		if err != nil {
			return err
		}

		page := float32(1.0)
		limit := float32(100.0)

		if cmd.Flags().Changed("page") {
			page = float32(pageFlag)
		}

		if cmd.Flags().Changed("limit") {
			limit = float32(limitFlag)
		}

		images, res, err := apiClient.ImagesAPI.GetAllImages(ctx).Page(page).Limit(limit).Execute()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return apiclient.HandleErrorResponse(res, err)
		}

		if common.FormatFlag != "" {
			formattedData := common.NewFormatter(images.Items)
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

		image.ListImages(images.Items, activeOrganizationName)
		return nil
	},
}

func init() {
	common.RegisterFormatFlag(ListCmd)
	ListCmd.Flags().IntVarP(&pageFlag, "page", "p", 1, "Page number for pagination (starting from 1)")
	ListCmd.Flags().IntVarP(&limitFlag, "limit", "l", 100, "Maximum number of items per page")
}
