// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package volume

import (
	"context"

	"github.com/daytonaio/daytona-ai-saas/cli/apiclient"
	"github.com/daytonaio/daytona-ai-saas/cli/cmd/common"
	"github.com/daytonaio/daytona-ai-saas/cli/views/volume"
	"github.com/spf13/cobra"
)

var ListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all volumes",
	Args:    cobra.NoArgs,
	Aliases: common.GetAliases("list"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		apiClient, err := apiclient.GetApiClient(nil, nil)
		if err != nil {
			return err
		}

		volumes, res, err := apiClient.VolumesAPI.ListVolumes(ctx).Execute()
		if err != nil {
			return apiclient.HandleErrorResponse(res, err)
		}

		if common.FormatFlag != "" {
			formattedData := common.NewFormatter(volumes)
			formattedData.Print()
			return nil
		}

		activeOrganizationName, err := common.GetActiveOrganizationName(apiClient, ctx)
		if err != nil {
			return err
		}

		volume.ListVolumes(volumes, &activeOrganizationName)
		return nil
	},
}

func init() {
	common.RegisterFormatFlag(ListCmd)
}
