// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views/workspace/create"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
	log "github.com/sirupsen/logrus"
)

type ProjectsDataPromptConfig struct {
	UserGitProviders    []apiclient.GitProvider
	ProjectConfigs      []apiclient.ProjectConfig
	Manual              bool
	SkipBranchSelection bool
	MultiProject        bool
	BlankProject        bool
	ApiClient           *apiclient.APIClient
	Defaults            *create.ProjectDefaults
}

func GetProjectsCreationDataFromPrompt(config ProjectsDataPromptConfig) ([]apiclient.CreateProjectDTO, error) {
	var projectList []apiclient.CreateProjectDTO
	// keep track of visited repos, will help in keeping project names unique
	// since these are later saved into the db under a unique constraint field.
	selectedRepos := make(map[string]int)

	addMore := config.MultiProject

	for i := 1; addMore || i == 1; i++ {

		if len(config.ProjectConfigs) > 0 && !config.BlankProject {
			projectConfig := selection.GetProjectConfigFromPrompt(config.ProjectConfigs, i, true, "Use")
			if projectConfig == nil {
				return nil, fmt.Errorf("must select a project config")
			}

			projectNames := []string{}
			for _, p := range projectList {
				var currentName string
				if p.NewConfig.Name != nil {
					currentName = *p.NewConfig.Name
				} else if p.ExistingConfig != nil && p.ExistingConfig.ProjectName != nil {
					currentName = *p.ExistingConfig.ProjectName
				}
				projectNames = append(projectNames, currentName)
			}

			// Append occurence number to keep duplicate entries unique
			repoUrl := *projectConfig.Repository.Url
			if len(selectedRepos) > 0 && selectedRepos[repoUrl] > 1 {
				*projectConfig.Name += strconv.Itoa(selectedRepos[repoUrl])
			}

			if *projectConfig.Name != selection.BlankProjectIdentifier {
				projectName := GetSuggestedName(*projectConfig.Name, projectNames)

				branch, err := GetBranchFromProjectConfig(projectConfig, config.ApiClient, i)
				if err != nil {
					return nil, err
				}

				projectList = append(projectList, apiclient.CreateProjectDTO{
					ExistingConfig: &apiclient.ExistingConfigDTO{
						ConfigName:  projectConfig.Name,
						ProjectName: &projectName,
						Branch:      &branch,
					},
				})
				continue
			}
		}

		providerRepo, err := getRepositoryFromWizard(RepositoryWizardConfig{
			ApiClient:        config.ApiClient,
			UserGitProviders: config.UserGitProviders,
			MultiProject:     config.MultiProject,
			ProjectOrder:     i,
			SelectedRepos:    selectedRepos,
		})
		if err != nil {
			return nil, err
		}

		if i > 1 {
			addMore, err = create.RunAddMoreProjectsForm()
			if err != nil {
				return nil, err
			}
		}

		providerRepoName, err := GetSanitizedProjectName(*providerRepo.Name)
		if err != nil {
			return nil, err
		}

		projectList = append(projectList, newCreateProjectDTO(config, providerRepo, providerRepoName))
	}

	return projectList, nil
}

func GetProjectNameFromRepo(repoUrl string) string {
	projectNameSlugRegex := regexp.MustCompile(`[^a-zA-Z0-9-]`)
	return projectNameSlugRegex.ReplaceAllString(strings.TrimSuffix(strings.ToLower(filepath.Base(repoUrl)), ".git"), "-")
}

func GetSuggestedName(initialSuggestion string, existingNames []string) string {
	suggestion := initialSuggestion

	if !slices.Contains(existingNames, suggestion) {
		return suggestion
	} else {
		i := 2
		for {
			newSuggestion := fmt.Sprintf("%s%d", suggestion, i)
			if !slices.Contains(existingNames, newSuggestion) {
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

func GetEnvVariables(project *apiclient.CreateProjectDTO, profileData *apiclient.ProfileData) *map[string]string {
	envVars := map[string]string{}

	if profileData.EnvVars != nil {
		for k, v := range *profileData.EnvVars {
			if strings.HasPrefix(v, "$") {
				env, ok := os.LookupEnv(v[1:])
				if ok {
					envVars[k] = env
				} else {
					log.Warnf("Environment variable %s not found", v[1:])
				}
			} else {
				envVars[k] = v
			}
		}
	}

	if project.NewConfig.EnvVars != nil {
		for k, v := range *project.NewConfig.EnvVars {
			if strings.HasPrefix(v, "$") {
				env, ok := os.LookupEnv(v[1:])
				if ok {
					envVars[k] = env
				} else {
					log.Warnf("Environment variable %s not found", v[1:])
				}
			} else {
				envVars[k] = v
			}
		}
	}

	return &envVars
}

func GetBranchFromProjectConfig(projectConfig *apiclient.ProjectConfig, apiClient *apiclient.APIClient, projectOrder int) (string, error) {
	ctx := context.Background()

	encodedURLParam := url.QueryEscape(*projectConfig.Repository.Url)

	repoResponse, res, err := apiClient.GitProviderAPI.GetGitContext(ctx, encodedURLParam).Execute()
	if err != nil {
		return "", apiclient_util.HandleErrorResponse(res, err)
	}

	providerId, res, err := apiClient.GitProviderAPI.GetGitProviderIdForUrl(ctx, encodedURLParam).Execute()
	if err != nil {
		return "", apiclient_util.HandleErrorResponse(res, err)
	}

	branchWizardConfig := BranchWizardConfig{
		ApiClient:    apiClient,
		ProviderId:   providerId,
		NamespaceId:  *repoResponse.Owner,
		ChosenRepo:   repoResponse,
		ProjectOrder: projectOrder,
	}

	repo, err := GetBranchFromWizard(branchWizardConfig)
	if err != nil {
		return "", err
	}

	return *repo.Branch, nil
}

func newCreateProjectDTO(config ProjectsDataPromptConfig, providerRepo *apiclient.GitRepository, providerRepoName string) apiclient.CreateProjectDTO {
	project := apiclient.CreateProjectDTO{
		NewConfig: &apiclient.CreateProjectConfigDTO{
			Name: &providerRepoName,
			Source: &apiclient.CreateProjectConfigSourceDTO{
				Repository: providerRepo,
			},
			Build:   &apiclient.ProjectBuildConfig{},
			Image:   config.Defaults.Image,
			User:    config.Defaults.ImageUser,
			EnvVars: &map[string]string{},
		},
	}

	return project
}
