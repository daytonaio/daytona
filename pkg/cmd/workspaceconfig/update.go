// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaceconfig

import (
	"context"
	"net/http"
	"net/url"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/selection"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/daytonaio/daytona/pkg/views/workspace/create"
	"github.com/spf13/cobra"
)

var workspaceConfigUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a workspace config",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var workspaceConfig *apiclient.WorkspaceConfig
		var res *http.Response
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		if len(args) == 0 {
			workspaceConfigList, res, err := apiClient.WorkspaceConfigAPI.ListWorkspaceConfigs(ctx).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			if len(workspaceConfigList) == 0 {
				views_util.NotifyEmptyWorkspaceConfigList(true)
				return nil
			}

			workspaceConfig = selection.GetWorkspaceConfigFromPrompt(workspaceConfigList, 0, false, false, "Update")
			if workspaceConfig == nil {
				return nil
			}
		} else {
			workspaceConfig, res, err = apiClient.WorkspaceConfigAPI.GetWorkspaceConfig(ctx, args[0]).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}
		}

		if workspaceConfig == nil {
			return nil
		}

		createDto := []apiclient.CreateWorkspaceDTO{
			{
				Name: workspaceConfig.Name,
				Source: apiclient.CreateWorkspaceSourceDTO{
					Repository: apiclient.GitRepository{
						Url: workspaceConfig.RepositoryUrl,
					},
				},
				BuildConfig:         workspaceConfig.BuildConfig,
				EnvVars:             workspaceConfig.EnvVars,
				GitProviderConfigId: workspaceConfig.GitProviderConfigId,
			},
		}

		workspaceDefaults := &views_util.WorkspaceConfigDefaults{
			BuildChoice: views_util.AUTOMATIC,
			Image:       &workspaceConfig.Image,
			ImageUser:   &workspaceConfig.User,
		}

		if workspaceConfig.BuildConfig != nil && workspaceConfig.BuildConfig.Devcontainer != nil {
			workspaceDefaults.DevcontainerFilePath = workspaceConfig.BuildConfig.Devcontainer.FilePath
		}

		_, err = create.RunWorkspaceConfiguration(&createDto, *workspaceDefaults)
		if err != nil {
			return err
		}

		eligibleGitProviders, res, err := apiClient.GitProviderAPI.ListGitProvidersForUrl(ctx, url.QueryEscape(workspaceConfig.RepositoryUrl)).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		if len(eligibleGitProviders) > 0 {
			selectedGitProvider := selection.GetGitProviderConfigFromPrompt(selection.GetGitProviderConfigParams{
				GitProviderConfigs:       eligibleGitProviders,
				ActionVerb:               "Use",
				WithNoneOption:           true,
				PreselectedGitProviderId: workspaceConfig.GitProviderConfigId,
			})

			if selectedGitProvider == nil {
				return nil
			}

			if selectedGitProvider.Id == selection.NoneGitProviderConfigIdentifier {
				createDto[0].GitProviderConfigId = nil
			} else {
				createDto[0].GitProviderConfigId = &selectedGitProvider.Id
			}
		}

		newWorkspaceConfig := apiclient.CreateWorkspaceConfigDTO{
			Name:                workspaceConfig.Name,
			BuildConfig:         createDto[0].BuildConfig,
			Image:               createDto[0].Image,
			User:                createDto[0].User,
			RepositoryUrl:       createDto[0].Source.Repository.Url,
			EnvVars:             createDto[0].EnvVars,
			GitProviderConfigId: createDto[0].GitProviderConfigId,
		}

		res, err = apiClient.WorkspaceConfigAPI.SetWorkspaceConfig(ctx).WorkspaceConfig(newWorkspaceConfig).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		views.RenderInfoMessage("Workspace config updated successfully")
		return nil
	},
}
