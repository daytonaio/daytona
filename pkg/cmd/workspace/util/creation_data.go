// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/views/workspace/create"
)

func GetCreationDataFromPrompt(apiServerConfig *serverapiclient.ServerConfig, existingWorkspaceNames []string, userGitProviders []serverapiclient.GitProvider, manual bool, multiProject bool) (string, []serverapiclient.CreateWorkspaceRequestProject, error) {
	var projectList []serverapiclient.CreateWorkspaceRequestProject
	var providerRepo serverapiclient.GitRepository
	var providerRepoUrl string
	var err error
	var workspaceName string

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

	suggestedName := GetSuggestedWorkspaceName(*workspaceCreationPromptResponse.PrimaryProject.Source.Repository.Url, existingWorkspaceNames)

	err = create.RunSubmissionForm(&workspaceName, suggestedName, existingWorkspaceNames, &projectList, apiServerConfig)
	if err != nil {
		return "", nil, err
	}

	return workspaceName, projectList, nil
}

func GetProjectNameFromRepo(repoUrl string) string {
	projectNameSlugRegex := regexp.MustCompile(`[^a-zA-Z0-9-]`)
	return projectNameSlugRegex.ReplaceAllString(strings.TrimSuffix(strings.ToLower(filepath.Base(repoUrl)), ".git"), "-")
}

func GetSuggestedWorkspaceName(repo string, existingWorkspaceNames []string) string {
	var result strings.Builder
	input := repo
	input = strings.TrimSuffix(input, "/")

	// Find the last index of '/' in the repo string
	lastIndex := strings.LastIndex(input, "/")

	// If '/' is found, extract the content after it
	if lastIndex != -1 && lastIndex < len(repo)-1 {
		input = repo[lastIndex+1:]
	}

	input = strings.TrimSuffix(input, ".git")

	for _, char := range input {
		if unicode.IsLetter(char) || unicode.IsNumber(char) || char == '-' {
			result.WriteRune(char)
		} else if char == ' ' {
			result.WriteRune('-')
		}
	}

	suggestion := strings.ToLower(result.String())

	if !checkContains(existingWorkspaceNames, suggestion) {
		return suggestion
	} else {
		i := 2
		for {
			newSuggestion := fmt.Sprintf("%s%d", suggestion, i)
			if !checkContains(existingWorkspaceNames, newSuggestion) {
				return newSuggestion
			}
			i++
		}
	}
}

func checkContains(arr []string, target string) bool {
	for _, s := range arr {
		if s == target {
			return true
		}
	}
	return false
}
