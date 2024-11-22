// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"context"
	"errors"
	"fmt"

	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/conversion"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/workspace/create"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/selection"
	"github.com/spf13/cobra"
)

var buildRunCmd = &cobra.Command{
	Use:     "run",
	Short:   "Run a build from a workspace config",
	Aliases: []string{"create"},
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var workspaceConfig *apiclient.WorkspaceConfig
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		workspaceConfigList, res, err := apiClient.WorkspaceConfigAPI.ListWorkspaceConfigs(ctx).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		workspaceConfig = selection.GetWorkspaceConfigFromPrompt(workspaceConfigList, 0, false, false, "Build")
		if workspaceConfig == nil {
			return nil
		}

		if workspaceConfig.BuildConfig == nil {
			return errors.New("The chosen workspace config does not have a build configuration")
		}

		chosenBranch, err := create.GetBranchFromWorkspaceConfig(ctx, workspaceConfig, apiClient, 0)
		if err != nil {
			return err
		}

		if chosenBranch == nil {
			fmt.Println("Operation canceled")
			return nil
		}

		buildId, err := CreateBuild(apiClient, workspaceConfig, chosenBranch.Name, nil)
		if err != nil {
			return err
		}

		views.RenderViewBuildLogsMessage(buildId)
		return nil
	},
}

func CreateBuild(apiClient *apiclient.APIClient, workspaceConfig *apiclient.WorkspaceConfig, branch string, prebuildId *string) (string, error) {
	ctx := context.Background()

	envVars, res, err := apiClient.EnvVarAPI.ListEnvironmentVariables(ctx).Execute()
	if err != nil {
		return "", apiclient_util.HandleErrorResponse(res, err)
	}

	if workspaceConfig.BuildConfig == nil {
		return "", errors.New("the chosen workspace config does not have a build configuration")
	}

	createBuildDto := apiclient.CreateBuildDTO{
		WorkspaceConfigName: workspaceConfig.Name,
		Branch:              branch,
		PrebuildId:          prebuildId,
	}

	if envVars != nil {
		createBuildDto.EnvVars = util.MergeEnvVars(conversion.ToEnvVarsMap(envVars), workspaceConfig.EnvVars)
	} else {
		createBuildDto.EnvVars = util.MergeEnvVars(workspaceConfig.EnvVars)
	}

	buildId, res, err := apiClient.BuildAPI.CreateBuild(ctx).CreateBuildDto(createBuildDto).Execute()
	if err != nil {
		return "", apiclient_util.HandleErrorResponse(res, err)
	}

	return buildId, nil
}
