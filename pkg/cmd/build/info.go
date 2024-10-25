// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"context"
	"net/http"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/format"
	"github.com/daytonaio/daytona/pkg/views/build/info"
	"github.com/daytonaio/daytona/pkg/views/target/selection"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/spf13/cobra"
)

var buildInfoCmd = &cobra.Command{
	Use:     "info [BUILD]",
	Short:   "Show build info",
	Aliases: []string{"view", "inspect"},
	Args:    cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		var build *apiclient.Build

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		apiServerConfig, res, err := apiClient.ServerAPI.GetConfig(context.Background()).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		if len(args) == 0 {
			buildList, res, err := apiClient.BuildAPI.ListBuilds(ctx).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			if len(buildList) == 0 {
				views_util.NotifyEmptyBuildList(true)
				return nil
			}

			if format.FormatFlag != "" {
				format.UnblockStdOut()
			}

			build = selection.GetBuildFromPrompt(buildList, "View")
			if format.FormatFlag != "" {
				format.BlockStdOut()
			}

			if build == nil {
				return nil
			}
		} else {
			var res *http.Response
			build, res, err = apiClient.BuildAPI.GetBuild(ctx, args[0]).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}
		}

		if format.FormatFlag != "" {
			formattedData := format.NewFormatter(build)
			formattedData.Print()
			return nil
		}

		info.Render(build, apiServerConfig, false)
		return nil
	},
}

func init() {
	format.RegisterFormatFlag(buildInfoCmd)
}
