// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"errors"

	"github.com/daytonaio/daytona/pkg/serverapiclient"
	create_view "github.com/daytonaio/daytona/pkg/views/workspace/create"
)

func GetCreationDataFromPrompt(workspaceNames []string, userGitProviders []serverapiclient.GitProvider, manual bool, multiProject bool) (workspaceName string, projectRepositoryList []serverapiclient.GitRepository, err error) {
	var projectRepoList []serverapiclient.GitRepository
	var providerRepo serverapiclient.GitRepository

	if !manual && userGitProviders != nil && len(userGitProviders) > 0 {
		providerRepo, err = getRepositoryFromWizard(userGitProviders, 0)
		if err != nil {
			return "", nil, err
		}
		if providerRepo == (serverapiclient.GitRepository{}) {
			return "", nil, nil
		}
	}

	workspaceCreationPromptResponse, err := create_view.RunInitialForm(providerRepo, multiProject)
	if err != nil {
		return "", nil, err
	}

	if workspaceCreationPromptResponse.PrimaryRepository == (serverapiclient.GitRepository{}) {
		return "", nil, errors.New("primary repository is required")
	}

	projectRepoList = []serverapiclient.GitRepository{workspaceCreationPromptResponse.PrimaryRepository}

	if workspaceCreationPromptResponse.SecondaryProjectCount > 0 {

		if !manual && userGitProviders != nil && len(userGitProviders) > 0 {
			for i := 0; i < workspaceCreationPromptResponse.SecondaryProjectCount; i++ {
				providerRepo, err = getRepositoryFromWizard(userGitProviders, i+1)
				if err != nil {
					return "", nil, err
				}
				if providerRepo == (serverapiclient.GitRepository{}) {
					return "", nil, nil
				}
				workspaceCreationPromptResponse.SecondaryRepositories = append(workspaceCreationPromptResponse.SecondaryRepositories, providerRepo)
			}
		}

		workspaceCreationPromptResponse, err = create_view.RunSecondaryProjectsForm(workspaceCreationPromptResponse)
		if err != nil {
			return "", nil, err
		}

		projectRepoList = append(projectRepoList, workspaceCreationPromptResponse.SecondaryRepositories...)
	}

	suggestedName := create_view.GetSuggestedWorkspaceName(*workspaceCreationPromptResponse.PrimaryRepository.Url)

	workspaceCreationPromptResponse, err = create_view.RunWorkspaceNameForm(workspaceCreationPromptResponse, suggestedName, workspaceNames)
	if err != nil {
		return "", nil, err
	}

	if workspaceCreationPromptResponse.WorkspaceName == "" {
		return "", nil, errors.New("workspace name is required")
	}

	return workspaceCreationPromptResponse.WorkspaceName, projectRepoList, nil
}
