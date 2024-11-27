// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package create

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
	cmd_common "github.com/daytonaio/daytona/pkg/cmd/common"
	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/views/selection"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/daytonaio/daytona/pkg/views/workspace/create"
	"github.com/docker/docker/pkg/stringid"
)

type WorkspacesDataPromptParams struct {
	UserGitProviders    []apiclient.GitProvider
	WorkspaceConfigs    []apiclient.WorkspaceConfig
	Manual              bool
	SkipBranchSelection bool
	MultiWorkspace      bool
	BlankWorkspace      bool
	ApiClient           *apiclient.APIClient
	Defaults            *views_util.WorkspaceConfigDefaults
}

func GetWorkspacesCreationDataFromPrompt(ctx context.Context, params WorkspacesDataPromptParams) ([]apiclient.CreateWorkspaceDTO, error) {
	var workspaceList []apiclient.CreateWorkspaceDTO
	// keep track of visited repos, will help in keeping workspace names unique
	// since these are later saved into the db under a unique constraint field.
	selectedRepos := make(map[string]int)

	for i := 1; params.MultiWorkspace || i == 1; i++ {
		var err error

		if i > 2 {
			addMore, err := create.RunAddMoreWorkspacesForm()
			if err != nil {
				return nil, err
			}
			if !addMore {
				break
			}
		}

		if len(params.WorkspaceConfigs) > 0 && !params.BlankWorkspace {
			workspaceConfig := selection.GetWorkspaceConfigFromPrompt(params.WorkspaceConfigs, i, true, false, "Use")
			if workspaceConfig == nil {
				return nil, common.ErrCtrlCAbort
			}

			workspaceNames := []string{}
			for _, w := range workspaceList {
				workspaceNames = append(workspaceNames, w.Name)
			}

			// Append occurence number to keep duplicate entries unique
			repoUrl := workspaceConfig.RepositoryUrl
			if len(selectedRepos) > 0 && selectedRepos[repoUrl] > 1 {
				workspaceConfig.Name += strconv.Itoa(selectedRepos[repoUrl])
			}

			if workspaceConfig.Name != selection.BlankWorkspaceIdentifier {
				workspaceName := GetSuggestedName(workspaceConfig.Name, workspaceNames)

				getRepoContext := apiclient.GetRepositoryContext{
					Url: workspaceConfig.RepositoryUrl,
				}

				branch, err := GetBranchFromWorkspaceConfig(ctx, workspaceConfig, params.ApiClient, i)
				if err != nil {
					return nil, err
				}

				if branch != nil {
					getRepoContext.Branch = &branch.Name
					getRepoContext.Sha = &branch.Sha
				}

				configRepo, res, err := params.ApiClient.GitProviderAPI.GetGitContext(ctx).Repository(getRepoContext).Execute()
				if err != nil {
					return nil, apiclient_util.HandleErrorResponse(res, err)
				}

				createWorkspaceDto := apiclient.CreateWorkspaceDTO{
					Name:                workspaceName,
					GitProviderConfigId: workspaceConfig.GitProviderConfigId,
					Source: apiclient.CreateWorkspaceSourceDTO{
						Repository: *configRepo,
					},
					BuildConfig: workspaceConfig.BuildConfig,
					Image:       params.Defaults.Image,
					User:        params.Defaults.ImageUser,
					EnvVars:     workspaceConfig.EnvVars,
				}

				if workspaceConfig.Image != "" {
					createWorkspaceDto.Image = &workspaceConfig.Image
				}

				if workspaceConfig.User != "" {
					createWorkspaceDto.User = &workspaceConfig.User
				}

				if workspaceConfig.GitProviderConfigId == nil || *workspaceConfig.GitProviderConfigId == "" {
					gitProviderConfigId, res, err := params.ApiClient.GitProviderAPI.GetGitProviderIdForUrl(ctx, url.QueryEscape(workspaceConfig.RepositoryUrl)).Execute()
					if err != nil {
						return nil, apiclient_util.HandleErrorResponse(res, err)
					}
					createWorkspaceDto.GitProviderConfigId = &gitProviderConfigId
				}

				workspaceList = append(workspaceList, createWorkspaceDto)
				continue
			}
		}

		providerRepo, gitProviderConfigId, err := getRepositoryFromWizard(ctx, RepositoryWizardParams{
			ApiClient:           params.ApiClient,
			UserGitProviders:    params.UserGitProviders,
			Manual:              params.Manual,
			MultiWorkspace:      params.MultiWorkspace,
			SkipBranchSelection: params.SkipBranchSelection,
			WorkspaceOrder:      i,
			SelectedRepos:       selectedRepos,
		})
		if err != nil {
			return nil, err
		}

		if gitProviderConfigId == selection.CustomRepoIdentifier || gitProviderConfigId == selection.CREATE_FROM_SAMPLE {
			gitProviderConfigs, res, err := params.ApiClient.GitProviderAPI.ListGitProvidersForUrl(ctx, url.QueryEscape(providerRepo.Url)).Execute()
			if err != nil {
				return nil, apiclient_util.HandleErrorResponse(res, err)
			}

			if len(gitProviderConfigs) == 1 {
				gitProviderConfigId = gitProviderConfigs[0].Id
			} else if len(gitProviderConfigs) > 1 {
				gp := selection.GetGitProviderConfigFromPrompt(selection.GetGitProviderConfigParams{
					GitProviderConfigs: gitProviderConfigs,
					ActionVerb:         "Use",
				})
				if gp == nil {
					return nil, common.ErrCtrlCAbort
				}
				gitProviderConfigId = gp.Id
			} else {
				gitProviderConfigId = ""
			}
		}

		getRepoContext := createGetRepoContextFromRepository(providerRepo)

		var res *http.Response
		providerRepo, res, err = params.ApiClient.GitProviderAPI.GetGitContext(ctx).Repository(getRepoContext).Execute()
		if err != nil {
			return nil, apiclient_util.HandleErrorResponse(res, err)
		}

		providerRepoName, err := GetSanitizedWorkspaceName(providerRepo.Name)
		if err != nil {
			return nil, err
		}

		workspaceList = append(workspaceList, newCreateWorkspaceConfigDTO(params, providerRepo, providerRepoName, gitProviderConfigId))
	}

	return workspaceList, nil
}

