// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaceconfig

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"

	"github.com/charmbracelet/huh"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/target/create"
	"github.com/daytonaio/daytona/pkg/views/target/selection"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/spf13/cobra"
)

var filePath string

var workspaceTemplateImportCmd = &cobra.Command{
	Use:     "import",
	Aliases: []string{"imp"},
	Short:   "Import workspace template from JSON",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var inputText string
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		workspaceConfigList, res, err := apiClient.WorkspaceConfigAPI.ListWorkspaceConfigs(ctx).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		if filePath != "" {
			if filePath == "-" {
				inputBytes, err := io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("error reading stdin: %v", err)
				}
				inputText = string(inputBytes)
			} else {
				inputBytes, err := os.ReadFile(filePath)
				if err != nil {
					return fmt.Errorf("error reading file: %v", err)
				}
				inputText = string(inputBytes)
			}

		} else {
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewText().
						Title("Import Workspace-Config").
						Description("Enter Workspace-Config as a JSON or an array of JSON objects").
						CharLimit(-1).
						Value(&inputText),
				),
			).WithTheme(views.GetCustomTheme()).WithHeight(20)
			err = form.Run()
			if err != nil {
				return err
			}
		}

		var config apiclient.WorkspaceConfig
		err = json.Unmarshal([]byte(inputText), &config)
		if err == nil {
			err = importWorkspaceConfig(ctx, apiClient, config, &workspaceConfigList)
			if err != nil {
				return fmt.Errorf("error importing workspace config: %v", err)
			}
		} else {
			var configs []apiclient.WorkspaceConfig
			err = json.Unmarshal([]byte(inputText), &configs)
			if err != nil {
				return fmt.Errorf("invalid JSON input: %v", err)
			}

			for _, config := range configs {
				err = importWorkspaceConfig(ctx, apiClient, config, &workspaceConfigList)
				if err != nil {
					return fmt.Errorf("error importing workspace config: %v", err)
				}
			}
		}
		return nil
	},
}

func init() {
	workspaceTemplateImportCmd.Flags().StringVarP(&filePath, "file", "f", "", "Import workspace template from a JSON file. Use '-' to read from stdin.")
}

func isWorkspaceConfigAlreadyExists(configName string, workspaceConfigList *[]apiclient.WorkspaceConfig) bool {
	for _, workspaceConfig := range *workspaceConfigList {
		if workspaceConfig.Name == configName {
			return true
		}
	}
	return false
}

func importWorkspaceConfig(ctx context.Context, apiClient *apiclient.APIClient, config apiclient.WorkspaceConfig, workspaceConfigList *[]apiclient.WorkspaceConfig) error {
	if isWorkspaceConfigAlreadyExists(config.Name, workspaceConfigList) {
		return fmt.Errorf("workspace config already present with name \"%s\"", config.Name)
	}

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

	newWorkspaceConfig := apiclient.CreateWorkspaceConfigDTO{
		Name:                config.Name,
		BuildConfig:         config.BuildConfig,
		Image:               &config.Image,
		User:                &config.User,
		RepositoryUrl:       config.RepositoryUrl,
		EnvVars:             config.EnvVars,
		GitProviderConfigId: config.GitProviderConfigId,
	}

	if newWorkspaceConfig.Image == nil {
		newWorkspaceConfig.Image = &apiServerConfig.DefaultWorkspaceImage
	}

	if newWorkspaceConfig.User == nil {
		newWorkspaceConfig.User = &apiServerConfig.DefaultWorkspaceUser
	}

	existingWorkspaceConfigNames, err := getExistingWorkspaceConfigNames(apiClient)
	if err != nil {
		return err
	}

	apiServerConfig, res, err = apiClient.ServerAPI.GetConfig(context.Background()).Execute()
	if err != nil {
		return apiclient_util.HandleErrorResponse(res, err)
	}

	workspaceDefaults := &views_util.WorkspaceConfigDefaults{
		BuildChoice:          views_util.AUTOMATIC,
		Image:                &apiServerConfig.DefaultWorkspaceImage,
		ImageUser:            &apiServerConfig.DefaultWorkspaceUser,
		DevcontainerFilePath: create.DEVCONTAINER_FILEPATH,
	}

	createDto := []apiclient.CreateWorkspaceDTO{
		{
			Name: config.Name,
			Source: apiclient.CreateWorkspaceSourceDTO{
				Repository: apiclient.GitRepository{
					Url: config.RepositoryUrl,
				},
			},
			BuildConfig:         config.BuildConfig,
			EnvVars:             config.EnvVars,
			GitProviderConfigId: config.GitProviderConfigId,
		},
	}

	create.WorkspacesConfigurationChanged, err = create.RunWorkspaceConfiguration(&createDto, *workspaceDefaults, true)
	if err != nil {
		return err
	}

	submissionFormConfig := create.SubmissionFormConfig{
		ChosenName:    &config.Name,
		SuggestedName: config.Name,
		ExistingNames: existingWorkspaceConfigNames,
		WorkspaceList: &createDto,
		NameLabel:     "Workspace config",
		Defaults:      workspaceDefaults,
	}

	confirmation := true
	err = create.RunSubmissionForm(submissionFormConfig, &confirmation)
	if err != nil {
		return err
	}

	if confirmation {
		res, err = apiClient.WorkspaceConfigAPI.SetWorkspaceConfig(ctx).WorkspaceConfig(newWorkspaceConfig).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		*workspaceConfigList = append(*workspaceConfigList, config)
		views.RenderInfoMessage(fmt.Sprintf("Workspace config %s imported successfully", newWorkspaceConfig.Name))
		return nil
	}

	views.RenderInfoMessage(fmt.Sprintf("Workspace config %s import cancelled", newWorkspaceConfig.Name))
	return nil
}
