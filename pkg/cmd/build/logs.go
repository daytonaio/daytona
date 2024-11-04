// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
	"github.com/spf13/cobra"
)

var buildLogsCmd = &cobra.Command{
	Use:     "logs",
	Short:   "View logs for build",
	Args:    cobra.RangeArgs(0, 1),
	Aliases: []string{"log"},
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := config.GetConfig()
		if err != nil {
			return err
		}

		activeProfile, err := c.GetActiveProfile()
		if err != nil {
			return err
		}

		query := ""
		if followFlag {
			query += "follow=true"
		}

		ctx := context.Background()
		var buildId string

		apiClient, err := apiclient_util.GetApiClient(&activeProfile)
		if err != nil {
			return err
		}

		if len(args) == 0 {
			buildList, res, err := apiClient.BuildAPI.ListBuilds(ctx).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			if len(buildList) == 0 {
				views_util.NotifyEmptyBuildList(false)
				return nil
			}

			build := selection.GetBuildFromPrompt(buildList, "Get Logs For")
			if build == nil {
				return nil
			}
			buildId = build.Id
		} else {
			buildId = args[0]
		}

		_, _, err = apiClient.BuildAPI.GetBuild(ctx, buildId).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(nil, err)
		}

		apiclient_util.ReadBuildLogs(ctx, activeProfile, buildId, query)

		// Make sure the terminal cursor is reset
		fmt.Print("\033[?25h")

		return nil
	},
}

var followFlag bool

func init() {
	buildLogsCmd.Flags().BoolVarP(&followFlag, "follow", "f", false, "Follow logs")
}
