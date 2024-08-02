// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"github.com/daytonaio/daytona/pkg/apiclient"
)

func AddProjectFromConfig(projectConfig *apiclient.ProjectConfig, apiClient *apiclient.APIClient, projects *[]apiclient.CreateProjectConfigDTO, branchFlag string) (*string, error) {
	var err error
	chosenBranch := branchFlag

	if chosenBranch == "" {
		chosenBranch, err = GetBranchFromProjectConfig(projectConfig, apiClient, 0)
		if err != nil {
			return nil, err
		}
	}

	configRepo := projectConfig.Repository
	configRepo.Branch = &chosenBranch

	project := &apiclient.CreateProjectConfigDTO{
		Name: projectConfig.Name,
		Source: &apiclient.CreateProjectConfigSourceDTO{
			Repository: configRepo,
		},
		BuildConfig: projectConfig.BuildConfig,
		Image:       projectConfig.Image,
		User:        projectConfig.User,
		EnvVars:     projectConfig.EnvVars,
	}
	*projects = append(*projects, *project)

	return projectConfig.Name, nil
}
