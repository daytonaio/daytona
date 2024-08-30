// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"slices"
	"time"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
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

		var exists *bool
		stopLogs := false
		go apiclient_util.ReadBuildLogs(activeProfile, buildId, query, &stopLogs)

		if !continueOnCompletedFlag {
			exists, err = waitForBuildToComplete(buildId, apiClient)
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

		if exists != nil && !*exists {
			views.RenderInfoMessage(fmt.Sprintf("Build with ID %s does not exist in the database", buildId))
		}
	},
}

func waitForBuildToComplete(buildId string, apiClient *apiclient.APIClient) (*bool, error) {
	for {
		build, res, err := apiClient.BuildAPI.GetBuild(context.Background(), buildId).Execute()
		if err != nil {
			if res.StatusCode == http.StatusNotFound {
				return util.Pointer(false), nil
			}
			return nil, apiclient_util.HandleErrorResponse(res, err)
		}

		completedStates := []apiclient.BuildBuildState{
			apiclient.BuildStatePublished,
			apiclient.BuildStateError,
		}

		if slices.Contains(completedStates, build.State) {
			// Allow the logs to be printed before exiting
			time.Sleep(time.Second)
			return util.Pointer(true), nil
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
