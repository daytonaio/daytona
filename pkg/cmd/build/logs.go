// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"context"
	"log"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
	"github.com/spf13/cobra"
)

var buildLogsCmd = &cobra.Command{
	Use:     "logs",
	Short:   "View logs for build",
	Args:    cobra.RangeArgs(0, 1),
	Aliases: []string{"log"},
	Run: func(cmd *cobra.Command, args []string) {
		c, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		activeProfile, err := c.GetActiveProfile()
		if err != nil {
			log.Fatal(err)
		}

		query := ""
		if followFlag {
			query += "follow=true"
		}

		ctx := context.Background()
		var buildId string

		apiClient, err := apiclient_util.GetApiClient(&activeProfile)
		if err != nil {
			log.Fatal(err)
		}

		if len(args) == 0 {
			buildList, res, err := apiClient.BuildAPI.ListBuilds(ctx).Execute()
			if err != nil {
				log.Fatal(apiclient_util.HandleErrorResponse(res, err))
			}
			build := selection.GetBuildFromPrompt(buildList, "Get Logs For")
			if build == nil {
				return
			}
			buildId = build.Id
		} else {
			buildId = args[0]
		}

		stopLogs := false
		apiclient_util.ReadBuildLogs(activeProfile, buildId, query, &stopLogs)
	},
}

var followFlag bool

func init() {
	buildLogsCmd.Flags().BoolVarP(&followFlag, "follow", "f", false, "Follow logs")
}
