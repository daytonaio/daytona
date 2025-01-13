// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspacetemplate

import (
	"context"
	"net/http"
	"net/url"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	cmd_common "github.com/daytonaio/daytona/pkg/cmd/common"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/selection"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/daytonaio/daytona/pkg/views/workspace/create"
	"github.com/spf13/cobra"
)

var workspaceTemplateUpdateCmd = &cobra.Command{
	Use:     "update",
	Short:   "Update a workspace template",
	Args:    cobra.MaximumNArgs(1),
	Aliases: cmd_common.GetAliases("update"),
	RunE: func(cmd *cobra.Command, args []string) error {
		var workspaceTemplate *apiclient.WorkspaceTemplate
		var res *http.Response
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		if len(args) == 0 {
			workspaceTemplateList, res, err := apiClient.WorkspaceTemplateAPI.ListWorkspaceTemplates(ctx).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			if len(workspaceTemplateList) == 0 {
				views_util.NotifyEmptyWorkspaceTemplateList(true)
				return nil
			}

			workspaceTemplate = selection.GetWorkspaceTemplateFromPrompt(workspaceTemplateList, 0, false, false, "Update")
			if workspaceTemplate == nil {
				return nil
			}
		} else {
			workspaceTemplate, res, err = apiClient.WorkspaceTemplateAPI.FindWorkspaceTemplate(ctx, args[0]).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}
		}

		if workspaceTemplate == nil {
			return nil
		}

		createDto := []apiclient.CreateWorkspaceDTO{
			{
				Name: workspaceTemplate.Name,
				Source: apiclient.CreateWorkspaceSourceDTO{
					Repository: apiclient.GitRepository{
						Url: workspaceTemplate.RepositoryUrl,
					},
				},
				BuildConfig:         workspaceTemplate.BuildConfig,
				EnvVars:             workspaceTemplate.EnvVars,
				GitProviderConfigId: workspaceTemplate.GitProviderConfigId,
			},
		}

		workspaceDefaults := &views_util.WorkspaceTemplateDefaults{
			BuildChoice: views_util.AUTOMATIC,
			Image:       &workspaceTemplate.Image,
			ImageUser:   &workspaceTemplate.User,
		}

		if workspaceTemplate.BuildConfig != nil && workspaceTemplate.BuildConfig.Devcontainer != nil {
			workspaceDefaults.DevcontainerFilePath = workspaceTemplate.BuildConfig.Devcontainer.FilePath
		}

		_, err = create.RunWorkspaceConfiguration(&createDto, *workspaceDefaults, false)
		if err != nil {
			return err
		}

		eligibleGitProviders, res, err := apiClient.GitProviderAPI.ListGitProvidersForUrl(ctx, url.QueryEscape(workspaceTemplate.RepositoryUrl)).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		if len(eligibleGitProviders) > 0 {
			selectedGitProvider := selection.GetGitProviderConfigFromPrompt(selection.GetGitProviderConfigParams{
				GitProviderConfigs:       eligibleGitProviders,
				ActionVerb:               "Use",
				WithNoneOption:           true,
				PreselectedGitProviderId: workspaceTemplate.GitProviderConfigId,
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

		newWorkspaceTemplate := apiclient.CreateWorkspaceTemplateDTO{
			Name:                workspaceTemplate.Name,
			BuildConfig:         createDto[0].BuildConfig,
			Image:               createDto[0].Image,
			User:                createDto[0].User,
			RepositoryUrl:       createDto[0].Source.Repository.Url,
			EnvVars:             createDto[0].EnvVars,
			GitProviderConfigId: createDto[0].GitProviderConfigId,
		}

		res, err = apiClient.WorkspaceTemplateAPI.SaveWorkspaceTemplate(ctx).WorkspaceTemplate(newWorkspaceTemplate).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		views.RenderInfoMessage("Workspace template updated successfully")
		return nil
	},
}
