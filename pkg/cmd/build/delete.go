// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"context"
	"fmt"
	"log"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
	"github.com/spf13/cobra"
)

var buildDeleteCmd = &cobra.Command{
	Use:     "delete [BUILD]",
	Short:   "Delete a build",
	Aliases: []string{"remove", "rm"},
	Args:    cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		var buildId string

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		if allFlag {
			res, err := apiClient.BuildAPI.DeleteAllBuilds(ctx).Execute()
			if err != nil {
				log.Fatal(apiclient_util.HandleErrorResponse(res, err))
			}

			views.RenderInfoMessage("All builds deleted successfully")
			return
		}

		if prebuildIdFlag != "" {
			res, err := apiClient.BuildAPI.DeleteBuildsFromPrebuild(ctx, prebuildIdFlag).Execute()
			if err != nil {
				log.Fatal(apiclient_util.HandleErrorResponse(res, err))
			}

			views.RenderInfoMessage(fmt.Sprintf("All builds from prebuild %s deleted\n", prebuildIdFlag))
			return
		}

		if len(args) == 0 {
			buildList, res, err := apiClient.BuildAPI.ListBuilds(ctx).Execute()
			if err != nil {
				log.Fatal(apiclient_util.HandleErrorResponse(res, err))
			}

			build := selection.GetBuildFromPrompt(buildList, "Delete")
			if build == nil {
				return
			}
			buildId = build.Id
		} else {
			buildId = args[0]
		}

		res, err := apiClient.BuildAPI.DeleteBuild(ctx, buildId).Execute()
		if err != nil {
			log.Fatal(apiclient_util.HandleErrorResponse(res, err))
		}
		views.RenderInfoMessage(fmt.Sprintf("Build %s deleted successfully", buildId))
	},
}

var allFlag bool
var prebuildIdFlag string

func init() {
	buildDeleteCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Delete ALL builds")
	buildDeleteCmd.Flags().StringVar(&prebuildIdFlag, "prebuild-id", "", "Delete ALL builds from prebuild")
}
