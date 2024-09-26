// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"context"
	"errors"
	"fmt"

	"github.com/daytonaio/daytona/internal/util"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		var projectConfig *apiclient.ProjectConfig
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		projectConfigList, res, err := apiClient.ProjectConfigAPI.ListProjectConfigs(ctx).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		projectConfig = selection.GetProjectConfigFromPrompt(projectConfigList, 0, false, false, "Build")
		if projectConfig == nil {
			return nil
		}

		if projectConfig.BuildConfig == nil {
			return errors.New("The chosen project config does not have a build configuration")
		}

		chosenBranch, err := workspace_util.GetBranchFromProjectConfig(projectConfig, apiClient, 0)
		if err != nil {
			return err
		}

		if chosenBranch == nil {
			fmt.Println("Operation canceled")
			return nil
		}

		buildId, err := CreateBuild(apiClient, projectConfig, chosenBranch.Name, nil)
		if err != nil {
			return err
		}

		views.RenderViewBuildLogsMessage(buildId)
		return nil
	},
}

func CreateBuild(apiClient *apiclient.APIClient, projectConfig *apiclient.ProjectConfig, branch string, prebuildId *string) (string, error) {
	ctx := context.Background()

	profileData, res, err := apiClient.ProfileAPI.GetProfileData(ctx).Execute()
	if err != nil {
		return "", apiclient_util.HandleErrorResponse(res, err)
	}

	if projectConfig.BuildConfig == nil {
		return "", errors.New("the chosen project config does not have a build configuration")
	}

	createBuildDto := apiclient.CreateBuildDTO{
		ProjectConfigName: projectConfig.Name,
		Branch:            branch,
		PrebuildId:        prebuildId,
	}

	if profileData != nil {
		createBuildDto.EnvVars = util.MergeEnvVars(profileData.EnvVars, projectConfig.EnvVars)
	} else {
		createBuildDto.EnvVars = util.MergeEnvVars(projectConfig.EnvVars)
	}

	buildId, res, err := apiClient.BuildAPI.CreateBuild(ctx).CreateBuildDto(createBuildDto).Execute()
	if err != nil {
		return "", apiclient_util.HandleErrorResponse(res, err)
	}

	return buildId, nil
}
