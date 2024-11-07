// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package create

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	workspace_common "github.com/daytonaio/daytona/pkg/cmd/workspace/common"
	"github.com/daytonaio/daytona/pkg/views/selection"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
)

type ProcessCmdArgumentsConfig struct {
	ApiClient                   *apiclient.APIClient
	RepoUrls                    []string
	CreateWorkspaceDtos         *[]apiclient.CreateWorkspaceDTO
	ExistingWorkspaces          *[]apiclient.WorkspaceDTO
	WorkspaceConfigurationFlags workspace_common.WorkspaceConfigurationFlags
	BlankFlag                   bool
}

type ProcessGitUrlConfig struct {
	ApiClient                   *apiclient.APIClient
	RepoUrl                     string
	CreateWorkspaceDtos         *[]apiclient.CreateWorkspaceDTO
	WorkspaceConfigurationFlags workspace_common.WorkspaceConfigurationFlags
	Branch                      *string
	BlankFlag                   bool
}

func ProcessCmdArguments(ctx context.Context, config ProcessCmdArgumentsConfig) ([]string, error) {
	if len(config.RepoUrls) == 0 {
		return nil, fmt.Errorf("no repository URLs provided")
	}

	if len(config.RepoUrls) > 1 && workspace_common.CheckAnyWorkspaceConfigurationFlagSet(config.WorkspaceConfigurationFlags) {
		return nil, fmt.Errorf("can't set custom workspace configuration properties for multiple workspaces")
	}

	if *config.WorkspaceConfigurationFlags.Builder != "" && *config.WorkspaceConfigurationFlags.Builder != views_util.DEVCONTAINER && *config.WorkspaceConfigurationFlags.DevcontainerPath != "" {
		return nil, fmt.Errorf("can't set devcontainer file path if builder is not set to %s", views_util.DEVCONTAINER)
	}

	var workspaceConfig *apiclient.WorkspaceConfig

	existingWorkspaceConfigNames := []string{}

	for i, repoUrl := range config.RepoUrls {
		var branch *string
		if len(*config.WorkspaceConfigurationFlags.Branches) > i {
			branch = &(*config.WorkspaceConfigurationFlags.Branches)[i]
		}

		validatedUrl, err := util.GetValidatedUrl(repoUrl)
		if err == nil {
			// The argument is a Git URL
			existingWorkspaceConfigName, err := processGitURL(ctx, ProcessGitUrlConfig{
				ApiClient:                   config.ApiClient,
				RepoUrl:                     validatedUrl,
				CreateWorkspaceDtos:         config.CreateWorkspaceDtos,
				WorkspaceConfigurationFlags: config.WorkspaceConfigurationFlags,
				Branch:                      branch,
				BlankFlag:                   config.BlankFlag,
			})
			if err != nil {
				return nil, err
			}
			if existingWorkspaceConfigName != nil {
				existingWorkspaceConfigNames = append(existingWorkspaceConfigNames, *existingWorkspaceConfigName)
			} else {
				existingWorkspaceConfigNames = append(existingWorkspaceConfigNames, "")
			}

			continue
		}

		// The argument is not a Git URL - try getting the workspace config
		workspaceConfig, _, err = config.ApiClient.WorkspaceConfigAPI.GetWorkspaceConfig(ctx, repoUrl).Execute()
		if err != nil {
			return nil, fmt.Errorf("failed to parse the URL or fetch the workspace config for '%s'", repoUrl)
		}

		existingWorkspaceConfigName, err := AddWorkspaceFromConfig(workspaceConfig, config.ApiClient, config.CreateWorkspaceDtos, branch)
		if err != nil {
			return nil, err
		}
		if existingWorkspaceConfigName != nil {
			existingWorkspaceConfigNames = append(existingWorkspaceConfigNames, *existingWorkspaceConfigName)
		} else {
			existingWorkspaceConfigNames = append(existingWorkspaceConfigNames, "")
		}
	}

	generateWorkspaceIds(config.CreateWorkspaceDtos)
	setInitialWorkspaceNames(config.CreateWorkspaceDtos, *config.ExistingWorkspaces)

	return existingWorkspaceConfigNames, nil
}

func processGitURL(ctx context.Context, config ProcessGitUrlConfig) (*string, error) {
	encodedURLParam := url.QueryEscape(config.RepoUrl)

	if !config.BlankFlag {
		workspaceConfig, res, err := config.ApiClient.WorkspaceConfigAPI.GetDefaultWorkspaceConfig(ctx, encodedURLParam).Execute()
		if err == nil {
			workspaceConfig.GitProviderConfigId = config.WorkspaceConfigurationFlags.GitProviderConfig
			return AddWorkspaceFromConfig(workspaceConfig, config.ApiClient, config.CreateWorkspaceDtos, config.Branch)
		}

		if res.StatusCode != http.StatusNotFound {
			return nil, apiclient_util.HandleErrorResponse(res, err)
		}
	}

	repo, res, err := config.ApiClient.GitProviderAPI.GetGitContext(ctx).Repository(apiclient.GetRepositoryContext{
		Url:    config.RepoUrl,
		Branch: config.Branch,
	}).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	workspaceName, err := GetSanitizedWorkspaceName(repo.Name)
	if err != nil {
		return nil, err
	}

	config.WorkspaceConfigurationFlags.GitProviderConfig, err = GetGitProviderConfigIdFromFlag(ctx, config.ApiClient, config.WorkspaceConfigurationFlags.GitProviderConfig)
	if err != nil {
		return nil, err
	}

	gitProviderConfigs, res, err := config.ApiClient.GitProviderAPI.ListGitProvidersForUrl(context.Background(), url.QueryEscape(config.RepoUrl)).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	if len(gitProviderConfigs) == 1 {
		config.WorkspaceConfigurationFlags.GitProviderConfig = &gitProviderConfigs[0].Id
	} else if len(gitProviderConfigs) > 1 {
		gp := selection.GetGitProviderConfigFromPrompt(selection.GetGitProviderConfigParams{
			GitProviderConfigs: gitProviderConfigs,
			ActionVerb:         "Use",
		})
		config.WorkspaceConfigurationFlags.GitProviderConfig = &gp.Id
	}

	workspace, err := GetCreateWorkspaceDtoFromFlags(config.WorkspaceConfigurationFlags)
	if err != nil {
		return nil, err
	}

	workspace.Name = workspaceName
	workspace.Source = apiclient.CreateWorkspaceSourceDTO{
		Repository: *repo,
	}

	*config.CreateWorkspaceDtos = append(*config.CreateWorkspaceDtos, *workspace)

	return nil, nil
}
