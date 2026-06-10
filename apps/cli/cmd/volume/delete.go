// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package volume

import (
	"context"
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/cli/apiclient"
	"github.com/daytonaio/daytona/cli/cmd/common"
	view_common "github.com/daytonaio/daytona/cli/views/common"
	"github.com/spf13/cobra"
)

var DeleteCmd = &cobra.Command{
	Use:     "delete [VOLUME_ID_OR_NAME]",
	Short:   "Delete a volume",
	Args:    cobra.ExactArgs(1),
	Aliases: common.GetAliases("delete"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		apiClient, err := apiclient.GetApiClient(nil, nil)
		if err != nil {
			return err
		}

		// UUID-shaped arguments are treated as volume IDs first, but volume names
		// may also be UUID-shaped, so resolve by name when no ID matches. The
		// fallback only runs when the ID lookup found nothing, so it can never
		// delete a different volume than the one referenced.
		if isVolumeId(args[0]) {
			res, err := apiClient.VolumesAPI.DeleteVolume(ctx, args[0]).Execute()
			if err == nil {
				view_common.RenderInfoMessageBold(fmt.Sprintf("Volume %s deleted", args[0]))
				return nil
			}
			if res == nil || res.StatusCode != http.StatusNotFound {
				return apiclient.HandleErrorResponse(res, err)
			}
		}

		vol, res, err := apiClient.VolumesAPI.GetVolumeByName(ctx, args[0]).Execute()
		if err != nil {
			return apiclient.HandleErrorResponse(res, err)
		}

		res, err = apiClient.VolumesAPI.DeleteVolume(ctx, vol.Id).Execute()
		if err != nil {
			return apiclient.HandleErrorResponse(res, err)
		}

		view_common.RenderInfoMessageBold(fmt.Sprintf("Volume %s deleted", args[0]))
		return nil
	},
}
