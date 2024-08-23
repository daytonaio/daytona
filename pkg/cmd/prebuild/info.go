// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package prebuild

import (
	"context"
	"log"
	"net/http"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/output"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/prebuild/info"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
	"github.com/spf13/cobra"
)

var prebuildInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show prebuild configuration info",
	Args:  cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		var prebuild *apiclient.PrebuildDTO
		var res *http.Response

		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		if len(args) < 2 {
			var prebuilds []apiclient.PrebuildDTO
			var selectedProjectConfigName string

			if len(args) == 1 {
				selectedProjectConfigName = args[0]
				prebuilds, res, err = apiClient.PrebuildAPI.ListPrebuildsForProjectConfig(context.Background(), selectedProjectConfigName).Execute()
				if err != nil {
					log.Fatal(apiclient_util.HandleErrorResponse(res, err))
				}
			} else {
				prebuilds, res, err = apiClient.PrebuildAPI.ListPrebuilds(context.Background()).Execute()
				if err != nil {
					log.Fatal(apiclient_util.HandleErrorResponse(res, err))
				}
			}

			if len(prebuilds) == 0 {
				views.RenderInfoMessage("No prebuilds found")
				return
			}

			prebuild = selection.GetPrebuildFromPrompt(prebuilds, "View")
			if prebuild == nil {
				return
			}
		} else {
			prebuild, res, err = apiClient.PrebuildAPI.GetPrebuild(ctx, args[0], args[1]).Execute()
			if err != nil {
				log.Fatal(apiclient_util.HandleErrorResponse(res, err))
			}
		}

		if output.FormatFlag != "" {
			output.Output = prebuild
			return
		}

		info.Render(prebuild, false)
	},
}