func GetWorkspaceNameFromRepo(repoUrl string) string {
	workspaceNameSlugRegex := regexp.MustCompile(`[^a-zA-Z0-9-]`)
	return workspaceNameSlugRegex.ReplaceAllString(strings.TrimSuffix(strings.ToLower(filepath.Base(repoUrl)), ".git"), "-")
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

func GetSanitizedWorkspaceName(workspaceName string) (string, error) {
	workspaceName, err := url.QueryUnescape(workspaceName)
	if err != nil {
		return "", err
	}
	workspaceName = strings.ReplaceAll(workspaceName, " ", "-")

	return workspaceName, nil
}

func GetBranchFromWorkspaceConfig(ctx context.Context, workspaceConfig *apiclient.WorkspaceConfig, apiClient *apiclient.APIClient, workspaceOrder int) (*apiclient.GitBranch, error) {
	encodedURLParam := url.QueryEscape(workspaceConfig.RepositoryUrl)

	repoResponse, res, err := apiClient.GitProviderAPI.GetGitContext(ctx).Repository(apiclient.GetRepositoryContext{
		Url: workspaceConfig.RepositoryUrl,
	}).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	gitProviderConfigId, res, err := apiClient.GitProviderAPI.GetGitProviderIdForUrl(ctx, encodedURLParam).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	branchWizardConfig := BranchWizardParams{
		ApiClient:           apiClient,
		GitProviderConfigId: gitProviderConfigId,
		NamespaceId:         repoResponse.Owner,
		ChosenRepo:          repoResponse,
		WorkspaceOrder:      workspaceOrder,
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

func GetCreateWorkspaceDtoFromFlags(workspaceConfigurationFlags cmd_common.WorkspaceConfigurationFlags) (*apiclient.CreateWorkspaceDTO, error) {
	workspace := &apiclient.CreateWorkspaceDTO{
		GitProviderConfigId: workspaceConfigurationFlags.GitProviderConfig,
		BuildConfig:         &apiclient.BuildConfig{},
	}

	if *workspaceConfigurationFlags.Builder == views_util.DEVCONTAINER || *workspaceConfigurationFlags.DevcontainerPath != "" {
		devcontainerFilePath := create.DEVCONTAINER_FILEPATH
		if *workspaceConfigurationFlags.DevcontainerPath != "" {
			devcontainerFilePath = *workspaceConfigurationFlags.DevcontainerPath
		}
		workspace.BuildConfig.Devcontainer = &apiclient.DevcontainerConfig{
			FilePath: devcontainerFilePath,
		}

	}

	if *workspaceConfigurationFlags.Builder == views_util.NONE || *workspaceConfigurationFlags.CustomImage != "" || *workspaceConfigurationFlags.CustomImageUser != "" {
		workspace.BuildConfig = nil
		if *workspaceConfigurationFlags.CustomImage != "" || *workspaceConfigurationFlags.CustomImageUser != "" {
			workspace.Image = workspaceConfigurationFlags.CustomImage
			workspace.User = workspaceConfigurationFlags.CustomImageUser
		}
	}

	envVars := make(map[string]string)

	for _, envVar := range *workspaceConfigurationFlags.EnvVars {
		parts := strings.SplitN(envVar, "=", 2)
		if len(parts) == 2 {
			envVars[parts[0]] = parts[1]
		} else {
			return nil, fmt.Errorf("Invalid environment variable format: %s\n", envVar)
		}
	}

	workspace.EnvVars = envVars

	return workspace, nil
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

func newCreateWorkspaceConfigDTO(params WorkspacesDataPromptParams, providerRepo *apiclient.GitRepository, providerRepoName string, gitProviderConfigId string) apiclient.CreateWorkspaceDTO {
	workspace := apiclient.CreateWorkspaceDTO{
		Name:                providerRepoName,
		GitProviderConfigId: &gitProviderConfigId,
		Source: apiclient.CreateWorkspaceSourceDTO{
			Repository: *providerRepo,
		},
		BuildConfig: &apiclient.BuildConfig{},
		Image:       params.Defaults.Image,
		User:        params.Defaults.ImageUser,
		EnvVars:     map[string]string{},
	}

	return workspace
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

func setInitialWorkspaceNames(createWorkspaceDtos *[]apiclient.CreateWorkspaceDTO, existingWorkspaces []apiclient.WorkspaceDTO) {
	existingNames := make(map[string]bool)
	for _, workspace := range existingWorkspaces {
		existingNames[workspace.Name] = true
	}

	for i := range *createWorkspaceDtos {
		originalName := (*createWorkspaceDtos)[i].Name
		newName := originalName
		counter := 2

		for existingNames[newName] {
			newName = fmt.Sprintf("%s%d", originalName, counter)
			counter++
		}

		(*createWorkspaceDtos)[i].Name = newName
		existingNames[newName] = true
	}
}

func generateWorkspaceIds(createWorkspaceDtos *[]apiclient.CreateWorkspaceDTO) []string {
	for i := range *createWorkspaceDtos {
		wsId := stringid.GenerateRandomID()
		wsId = stringid.TruncateID(wsId)
		(*createWorkspaceDtos)[i].Id = wsId
	}

	return nil
}
