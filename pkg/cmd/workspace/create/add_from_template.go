// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package create

import (
	"context"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/common"
)

type AddWorkspaceFromTemplateParams struct {
	WorkspaceTemplate *apiclient.WorkspaceTemplate
	ApiClient         *apiclient.APIClient
	Workspaces        *[]apiclient.CreateWorkspaceDTO
	BranchFlag        *string
}

func AddWorkspaceFromTemplate(ctx context.Context, params AddWorkspaceFromTemplateParams) (*string, error) {
	chosenBranchName := ""
	if params.BranchFlag != nil {
		chosenBranchName = *params.BranchFlag
	}

	if chosenBranchName == "" {
		chosenBranch, err := GetBranchFromWorkspaceTemplate(ctx, params.WorkspaceTemplate, params.ApiClient, 0)
		if err != nil {
			return nil, err
		}
		if chosenBranch == nil {
			return nil, common.ErrCtrlCAbort
		}

		chosenBranchName = chosenBranch.Name
	}

	templateRepo, res, err := params.ApiClient.GitProviderAPI.GetGitContext(ctx).Repository(apiclient.GetRepositoryContext{
		Url:    params.WorkspaceTemplate.RepositoryUrl,
		Branch: &chosenBranchName,
	}).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	workspace := &apiclient.CreateWorkspaceDTO{
		Name:                params.WorkspaceTemplate.Name,
		GitProviderConfigId: params.WorkspaceTemplate.GitProviderConfigId,
		Source: apiclient.CreateWorkspaceSourceDTO{
			Repository: *templateRepo,
		},
		BuildConfig: params.WorkspaceTemplate.BuildConfig,
		Image:       &params.WorkspaceTemplate.Image,
		User:        &params.WorkspaceTemplate.User,
		EnvVars:     params.WorkspaceTemplate.EnvVars,
	}
	*params.Workspaces = append(*params.Workspaces, *workspace)

	return &params.WorkspaceTemplate.Name, nil
}
