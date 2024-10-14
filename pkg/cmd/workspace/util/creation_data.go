// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/common"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/daytonaio/daytona/pkg/views/workspace/create"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
)

type ProjectsDataPromptConfig struct {
	UserGitProviders    []apiclient.GitProvider
	ProjectConfigs      []apiclient.ProjectConfig
	Manual              bool
	SkipBranchSelection bool
	MultiProject        bool
	BlankProject        bool
	ApiClient           *apiclient.APIClient
	Defaults            *views_util.ProjectConfigDefaults
}

func GetProjectsCreationDataFromPrompt(config ProjectsDataPromptConfig) ([]apiclient.CreateProjectDTO, error) {
	var projectList []apiclient.CreateProjectDTO
	// keep track of visited repos, will help in keeping project names unique
	// since these are later saved into the db under a unique constraint field.
	selectedRepos := make(map[string]int)

	for i := 1; config.MultiProject || i == 1; i++ {
		var err error

		if i > 2 {
			addMore, err := create.RunAddMoreProjectsForm()
			if err != nil {
				return nil, err
			}
			if !addMore {
				break
			}
		}

		if len(config.ProjectConfigs) > 0 && !config.BlankProject {
			projectConfig := selection.GetProjectConfigFromPrompt(config.ProjectConfigs, i, true, false, "Use")
			if projectConfig == nil {
				return nil, common.ErrCtrlCAbort
			}

			projectNames := []string{}
			for _, p := range projectList {
				projectNames = append(projectNames, p.Name)
			}

			// Append occurence number to keep duplicate entries unique
			repoUrl := projectConfig.RepositoryUrl
			if len(selectedRepos) > 0 && selectedRepos[repoUrl] > 1 {
				projectConfig.Name += strconv.Itoa(selectedRepos[repoUrl])
			}

			if projectConfig.Name != selection.BlankProjectIdentifier {
				projectName := GetSuggestedName(projectConfig.Name, projectNames)

				getRepoContext := apiclient.GetRepositoryContext{
					Url: projectConfig.RepositoryUrl,
				}

				branch, err := GetBranchFromProjectConfig(projectConfig, config.ApiClient, i)
				if err != nil {
					return nil, err
				}

				if branch != nil {
					getRepoContext.Branch = &branch.Name
					getRepoContext.Sha = &branch.Sha
				}

				configRepo, res, err := config.ApiClient.GitProviderAPI.GetGitContext(context.Background()).Repository(getRepoContext).Execute()
				if err != nil {
					return nil, apiclient_util.HandleErrorResponse(res, err)
				}

				createProjectDto := apiclient.CreateProjectDTO{
					Name:                projectName,
					GitProviderConfigId: projectConfig.GitProviderConfigId,
					Source: apiclient.CreateProjectSourceDTO{
						Repository: *configRepo,
					},
					BuildConfig: projectConfig.BuildConfig,
					Image:       config.Defaults.Image,
					User:        config.Defaults.ImageUser,
					EnvVars:     projectConfig.EnvVars,
				}

				if projectConfig.Image != "" {
					createProjectDto.Image = &projectConfig.Image
				}

				if projectConfig.User != "" {
					createProjectDto.User = &projectConfig.User
				}

				if projectConfig.GitProviderConfigId == nil || *projectConfig.GitProviderConfigId == "" {
					gitProviderConfigId, res, err := config.ApiClient.GitProviderAPI.GetGitProviderIdForUrl(context.Background(), url.QueryEscape(projectConfig.RepositoryUrl)).Execute()
					if err != nil {
						return nil, apiclient_util.HandleErrorResponse(res, err)
					}
					createProjectDto.GitProviderConfigId = &gitProviderConfigId
				}

				projectList = append(projectList, createProjectDto)
				continue
			}
		}

		providerRepo, gitProviderConfigId, err := getRepositoryFromWizard(RepositoryWizardConfig{
			ApiClient:           config.ApiClient,
			UserGitProviders:    config.UserGitProviders,
			Manual:              config.Manual,
			MultiProject:        config.MultiProject,
			SkipBranchSelection: config.SkipBranchSelection,
			ProjectOrder:        i,
			SelectedRepos:       selectedRepos,
		})
		if err != nil {
			return nil, err
		}

		if gitProviderConfigId == selection.CustomRepoIdentifier || gitProviderConfigId == selection.CREATE_FROM_SAMPLE {
			gitProviderConfigs, res, err := config.ApiClient.GitProviderAPI.ListGitProvidersForUrl(context.Background(), url.QueryEscape(providerRepo.Url)).Execute()
			if err != nil {
				return nil, apiclient_util.HandleErrorResponse(res, err)
			}

			if len(gitProviderConfigs) == 1 {
				gitProviderConfigId = gitProviderConfigs[0].Id
			} else if len(gitProviderConfigs) > 1 {
				gp := selection.GetGitProviderConfigFromPrompt(gitProviderConfigs, false, "Use")
				gitProviderConfigId = gp.Id
			}
		}

		getRepoContext := createGetRepoContextFromRepository(providerRepo)

		var res *http.Response
		providerRepo, res, err = config.ApiClient.GitProviderAPI.GetGitContext(context.Background()).Repository(getRepoContext).Execute()
		if err != nil {
			return nil, apiclient_util.HandleErrorResponse(res, err)
		}

		providerRepoName, err := GetSanitizedProjectName(providerRepo.Name)
		if err != nil {
			return nil, err
		}

		projectList = append(projectList, newCreateProjectConfigDTO(config, providerRepo, providerRepoName, gitProviderConfigId))
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

func GetBranchFromProjectConfig(projectConfig *apiclient.ProjectConfig, apiClient *apiclient.APIClient, projectOrder int) (*apiclient.GitBranch, error) {
	ctx := context.Background()

	encodedURLParam := url.QueryEscape(projectConfig.RepositoryUrl)

	repoResponse, res, err := apiClient.GitProviderAPI.GetGitContext(ctx).Repository(apiclient.GetRepositoryContext{
		Url: projectConfig.RepositoryUrl,
	}).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	gitProviderConfigId, res, err := apiClient.GitProviderAPI.GetGitProviderIdForUrl(ctx, encodedURLParam).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	branchWizardConfig := BranchWizardConfig{
		ApiClient:           apiClient,
		GitProviderConfigId: gitProviderConfigId,
		NamespaceId:         repoResponse.Owner,
		ChosenRepo:          repoResponse,
		ProjectOrder:        projectOrder,
	}

	repo, err := SetBranchFromWizard(branchWizardConfig)
	if err != nil {
		return nil, err
	}

	if repo == nil {
		return nil, common.ErrCtrlCAbort
	}

	return &apiclient.GitBranch{
		Name: repo.Branch,
		Sha:  repo.Sha,
	}, nil
}

func GetCreateProjectDtoFromFlags(projectConfigurationFlags ProjectConfigurationFlags) (*apiclient.CreateProjectDTO, error) {
	project := &apiclient.CreateProjectDTO{
		GitProviderConfigId: projectConfigurationFlags.GitProviderConfig,
		BuildConfig:         &apiclient.BuildConfig{},
	}

	if *projectConfigurationFlags.Builder == views_util.DEVCONTAINER || *projectConfigurationFlags.DevcontainerPath != "" {
		devcontainerFilePath := create.DEVCONTAINER_FILEPATH
		if *projectConfigurationFlags.DevcontainerPath != "" {
			devcontainerFilePath = *projectConfigurationFlags.DevcontainerPath
		}
		project.BuildConfig.Devcontainer = &apiclient.DevcontainerConfig{
			FilePath: devcontainerFilePath,
		}

	}

	if *projectConfigurationFlags.Builder == views_util.NONE || *projectConfigurationFlags.CustomImage != "" || *projectConfigurationFlags.CustomImageUser != "" {
		project.BuildConfig = nil
		if *projectConfigurationFlags.CustomImage != "" || *projectConfigurationFlags.CustomImageUser != "" {
			project.Image = projectConfigurationFlags.CustomImage
			project.User = projectConfigurationFlags.CustomImageUser
		}
	}

	envVars := make(map[string]string)

	for _, envVar := range *projectConfigurationFlags.EnvVars {
		parts := strings.SplitN(envVar, "=", 2)
		if len(parts) == 2 {
			envVars[parts[0]] = parts[1]
		} else {
			return nil, fmt.Errorf("Invalid environment variable format: %s\n", envVar)
		}
	}

	project.EnvVars = envVars

	return project, nil
}

func GetGitProviderConfigIdFromFlag(ctx context.Context, apiClient *apiclient.APIClient, gitProviderConfigFlag *string) (*string, error) {
	if gitProviderConfigFlag == nil || *gitProviderConfigFlag == "" {
		return gitProviderConfigFlag, nil
	}

	gitProviderConfigs, res, err := apiClient.GitProviderAPI.ListGitProviders(ctx).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	for _, gitProviderConfig := range gitProviderConfigs {
		if gitProviderConfig.Id == *gitProviderConfigFlag {
			return &gitProviderConfig.Id, nil
		}
		if gitProviderConfig.Alias == *gitProviderConfigFlag {
			return &gitProviderConfig.Id, nil
		}
	}

	return nil, fmt.Errorf("git provider config '%s' not found", *gitProviderConfigFlag)
}

func newCreateProjectConfigDTO(config ProjectsDataPromptConfig, providerRepo *apiclient.GitRepository, providerRepoName string, gitProviderConfigId string) apiclient.CreateProjectDTO {
	project := apiclient.CreateProjectDTO{
		Name:                providerRepoName,
		GitProviderConfigId: &gitProviderConfigId,
		Source: apiclient.CreateProjectSourceDTO{
			Repository: *providerRepo,
		},
		BuildConfig: &apiclient.BuildConfig{},
		Image:       config.Defaults.Image,
		User:        config.Defaults.ImageUser,
		EnvVars:     map[string]string{},
	}

	return project
}

func createGetRepoContextFromRepository(providerRepo *apiclient.GitRepository) apiclient.GetRepositoryContext {
	result := apiclient.GetRepositoryContext{
		Id:     &providerRepo.Id,
		Name:   &providerRepo.Name,
		Owner:  &providerRepo.Owner,
		Sha:    &providerRepo.Sha,
		Source: &providerRepo.Source,
		Url:    providerRepo.Url,
		Branch: &providerRepo.Branch,
	}

	if providerRepo.Path != nil {
		result.Path = providerRepo.Path
	}

	return result
}
