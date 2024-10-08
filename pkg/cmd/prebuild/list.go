// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package prebuild

import (
	"context"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/format"
	"github.com/daytonaio/daytona/pkg/views"
	view "github.com/daytonaio/daytona/pkg/views/prebuild/list"
	"github.com/spf13/cobra"
)

var prebuildListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List prebuild configurations",
	Aliases: []string{"ls"},
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		prebuildList, res, err := apiClient.PrebuildAPI.ListPrebuilds(ctx).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		if len(prebuildList) == 0 {
			views.RenderInfoMessage("No prebuilds found. Add a new prebuild by running 'daytona prebuild add'")
			return nil
		}

		if format.FormatFlag != "" {
			formattedData := format.NewFormatter(prebuildList)
			formattedData.Print()
			return nil
		}

		view.ListPrebuilds(prebuildList)
		return nil
	},
}

func init() {
	format.RegisterFormatFlag(prebuildListCmd)
}
