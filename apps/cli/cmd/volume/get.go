// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package volume

import (
	"context"
	"net/http"

	apiclient_cli "github.com/daytonaio/daytona/cli/apiclient"
	"github.com/daytonaio/daytona/cli/cmd/common"
	"github.com/daytonaio/daytona/cli/views/volume"
	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// isVolumeId reports whether the argument is a canonical UUID (the length check
// excludes braced/URN/dashless variants).
func isVolumeId(arg string) bool {
	return len(arg) == 36 && uuid.Validate(arg) == nil
}

var GetCmd = &cobra.Command{
	Use:     "get [VOLUME_ID_OR_NAME]",
	Short:   "Get volume details",
	Args:    cobra.ExactArgs(1),
	Aliases: common.GetAliases("get"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		apiClient, err := apiclient_cli.GetApiClient(nil, nil)
		if err != nil {
			return err
		}

		// UUID-shaped arguments are volume IDs, anything else is a name
		var vol *apiclient.VolumeDto
		var res *http.Response
		if isVolumeId(args[0]) {
			vol, res, err = apiClient.VolumesAPI.GetVolume(ctx, args[0]).Execute()
		} else {
			vol, res, err = apiClient.VolumesAPI.GetVolumeByName(ctx, args[0]).Execute()
		}
		if err != nil {
			return apiclient_cli.HandleErrorResponse(res, err)
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
