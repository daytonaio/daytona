// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"fmt"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views/workspace/create"
)

type CreateDataPromptConfig struct {
	ApiServerConfig        *apiclient.ServerConfig
	ExistingWorkspaceNames []string
	UserGitProviders       []apiclient.GitProvider
	Manual                 bool
	MultiProject           bool
	ApiClient              *apiclient.APIClient
}

func GetCreationDataFromPrompt(config CreateDataPromptConfig) (string, []apiclient.CreateWorkspaceRequestProject, error) {
	var projectList []apiclient.CreateWorkspaceRequestProject
	var providerRepo *apiclient.GitRepository
	var err error
	var workspaceName string

	if !config.Manual && config.UserGitProviders != nil && len(config.UserGitProviders) > 0 {
		providerRepo, err = getRepositoryFromWizard(config.UserGitProviders, 0)
		if err != nil {
			return "", nil, err
		}
	}

	if providerRepo == nil {
		providerRepo, err = create.GetRepositoryFromUrlInput(config.MultiProject, config.ApiClient)
		if err != nil {
			return "", nil, err
		}
	}

	projectList = []apiclient.CreateWorkspaceRequestProject{{
		Name: *providerRepo.Name,
		Source: &apiclient.CreateWorkspaceRequestProjectSource{
			Repository: providerRepo,
		},
	}}

	if config.MultiProject {
		addMore := true
		for i := 2; addMore; i++ {
			var providerRepo *apiclient.GitRepository

			if !config.Manual && config.UserGitProviders != nil && len(config.UserGitProviders) > 0 {
				providerRepo, err = getRepositoryFromWizard(config.UserGitProviders, i+2)
				if err != nil {
					return "", nil, err
				}
			}

			if providerRepo == nil {
				providerRepo, addMore, err = create.RunAdditionalProjectRepoForm(i, config.ApiClient)
				if err != nil {
					return "", nil, err
				}
			} else {
				addMore, err = create.RunAddMoreProjectsForm()
				if err != nil {
					return "", nil, err
				}
			}

			projectList = append(projectList, apiclient.CreateWorkspaceRequestProject{
				Name: *providerRepo.Name,
				Source: &apiclient.CreateWorkspaceRequestProjectSource{
					Repository: providerRepo,
				},
			})
		}
	}

	suggestedName := GetSuggestedWorkspaceName(projectList[0].Name, config.ExistingWorkspaceNames)

	err = create.RunSubmissionForm(&workspaceName, suggestedName, config.ExistingWorkspaceNames, &projectList, config.ApiServerConfig)
	if err != nil {
		return "", nil, err
	}

	return workspaceName, projectList, nil
}

func GetProjectNameFromRepo(repoUrl string) string {
	projectNameSlugRegex := regexp.MustCompile(`[^a-zA-Z0-9-]`)
	return projectNameSlugRegex.ReplaceAllString(strings.TrimSuffix(strings.ToLower(filepath.Base(repoUrl)), ".git"), "-")
}

func GetSuggestedWorkspaceName(firstProjectName string, existingWorkspaceNames []string) string {
	suggestion := firstProjectName

	if !slices.Contains(existingWorkspaceNames, suggestion) {
		return suggestion
	} else {
		i := 2
		for {
			newSuggestion := fmt.Sprintf("%s%d", suggestion, i)
			if !slices.Contains(existingWorkspaceNames, newSuggestion) {
				return newSuggestion
			}
			i++
		}
	}
}
