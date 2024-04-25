// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"errors"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/views/workspace/create"
)

func GetCreationDataFromPrompt(apiServerConfig *serverapiclient.ServerConfig, workspaceNames []string, userGitProviders []serverapiclient.GitProvider, manual bool, multiProject bool) (string, []serverapiclient.CreateWorkspaceRequestProject, error) {
	var projectList []serverapiclient.CreateWorkspaceRequestProject
	var providerRepo serverapiclient.GitRepository
	var providerRepoUrl string
	var err error
	var workspaceName string
	var primaryContainerImage string
	var primaryContainerUser string
	doneCheck := true

	if !manual && userGitProviders != nil && len(userGitProviders) > 0 {
		providerRepo, err = getRepositoryFromWizard(userGitProviders, 0)
		if err != nil {
			return "", nil, err
		}
		if providerRepo == (serverapiclient.GitRepository{}) {
			return "", nil, nil
		}
	}

	if providerRepo.Url == nil {
		providerRepo.Url = new(string)
	}

	workspaceCreationPromptResponse, err := create.RunInitialForm(*providerRepo.Url, multiProject)
	if err != nil {
		return "", nil, err
	}

	if workspaceCreationPromptResponse.PrimaryProject.Source == nil {
		return "", nil, errors.New("primary project is required")
	}

	projectList = []serverapiclient.CreateWorkspaceRequestProject{workspaceCreationPromptResponse.PrimaryProject}

	if multiProject {
		for i := 0; workspaceCreationPromptResponse.AddingMoreProjects; i++ {

			if !manual && userGitProviders != nil && len(userGitProviders) > 0 {
				providerRepo, err = getRepositoryFromWizard(userGitProviders, i+1)
				if err != nil {
					return "", nil, err
				}
				if providerRepo == (serverapiclient.GitRepository{}) {
					return "", nil, nil
				}

				providerRepoUrl = *providerRepo.Url
			}

			workspaceCreationPromptResponse, err = create.RunProjectForm(workspaceCreationPromptResponse, providerRepoUrl)
			if err != nil {
				return "", nil, err
			}
			providerRepoUrl = ""
		}
		projectList = append(projectList, workspaceCreationPromptResponse.SecondaryProjects...)
	}

	for i, project := range projectList {
		if project.Source == nil || project.Source.Repository == nil || project.Source.Repository.Url == nil {
			return "", nil, errors.New("repository is required")
		}
		projectName := GetProjectNameFromRepo(*project.Source.Repository.Url)
		projectList[i].Name = projectName
	}

	suggestedName := create.GetSuggestedWorkspaceName(*workspaceCreationPromptResponse.PrimaryProject.Source.Repository.Url)

	workspaceName, primaryContainerImage, primaryContainerUser, err = create.GetWorkspaceDataFromPrompt(apiServerConfig, suggestedName, workspaceNames, !multiProject)
	if err != nil {
		return "", nil, err
	}

	if workspaceName == "" {
		return "", nil, errors.New("workspace name is required")
	}

	if primaryContainerImage != "" {
		projectList[0].Image = &primaryContainerImage
	}

	if primaryContainerUser != "" {
		projectList[0].User = &primaryContainerUser
	}

	if multiProject {
		create.DisplayMultiSubmitForm(workspaceName, &projectList, apiServerConfig, &doneCheck)
		if !doneCheck {
			return "", nil, errors.New("operation cancelled")
		}
	}

	return workspaceName, projectList, nil
}

func GetProjectNameFromRepo(repoUrl string) string {
	projectNameSlugRegex := regexp.MustCompile(`[^a-zA-Z0-9-]`)
	return projectNameSlugRegex.ReplaceAllString(strings.TrimSuffix(strings.ToLower(filepath.Base(repoUrl)), ".git"), "-")
}
