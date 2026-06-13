// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package volume

import (
	"context"
	"fmt"

	apiclient_cli "github.com/daytonaio/daytona/cli/apiclient"
	"github.com/daytonaio/daytona/cli/cmd/common"
	view_common "github.com/daytonaio/daytona/cli/views/common"
	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
	"github.com/spf13/cobra"
)

var CreateCmd = &cobra.Command{
	Use:   "create [NAME]",
	Short: "Create a volume",
	Example: `  daytona volume create my-volume
  # Mount it when creating a sandbox
  daytona create --snapshot my-snapshot:1.0 --volume my-volume:/data`,
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
