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
	"github.com/daytonaio/daytona/pkg/views"
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
			fmt.Printf("Single Project Config: %+v\n", config)

			// if config.GitProviderConfigId != nil {
			// 	// Verify if the Git provider config exists
			// 	_, _, err := apiClient.GitProviderAPI.GetGitProvider(ctx, *config.GitProviderConfigId).Execute()
			// 	if err == nil {
			// 		// Git provider config exists, proceed with import
			// 		return nil
			// 	}
			// }

			// GitProviderConfigId is not set or does not exist, fetch it using the repo URL
			gitProviders, _, err := apiClient.GitProviderAPI.ListGitProvidersForUrl(ctx, url.QueryEscape(config.RepositoryUrl)).Execute()
			if err != nil {
				return fmt.Errorf("error fetching Git providers: %v", err)
			}

			if len(gitProviders) == 0 {
				return fmt.Errorf("no Git providers found for URL: %s", config.RepositoryUrl)
			}

			if len(gitProviders) == 1 {
				config.GitProviderConfigId = &gitProviders[0].Id
			} else {
				// Multiple Git providers available, prompt the user to select one
				selectedGitProvider := selection.GetGitProviderConfigFromPrompt(selection.GetGitProviderConfigParams{
					GitProviderConfigs: gitProviders,
					ActionVerb:         "Use",
				})
				config.GitProviderConfigId = &selectedGitProvider.Id
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

			res, err = apiClient.ProjectConfigAPI.SetProjectConfig(ctx).ProjectConfig(newProjectConfig).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}
		} else {
			var configs []apiclient.ProjectConfig
			err = json.Unmarshal([]byte(inputText), &configs)
			if err != nil {
				return fmt.Errorf("invalid JSON input: %v", err)
			}

			fmt.Println("Multiple Project Configs:")
			for i, config := range configs {
				fmt.Printf("Project Config %d: %+v\n", i+1, config)
			}
		}

		return nil
	},
}

func init() {
	ProjectConfigCmd.AddCommand(projectConfigImportCmd)
}
