// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package volume

import (
	"context"
	"fmt"

	"github.com/daytonaio/apiclient"
	apiclient_cli "github.com/daytonaio/daytona/cli/apiclient"
	"github.com/daytonaio/daytona/cli/cmd/common"
	view_common "github.com/daytonaio/daytona/cli/views/common"
	"github.com/spf13/cobra"
)

var CreateCmd = &cobra.Command{
	Use:     "create [NAME]",
	Short:   "Create a volume",
	Args:    cobra.ExactArgs(1),
	Aliases: common.GetAliases("create"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		apiClient, err := apiclient_cli.GetApiClient(nil, nil)
		if err != nil {
			return err
		}

		volume, res, err := apiClient.VolumesAPI.CreateVolume(ctx).CreateVolume(apiclient.CreateVolume{
			Name: args[0],
		}).Execute()
		if err != nil {
			return apiclient_cli.HandleErrorResponse(res, err)
		}

		view_common.RenderInfoMessageBold(fmt.Sprintf("Volume %s successfully created", volume.Name))
		return nil
	},
}

var sizeFlag int32

func init() {
	CreateCmd.Flags().Int32VarP(&sizeFlag, "size", "s", 10, "Size of the volume in GB")
}
