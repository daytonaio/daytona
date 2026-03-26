// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package volume

import (
	"context"

	"github.com/daytonaio/daytona/cli/apiclient"
	"github.com/daytonaio/daytona/cli/cmd/common"
	"github.com/daytonaio/daytona/cli/views/volume"
	"github.com/spf13/cobra"
)

var GetCmd = &cobra.Command{
	Use:     "get [VOLUME_ID]",
	Short:   "Get volume details",
	Args:    cobra.ExactArgs(1),
	Aliases: common.GetAliases("get"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		apiClient, err := apiclient.GetApiClient(nil, nil)
		if err != nil {
			return err
		}

		vol, res, err := apiClient.VolumesAPI.GetVolume(ctx, args[0]).Execute()
		if err != nil {
			return apiclient.HandleErrorResponse(res, err)
		}

		if common.FormatFlag != "" {
			formattedData := common.NewFormatter(vol)
			formattedData.Print()
			return nil
		}

		volume.RenderInfo(vol, false)
		return nil
	},
}

func init() {
	common.RegisterFormatFlag(GetCmd)
}
