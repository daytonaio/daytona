// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"context"
	"log"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	workspace_util "github.com/daytonaio/daytona/pkg/cmd/workspace/util"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
	"github.com/spf13/cobra"
)

var buildRunCmd = &cobra.Command{
	Use:     "run",
	Short:   "Run a build from a project config",
	Aliases: []string{"create"},
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		var projectConfig *apiclient.ProjectConfig
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		projectConfigList, res, err := apiClient.ProjectConfigAPI.ListProjectConfigs(ctx).Execute()
		if err != nil {
			log.Fatal(apiclient_util.HandleErrorResponse(res, err))
		}

		projectConfig = selection.GetProjectConfigFromPrompt(projectConfigList, 0, false, "Update")
		if projectConfig == nil {
			return
		}

		if projectConfig.BuildConfig == nil {
			log.Fatal("The chosen project config does not have a build configuration")
		}

		chosenBranch, err := workspace_util.GetBranchFromProjectConfig(projectConfig, apiClient, 0)
		if err != nil {
			log.Fatal(err)
		}

		createBuildDto := apiclient.CreateBuildDTO{
			ProjectConfigName: projectConfig.Name,
		}

		if chosenBranch != nil && chosenBranch.Name != "" {
			createBuildDto.Branch = &chosenBranch.Name
		}

		buildId, res, err := apiClient.BuildAPI.CreateBuild(ctx).CreateBuildDto(createBuildDto).Execute()
		if err != nil {
			log.Fatal(apiclient_util.HandleErrorResponse(res, err))
		}

		views.RenderViewBuildLogsMessage(buildId)
	},
}
