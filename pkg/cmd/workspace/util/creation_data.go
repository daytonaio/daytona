// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views/workspace/create"
)

type CmdCustoms struct {
	BuildChoice          create.BuilderChoice
	Image                string
	ImageUser            string
	DevcontainerFilePath string
}

type CreateDataPromptConfig struct {
	ApiServerConfig        *apiclient.ServerConfig
	ExistingWorkspaceNames []string
	UserGitProviders       []apiclient.GitProvider
	Manual                 bool
	MultiProject           bool
	ApiClient              *apiclient.APIClient
	Customs                *CmdCustoms
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

	providerRepoName, err := GetSanitizedProjectName(*providerRepo.Name)
	if err != nil {
		return "", nil, err
	}

	projectList = initializeProjectList(config, providerRepo, providerRepoName)

	if config.MultiProject {
		addMore := true
		for i := 2; addMore; i++ {
			var providerRepo *apiclient.GitRepository

			if !config.Manual && config.UserGitProviders != nil && len(config.UserGitProviders) > 0 {
				providerRepo, err = getRepositoryFromWizard(config.UserGitProviders, i)
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

			providerRepoName, err := GetSanitizedProjectName(*providerRepo.Name)
			if err != nil {
				return "", nil, err
			}

			projectList = append(projectList, apiclient.CreateWorkspaceRequestProject{
				Name: providerRepoName,
				Source: &apiclient.CreateWorkspaceRequestProjectSource{
					Repository: providerRepo,
				},
				Build:             &apiclient.ProjectBuild{},
				Image:             config.ApiServerConfig.DefaultProjectImage,
				User:              config.ApiServerConfig.DefaultProjectUser,
				PostStartCommands: config.ApiServerConfig.DefaultProjectPostStartCommands,
				EnvVars:           &map[string]string{},
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

func initializeProjectList(config CreateDataPromptConfig, providerRepo *apiclient.GitRepository, providerRepoName string) []apiclient.CreateWorkspaceRequestProject {
	project := apiclient.CreateWorkspaceRequestProject{
		Name: providerRepoName,
		Source: &apiclient.CreateWorkspaceRequestProjectSource{
			Repository: providerRepo,
		},
		Build:             &apiclient.ProjectBuild{},
		Image:             config.ApiServerConfig.DefaultProjectImage,
		User:              config.ApiServerConfig.DefaultProjectUser,
		PostStartCommands: config.ApiServerConfig.DefaultProjectPostStartCommands,
		EnvVars:           &map[string]string{},
	}

	if config.Customs != nil {
		if config.Customs.BuildChoice == create.DEVCONTAINER || config.Customs.DevcontainerFilePath != "" {
			project.Image = nil
			project.User = nil
			project.PostStartCommands = nil

			devcontainerFilePath := create.DEVCONTAINER_FILEPATH
			if config.Customs.DevcontainerFilePath != "" {
				devcontainerFilePath = config.Customs.DevcontainerFilePath
			}
			project.Build.Devcontainer = &apiclient.ProjectBuildDevcontainer{
				DevContainerFilePath: &devcontainerFilePath,
			}
		}

		if config.Customs.BuildChoice == create.NONE || config.Customs.Image != "" || config.Customs.ImageUser != "" {
			project.Build = nil
			if config.Customs.Image != "" || config.Customs.ImageUser != "" {
				project.Image = &config.Customs.Image
				project.User = &config.Customs.ImageUser
			}
		}
	}

	return []apiclient.CreateWorkspaceRequestProject{project}

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

func GetSanitizedProjectName(projectName string) (string, error) {
	projectName, err := url.QueryUnescape(projectName)
	if err != nil {
		return "", err
	}
	projectName = strings.ReplaceAll(projectName, " ", "-")

	return projectName, nil
}
