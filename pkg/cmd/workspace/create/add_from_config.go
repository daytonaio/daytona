// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package create

import (
	"context"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/common"
)

type AddWorkspaceFromConfigParams struct {
	WorkspaceConfig *apiclient.WorkspaceConfig
	ApiClient       *apiclient.APIClient
	Workspaces      *[]apiclient.CreateWorkspaceDTO
	BranchFlag      *string
}

func AddWorkspaceFromConfig(ctx context.Context, params AddWorkspaceFromConfigParams) (*string, error) {
	chosenBranchName := ""
	if params.BranchFlag != nil {
		chosenBranchName = *params.BranchFlag
	}

	if chosenBranchName == "" {
		chosenBranch, err := GetBranchFromWorkspaceConfig(ctx, params.WorkspaceConfig, params.ApiClient, 0)
		if err != nil {
			return nil, err
		}
		if chosenBranch == nil {
			return nil, common.ErrCtrlCAbort
		}

		chosenBranchName = chosenBranch.Name
	}

	configRepo, res, err := params.ApiClient.GitProviderAPI.GetGitContext(ctx).Repository(apiclient.GetRepositoryContext{
		Url:    params.WorkspaceConfig.RepositoryUrl,
		Branch: &chosenBranchName,
	}).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	workspace := &apiclient.CreateWorkspaceDTO{
		Name:                params.WorkspaceConfig.Name,
		GitProviderConfigId: params.WorkspaceConfig.GitProviderConfigId,
		Source: apiclient.CreateWorkspaceSourceDTO{
			Repository: *configRepo,
		},
		BuildConfig: params.WorkspaceConfig.BuildConfig,
		Image:       &params.WorkspaceConfig.Image,
		User:        &params.WorkspaceConfig.User,
		EnvVars:     params.WorkspaceConfig.EnvVars,
	}
	*params.Workspaces = append(*params.Workspaces, *workspace)

	return &params.WorkspaceConfig.Name, nil
}
