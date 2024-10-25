// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package prebuild

import (
	"context"
	"net/http"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/format"
	"github.com/daytonaio/daytona/pkg/views/prebuild/info"
	"github.com/daytonaio/daytona/pkg/views/target/selection"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/spf13/cobra"
)

var prebuildInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show prebuild configuration info",
	Args:  cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		var prebuild *apiclient.PrebuildDTO
		var res *http.Response

		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		if len(args) < 2 {
			var prebuilds []apiclient.PrebuildDTO
			var selectedProjectConfigName string

			if len(args) == 1 {
				selectedProjectConfigName = args[0]
				prebuilds, res, err = apiClient.PrebuildAPI.ListPrebuildsForProjectConfig(context.Background(), selectedProjectConfigName).Execute()
				if err != nil {
					return apiclient_util.HandleErrorResponse(res, err)
				}
			} else {
				prebuilds, res, err = apiClient.PrebuildAPI.ListPrebuilds(context.Background()).Execute()
				if err != nil {
					return apiclient_util.HandleErrorResponse(res, err)
				}
			}

			if len(prebuilds) == 0 {
				views_util.NotifyEmptyPrebuildList(true)
				return nil
			}

			if format.FormatFlag != "" {
				format.UnblockStdOut()
			}

			prebuild = selection.GetPrebuildFromPrompt(prebuilds, "View")
			if format.FormatFlag != "" {
				format.BlockStdOut()
			}

			if prebuild == nil {
				return nil
			}
		} else {
			prebuild, res, err = apiClient.PrebuildAPI.GetPrebuild(ctx, args[0], args[1]).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}
		}

		if format.FormatFlag != "" {
			formattedData := format.NewFormatter(prebuild)
			formattedData.Print()
			return nil
		}

		info.Render(prebuild, false)
		return nil
	},
}

func init() {
	format.RegisterFormatFlag(prebuildInfoCmd)
}
