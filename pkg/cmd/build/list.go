// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"context"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/format"
	view "github.com/daytonaio/daytona/pkg/views/build/list"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/spf13/cobra"
)

var buildListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all builds",
	Aliases: []string{"ls"},
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		apiServerConfig, res, err := apiClient.ServerAPI.GetConfig(context.Background()).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		buildList, res, err := apiClient.BuildAPI.ListBuilds(ctx).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		if format.FormatFlag != "" {
			formattedData := format.NewFormatter(buildList)
			formattedData.Print()
			return nil
		}

		if len(buildList) == 0 {
			views_util.NotifyEmptyBuildList(true)
			return nil
		}

		view.ListBuilds(buildList, apiServerConfig)
		return nil
	},
}

func init() {
	format.RegisterFormatFlag(buildListCmd)
}
