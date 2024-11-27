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
	cmd_common "github.com/daytonaio/daytona/pkg/cmd/common"
	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/views/selection"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
)

type ProcessCmdArgumentsParams struct {
	ApiClient                   *apiclient.APIClient
	RepoUrls                    []string
	CreateWorkspaceDtos         *[]apiclient.CreateWorkspaceDTO
	ExistingWorkspaces          *[]apiclient.WorkspaceDTO
	WorkspaceConfigurationFlags cmd_common.WorkspaceConfigurationFlags
	BlankFlag                   bool
}

type ProcessGitUrlParams struct {
	ApiClient                   *apiclient.APIClient
	RepoUrl                     string
	CreateWorkspaceDtos         *[]apiclient.CreateWorkspaceDTO
	WorkspaceConfigurationFlags cmd_common.WorkspaceConfigurationFlags
	Branch                      *string
	BlankFlag                   bool
}

func ProcessCmdArguments(ctx context.Context, params ProcessCmdArgumentsParams) ([]string, error) {
	if len(params.RepoUrls) == 0 {
		return nil, fmt.Errorf("no repository URLs provided")
	}

	if len(params.RepoUrls) > 1 && cmd_common.CheckAnyWorkspaceConfigurationFlagSet(params.WorkspaceConfigurationFlags) {
		return nil, fmt.Errorf("can't set custom workspace configuration properties for multiple workspaces")
	}

	if *params.WorkspaceConfigurationFlags.Builder != "" && *params.WorkspaceConfigurationFlags.Builder != views_util.DEVCONTAINER && *params.WorkspaceConfigurationFlags.DevcontainerPath != "" {
		return nil, fmt.Errorf("can't set devcontainer file path if builder is not set to %s", views_util.DEVCONTAINER)
	}

	var workspaceConfig *apiclient.WorkspaceConfig

	existingWorkspaceConfigNames := []string{}

	for i, repoUrl := range params.RepoUrls {
		var branch *string
		if len(*params.WorkspaceConfigurationFlags.Branches) > i {
			branch = &(*params.WorkspaceConfigurationFlags.Branches)[i]
		}

		validatedUrl, err := util.GetValidatedUrl(repoUrl)
		if err == nil {
			// The argument is a Git URL
			existingWorkspaceConfigName, err := processGitURL(ctx, ProcessGitUrlParams{
				ApiClient:                   params.ApiClient,
				RepoUrl:                     validatedUrl,
				CreateWorkspaceDtos:         params.CreateWorkspaceDtos,
				WorkspaceConfigurationFlags: params.WorkspaceConfigurationFlags,
				Branch:                      branch,
				BlankFlag:                   params.BlankFlag,
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
		workspaceConfig, _, err = params.ApiClient.WorkspaceConfigAPI.GetWorkspaceConfig(ctx, repoUrl).Execute()
		if err != nil {
			return nil, fmt.Errorf("failed to parse the URL or fetch the workspace config for '%s'", repoUrl)
		}

		existingWorkspaceConfigName, err := AddWorkspaceFromConfig(ctx, AddWorkspaceFromConfigParams{
			WorkspaceConfig: workspaceConfig,
			ApiClient:       params.ApiClient,
			Workspaces:      params.CreateWorkspaceDtos,
			BranchFlag:      branch,
		})
		if err != nil {
			return nil, err
		}
		if existingWorkspaceConfigName != nil {
			existingWorkspaceConfigNames = append(existingWorkspaceConfigNames, *existingWorkspaceConfigName)
		} else {
			existingWorkspaceConfigNames = append(existingWorkspaceConfigNames, "")
		}
	}

	generateWorkspaceIds(params.CreateWorkspaceDtos)
	setInitialWorkspaceNames(params.CreateWorkspaceDtos, *params.ExistingWorkspaces)

	return existingWorkspaceConfigNames, nil
}

func processGitURL(ctx context.Context, params ProcessGitUrlParams) (*string, error) {
	encodedURLParam := url.QueryEscape(params.RepoUrl)

	if !params.BlankFlag {
		workspaceConfig, res, err := params.ApiClient.WorkspaceConfigAPI.GetDefaultWorkspaceConfig(ctx, encodedURLParam).Execute()
		if err == nil {
			workspaceConfig.GitProviderConfigId = params.WorkspaceConfigurationFlags.GitProviderConfig
			return AddWorkspaceFromConfig(ctx, AddWorkspaceFromConfigParams{
				WorkspaceConfig: workspaceConfig,
				ApiClient:       params.ApiClient,
				Workspaces:      params.CreateWorkspaceDtos,
				BranchFlag:      params.Branch,
			})
		}

		if res.StatusCode != http.StatusNotFound {
			return nil, apiclient_util.HandleErrorResponse(res, err)
		}
	}

	repo, res, err := params.ApiClient.GitProviderAPI.GetGitContext(ctx).Repository(apiclient.GetRepositoryContext{
		Url:    params.RepoUrl,
		Branch: params.Branch,
	}).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	workspaceName, err := GetSanitizedWorkspaceName(repo.Name)
	if err != nil {
		return nil, err
	}

	params.WorkspaceConfigurationFlags.GitProviderConfig, err = GetGitProviderConfigIdFromFlag(ctx, params.ApiClient, params.WorkspaceConfigurationFlags.GitProviderConfig)
	if err != nil {
		return nil, err
	}

	if params.WorkspaceConfigurationFlags.GitProviderConfig == nil || *params.WorkspaceConfigurationFlags.GitProviderConfig == "" {
		gitProviderConfigs, res, err := params.ApiClient.GitProviderAPI.ListGitProvidersForUrl(ctx, url.QueryEscape(params.RepoUrl)).Execute()
		if err != nil {
			return nil, apiclient_util.HandleErrorResponse(res, err)
		}

		if len(gitProviderConfigs) == 1 {
			params.WorkspaceConfigurationFlags.GitProviderConfig = &gitProviderConfigs[0].Id
		} else if len(gitProviderConfigs) > 1 {
			gp := selection.GetGitProviderConfigFromPrompt(selection.GetGitProviderConfigParams{
				GitProviderConfigs: gitProviderConfigs,
				ActionVerb:         "Use",
			})
			if gp == nil {
				return nil, common.ErrCtrlCAbort
			}
			params.WorkspaceConfigurationFlags.GitProviderConfig = &gp.Id
		}
	}

	workspace, err := GetCreateWorkspaceDtoFromFlags(params.WorkspaceConfigurationFlags)
	if err != nil {
		return nil, err
	}

	workspace.Name = workspaceName
	workspace.Source = apiclient.CreateWorkspaceSourceDTO{
		Repository: *repo,
	}

	*params.CreateWorkspaceDtos = append(*params.CreateWorkspaceDtos, *workspace)

	return nil, nil
}
