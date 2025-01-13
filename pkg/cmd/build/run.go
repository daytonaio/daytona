// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"context"
	"errors"
	"fmt"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/common"
	"github.com/daytonaio/daytona/pkg/cmd/workspace/create"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/selection"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/spf13/cobra"
)

var buildRunCmd = &cobra.Command{
	Use:     "run",
	Short:   "Run a build from a workspace template",
	Args:    cobra.NoArgs,
	Aliases: common.GetAliases("run"),
	RunE: func(cmd *cobra.Command, args []string) error {
		var workspaceTemplate *apiclient.WorkspaceTemplate
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		workspaceTemplateList, res, err := apiClient.WorkspaceTemplateAPI.ListWorkspaceTemplates(ctx).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		if len(workspaceTemplateList) == 0 {
			views_util.NotifyEmptyWorkspaceTemplateList(true)
			return nil
		}

		workspaceTemplate = selection.GetWorkspaceTemplateFromPrompt(workspaceTemplateList, 0, false, false, "Build")
		if workspaceTemplate == nil {
			return nil
		}

		if workspaceTemplate.BuildConfig == nil {
			return errors.New("The chosen workspace template does not have a build configuration")
		}

		chosenBranch, err := create.GetBranchFromWorkspaceTemplate(ctx, workspaceTemplate, apiClient, 0)
		if err != nil {
			return err
		}

		if chosenBranch == nil {
			fmt.Println("Operation canceled")
			return nil
		}

		buildId, err := CreateBuild(apiClient, workspaceTemplate, chosenBranch.Name, nil)
		if err != nil {
			return err
		}

		views.RenderViewBuildLogsMessage(buildId)
		return nil
	},
}

func CreateBuild(apiClient *apiclient.APIClient, workspaceTemplate *apiclient.WorkspaceTemplate, branch string, prebuildId *string) (string, error) {
	ctx := context.Background()

	if workspaceTemplate.BuildConfig == nil {
		return "", errors.New("the chosen workspace template does not have a build configuration")
	}

	createBuildDto := apiclient.CreateBuildDTO{
		WorkspaceTemplateName: workspaceTemplate.Name,
		Branch:                branch,
		PrebuildId:            prebuildId,
		EnvVars:               workspaceTemplate.EnvVars,
	}

	buildId, res, err := apiClient.BuildAPI.CreateBuild(ctx).CreateBuildDto(createBuildDto).Execute()
	if err != nil {
		return "", apiclient_util.HandleErrorResponse(res, err)
	}

	return buildId, nil
}
