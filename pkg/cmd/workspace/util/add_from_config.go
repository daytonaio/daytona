// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"
	"net/url"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
)

func AddProjectFromConfig(projectConfig *apiclient.ProjectConfig, apiClient *apiclient.APIClient, projects *[]apiclient.CreateProjectDTO, branchFlag string) (*string, error) {
	chosenBranchName := branchFlag

	if chosenBranchName == "" {
		chosenBranch, err := GetBranchFromProjectConfig(projectConfig, apiClient, 0)
		if err != nil {
			return nil, err
		}
		if chosenBranch != nil {
			chosenBranchName = chosenBranch.Name
		}
	}

	repo := &apiclient.GitRepository{
		Url:    projectConfig.RepositoryUrl,
		Branch: &chosenBranchName,
	}

	newRepoUrl, res, err := apiClient.GitProviderAPI.GetUrlFromRepository(context.Background()).Repository(*repo).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	configRepo, res, err := apiClient.GitProviderAPI.GetGitContext(context.Background(), url.QueryEscape(newRepoUrl.Url)).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	project := &apiclient.CreateProjectDTO{
		Name: projectConfig.Name,
		Source: apiclient.CreateProjectSourceDTO{
			Repository: *configRepo,
		},
		BuildConfig: projectConfig.BuildConfig,
		Image:       &projectConfig.Image,
		User:        &projectConfig.User,
		EnvVars:     projectConfig.EnvVars,
	}
	*projects = append(*projects, *project)

	return &projectConfig.Name, nil
}
