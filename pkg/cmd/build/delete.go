// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"context"
	"fmt"

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
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		var buildId string

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		if allFlag {
			res, err := apiClient.BuildAPI.DeleteAllBuilds(ctx).Force(forceFlag).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			views.RenderInfoMessage("All builds have been marked for deletion")
			return nil
		}

		if prebuildIdFlag != "" {
			res, err := apiClient.BuildAPI.DeleteBuildsFromPrebuild(ctx, prebuildIdFlag).Force(forceFlag).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			views.RenderInfoMessage(fmt.Sprintf("All builds from prebuild %s have been marked for deletion\n", prebuildIdFlag))
			return nil
		}

		if len(args) == 0 {
			buildList, res, err := apiClient.BuildAPI.ListBuilds(ctx).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			build := selection.GetBuildFromPrompt(buildList, "Delete")
			if build == nil {
				return nil
			}
			buildId = build.Id
		} else {
			buildId = args[0]
		}

		res, err := apiClient.BuildAPI.DeleteBuild(ctx, buildId).Force(forceFlag).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}
		views.RenderInfoMessage(fmt.Sprintf("Build %s has been marked for deletion", buildId))
		return nil
	},
}

var allFlag bool
var forceFlag bool
var prebuildIdFlag string

func init() {
	buildDeleteCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Delete ALL builds")
	buildDeleteCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Force delete build")
	buildDeleteCmd.Flags().StringVar(&prebuildIdFlag, "prebuild-id", "", "Delete ALL builds from prebuild")
}
