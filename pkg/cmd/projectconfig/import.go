// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package projectconfig

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/charmbracelet/huh"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	workspace_util "github.com/daytonaio/daytona/pkg/cmd/workspace/util"
	"github.com/daytonaio/daytona/pkg/views"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/daytonaio/daytona/pkg/views/workspace/create"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
	"github.com/spf13/cobra"
)

var projectConfigImportCmd = &cobra.Command{
	Use:     "import",
	Aliases: []string{"imp"},
	Short:   "Import project config from JSON",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var inputText string
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewText().
					Title("Import Project-Config").
					Description("Enter Project-Config as a JSON or an array of JSON objects").
					CharLimit(-1).
					Value(&inputText),
			),
		).WithTheme(views.GetCustomTheme()).WithHeight(20)
		err = form.Run()
		if err != nil {
			return err
		}

		var config apiclient.ProjectConfig
		err = json.Unmarshal([]byte(inputText), &config)
		if err == nil {
			projectConfigList, res, err := apiClient.ProjectConfigAPI.ListProjectConfigs(ctx).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			if isProjectConfigAlreadyExists(projectConfigList, config) {
				views.RenderInfoMessage(fmt.Sprintf("Project config already present with name \"%s\"", config.Name))
				return nil
			}
			err = importProjectConfig(ctx, apiClient, config)
			if err != nil {
				return err
			}
		} else {
			var configs []apiclient.ProjectConfig
			err = json.Unmarshal([]byte(inputText), &configs)
			if err != nil {
				return fmt.Errorf("invalid JSON input: %v", err)
			}

			for _, config := range configs {
				projectConfigList, res, err := apiClient.ProjectConfigAPI.ListProjectConfigs(ctx).Execute()
				if err != nil {
					return apiclient_util.HandleErrorResponse(res, err)
				}

				if isProjectConfigAlreadyExists(projectConfigList, config) {
					views.RenderInfoMessage(fmt.Sprintf("Project config already present with name \"%s\"", config.Name))
					continue
				}
				err = importProjectConfig(ctx, apiClient, config)
				if err != nil {
					return err
				}
			}
		}

		return nil
	},
}

func isProjectConfigAlreadyExists(projectConfigList []apiclient.ProjectConfig, config apiclient.ProjectConfig) bool {
	for _, projectConfig := range projectConfigList {
		if projectConfig.Name == config.Name {
			return true
		}
	}
	return false
}

func importProjectConfig(ctx context.Context, apiClient *apiclient.APIClient, config apiclient.ProjectConfig) error {
	var verifiedGitProvider bool
	if config.GitProviderConfigId != nil {
		_, _, err := apiClient.GitProviderAPI.GetGitProvider(ctx, *config.GitProviderConfigId).Execute()
		if err == nil {
			verifiedGitProvider = true
		}
	}

	var gitProviders []apiclient.GitProvider

	if !verifiedGitProvider {
		var err error
		gitProviders, _, err = apiClient.GitProviderAPI.ListGitProvidersForUrl(ctx, url.QueryEscape(config.RepositoryUrl)).Execute()
		if err != nil {
			return fmt.Errorf("error fetching Git providers: %v", err)
		}

		if len(gitProviders) == 0 {
			gitProviderConfigId, _, err := apiClient.GitProviderAPI.GetGitProviderIdForUrl(ctx, url.QueryEscape(config.RepositoryUrl)).Execute()
			if err != nil {
				return fmt.Errorf("error fetching Git provider: %v", err)
			}
			config.GitProviderConfigId = &gitProviderConfigId
		}

		if len(gitProviders) == 1 && config.GitProviderConfigId == nil {
			config.GitProviderConfigId = &gitProviders[0].Id
		} else if len(gitProviders) > 1 && config.GitProviderConfigId == nil {
			selectedGitProvider := selection.GetGitProviderConfigFromPrompt(selection.GetGitProviderConfigParams{
				GitProviderConfigs: gitProviders,
				ActionVerb:         "Use",
			})
			config.GitProviderConfigId = &selectedGitProvider.Id
		}
	}

	apiServerConfig, res, err := apiClient.ServerAPI.GetConfig(context.Background()).Execute()
	if err != nil {
		return apiclient_util.HandleErrorResponse(res, err)
	}

	newProjectConfig := apiclient.CreateProjectConfigDTO{
		Name:                config.Name,
		BuildConfig:         config.BuildConfig,
		Image:               &config.Image,
		User:                &config.User,
		RepositoryUrl:       config.RepositoryUrl,
		EnvVars:             config.EnvVars,
		GitProviderConfigId: config.GitProviderConfigId,
	}

	if newProjectConfig.Image == nil {
		newProjectConfig.Image = &apiServerConfig.DefaultProjectImage
	}

	if newProjectConfig.User == nil {
		newProjectConfig.User = &apiServerConfig.DefaultProjectUser
	}

	existingProjectConfigNames, err := GetExistingProjectConfigNames(apiClient)
	if err != nil {
		return err
	}

	apiServerConfig, res, err = apiClient.ServerAPI.GetConfig(context.Background()).Execute()
	if err != nil {
		return apiclient_util.HandleErrorResponse(res, err)
	}

	projectDefaults := &views_util.ProjectConfigDefaults{
		BuildChoice:          views_util.AUTOMATIC,
		Image:                &apiServerConfig.DefaultProjectImage,
		ImageUser:            &apiServerConfig.DefaultProjectUser,
		DevcontainerFilePath: create.DEVCONTAINER_FILEPATH,
	}

	createDto := []apiclient.CreateProjectDTO{
		{
			Name: config.Name,
			Source: apiclient.CreateProjectSourceDTO{
				Repository: apiclient.GitRepository{
					Url: config.RepositoryUrl,
				},
			},
			BuildConfig:         config.BuildConfig,
			EnvVars:             config.EnvVars,
			GitProviderConfigId: config.GitProviderConfigId,
		},
	}

	create.ProjectsConfigurationChanged, err = create.RunProjectConfiguration(&createDto, *projectDefaults, true)
	if err != nil {
		return err
	}

	initialSuggestion := createDto[0].Name

	chosenName := workspace_util.GetSuggestedName(initialSuggestion, existingProjectConfigNames)

	submissionFormConfig := create.SubmissionFormConfig{
		ChosenName:    &config.Name,
		SuggestedName: chosenName,
		ExistingNames: existingProjectConfigNames,
		ProjectList:   &createDto,
		NameLabel:     "Project config",
		Defaults:      projectDefaults,
	}

	confirmation := true
	err = create.RunSubmissionForm(submissionFormConfig, &confirmation)
	if err != nil {
		return err
	}

	if confirmation {
		res, err = apiClient.ProjectConfigAPI.SetProjectConfig(ctx).ProjectConfig(newProjectConfig).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		views.RenderInfoMessage(fmt.Sprintf("Project config %s imported successfully", newProjectConfig.Name))
		return nil
	}

	views.RenderInfoMessage(fmt.Sprintf("Project config %s import cancelled", newProjectConfig.Name))
	return nil
}
