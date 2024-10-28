// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/common"
)

func AddWorkspaceFromConfig(workspaceConfig *apiclient.WorkspaceConfig, apiClient *apiclient.APIClient, workspaces *[]apiclient.CreateWorkspaceDTO, branchFlag *string) (*string, error) {
	chosenBranchName := ""
	if branchFlag != nil {
		chosenBranchName = *branchFlag
	}

	if chosenBranchName == "" {
		chosenBranch, err := GetBranchFromWorkspaceConfig(workspaceConfig, apiClient, 0)
		if err != nil {
			return nil, err
		}
		if chosenBranch == nil {
			return nil, common.ErrCtrlCAbort
		}

		chosenBranchName = chosenBranch.Name
	}

	configRepo, res, err := apiClient.GitProviderAPI.GetGitContext(context.Background()).Repository(apiclient.GetRepositoryContext{
		Url:    workspaceConfig.RepositoryUrl,
		Branch: &chosenBranchName,
	}).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	workspace := &apiclient.CreateWorkspaceDTO{
		Name:                workspaceConfig.Name,
		GitProviderConfigId: workspaceConfig.GitProviderConfigId,
		Source: apiclient.CreateWorkspaceSourceDTO{
			Repository: *configRepo,
		},
		BuildConfig: workspaceConfig.BuildConfig,
		Image:       &workspaceConfig.Image,
		User:        &workspaceConfig.User,
		EnvVars:     workspaceConfig.EnvVars,
	}
	*workspaces = append(*workspaces, *workspace)

	return &workspaceConfig.Name, nil
}
