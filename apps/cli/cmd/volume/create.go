// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package volume

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona-ai-saas/cli/apiclient"
	"github.com/daytonaio/daytona-ai-saas/cli/cmd/common"
	view_common "github.com/daytonaio/daytona-ai-saas/cli/views/common"
	"github.com/daytonaio/daytona-ai-saas/daytonaapiclient"
	"github.com/spf13/cobra"
)

var CreateCmd = &cobra.Command{
	Use:     "create [NAME]",
	Short:   "Create a volume",
	Args:    cobra.ExactArgs(1),
	Aliases: common.GetAliases("create"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		apiClient, err := apiclient.GetApiClient(nil, nil)
		if err != nil {
			return err
		}

		volume, res, err := apiClient.VolumesAPI.CreateVolume(ctx).CreateVolume(daytonaapiclient.CreateVolume{
			Name: args[0],
		}).Execute()
		if err != nil {
			return apiclient.HandleErrorResponse(res, err)
		}

		view_common.RenderInfoMessageBold(fmt.Sprintf("Volume %s successfully created", volume.Name))
		return nil
	},
}

var sizeFlag int32

func init() {
	CreateCmd.Flags().Int32VarP(&sizeFlag, "size", "s", 10, "Size of the volume in GB")
}
