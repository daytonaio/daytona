// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"context"
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
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
		go apiclient_util.ReadBuildLogs(activeProfile, buildId, query, &stopLogs)

		if !continueOnCompletedFlag {
			err = waitForBuildToComplete(buildId, apiClient)
			if err != nil {
				log.Fatal(err)
			}
			stopLogs = true
		} else {
			// Sleep indefinitely
			select {}
		}

		// Make sure the terminal cursor is reset
		fmt.Print("\033[?25h")
	},
}

func waitForBuildToComplete(buildId string, apiClient *apiclient.APIClient) error {
	for {
		build, res, err := apiClient.BuildAPI.GetBuild(context.Background(), buildId).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		completedStates := []apiclient.BuildBuildState{
			apiclient.BuildStatePublished,
			apiclient.BuildStateError,
			apiclient.BuildStateDeleting,
		}

		if slices.Contains(completedStates, build.State) {
			// Allow the logs to be printed before exiting
			time.Sleep(time.Second)
			return nil
		}

		time.Sleep(time.Second)
	}
}

var followFlag bool
var continueOnCompletedFlag bool

func init() {
	buildLogsCmd.Flags().BoolVarP(&followFlag, "follow", "f", false, "Follow logs")
	buildLogsCmd.Flags().BoolVar(&continueOnCompletedFlag, "continue-on-completed", false, "Continue streaming logs after the build is completed")
}
