// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"context"
	"log"
	"net/http"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/format"
	"github.com/daytonaio/daytona/pkg/views/build/info"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
	"github.com/spf13/cobra"
)

var buildInfoCmd = &cobra.Command{
	Use:     "info [BUILD]",
	Short:   "Show build info",
	Aliases: []string{"view", "inspect"},
	Args:    cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		var build *apiclient.Build

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		apiServerConfig, res, err := apiClient.ServerAPI.GetConfig(context.Background()).Execute()
		if err != nil {
			log.Fatal(apiclient_util.HandleErrorResponse(res, err))
		}

		if len(args) == 0 {
			buildList, res, err := apiClient.BuildAPI.ListBuilds(ctx).Execute()
			if err != nil {
				log.Fatal(apiclient_util.HandleErrorResponse(res, err))
			}

			if format.FormatFlag != "" {
				format.UnblockStdOut()
			}

			build = selection.GetBuildFromPrompt(buildList, "View")
			if format.FormatFlag != "" {
				format.BlockStdOut()
			}

			if build == nil {
				return
			}
		} else {
			var res *http.Response
			build, res, err = apiClient.BuildAPI.GetBuild(ctx, args[0]).Execute()
			if err != nil {
				log.Fatal(apiclient_util.HandleErrorResponse(res, err))
			}
		}

		if format.FormatFlag != "" {
			formattedData := format.NewFormatter(build)
			formattedData.Print()
			return
		}

		info.Render(build, apiServerConfig, false)
	},
}

func init() {
	format.RegisterFormatFlag(buildInfoCmd)
}
