// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspacetemplate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/selection"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/daytonaio/daytona/pkg/views/workspace/create"
	"github.com/spf13/cobra"
)

var filePath string

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import a workspace template from a JSON object",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var inputText string
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		workspaceTemplateList, res, err := apiClient.WorkspaceTemplateAPI.ListWorkspaceTemplates(ctx).Execute()
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
						Title("Import Workspace Template").
						Description("Enter Workspace Template as a JSON object or an array of JSON objects").
						CharLimit(-1).
						Value(&inputText),
				),
			).WithTheme(views.GetCustomTheme()).WithHeight(20)
			err = form.Run()
			if err != nil {
				return err
			}
		}

		var template apiclient.WorkspaceTemplate
		err = json.Unmarshal([]byte(inputText), &template)
		if err == nil {
			err = importWorkspaceTemplate(ctx, apiClient, template, &workspaceTemplateList)
			if err != nil {
				return fmt.Errorf("error importing workspace template: %v", err)
			}
		} else {
			var templates []apiclient.WorkspaceTemplate
			err = json.Unmarshal([]byte(inputText), &templates)
			if err != nil {
				return fmt.Errorf("invalid JSON input: %v", err)
			}

			for _, t := range templates {
				err = importWorkspaceTemplate(ctx, apiClient, t, &workspaceTemplateList)
				if err != nil {
					return fmt.Errorf("error importing workspace template: %v", err)
				}
			}
		}
		return nil
	},
}

func init() {
	importCmd.Flags().StringVarP(&filePath, "file", "f", "", "Import workspace template from a JSON file. Use '-' to read from stdin.")
}

func checkWorkspaceTemplateAlreadyExists(templateName string, workspaceTemplateList *[]apiclient.WorkspaceTemplate) bool {
	for _, t := range *workspaceTemplateList {
		if t.Name == templateName {
			return true
		}
	}
	return false
}

func importWorkspaceTemplate(ctx context.Context, apiClient *apiclient.APIClient, template apiclient.WorkspaceTemplate, workspaceTemplateList *[]apiclient.WorkspaceTemplate) error {
	if checkWorkspaceTemplateAlreadyExists(template.Name, workspaceTemplateList) {
		return fmt.Errorf("workspace template already present with name \"%s\"", template.Name)
	}

	var verifiedGitProvider bool
	if template.GitProviderConfigId != nil {
		_, _, err := apiClient.GitProviderAPI.FindGitProvider(ctx, *template.GitProviderConfigId).Execute()
		if err == nil {
			verifiedGitProvider = true
		}
	}

	var gitProviders []apiclient.GitProvider

	if !verifiedGitProvider {
		var err error
		gitProviders, _, err = apiClient.GitProviderAPI.ListGitProvidersForUrl(ctx, url.QueryEscape(template.RepositoryUrl)).Execute()
		if err != nil {
			return fmt.Errorf("error fetching Git providers: %v", err)
		}

		if len(gitProviders) == 0 {
			gitProviderConfigId, _, err := apiClient.GitProviderAPI.FindGitProviderIdForUrl(ctx, url.QueryEscape(template.RepositoryUrl)).Execute()
			if err != nil {
				return fmt.Errorf("error fetching Git provider: %v", err)
			}
			template.GitProviderConfigId = &gitProviderConfigId
		}

		if len(gitProviders) == 1 && template.GitProviderConfigId == nil {
			template.GitProviderConfigId = &gitProviders[0].Id
		} else if len(gitProviders) > 1 && template.GitProviderConfigId == nil {
			selectedGitProvider := selection.GetGitProviderConfigFromPrompt(selection.GetGitProviderConfigParams{
				GitProviderConfigs: gitProviders,
				ActionVerb:         "Use",
			})
			template.GitProviderConfigId = &selectedGitProvider.Id
		}
	}

	apiServerConfig, res, err := apiClient.ServerAPI.GetConfig(context.Background()).Execute()
	if err != nil {
		return apiclient_util.HandleErrorResponse(res, err)
	}

	newWorkspaceTemplate := apiclient.CreateWorkspaceTemplateDTO{
		Name:                template.Name,
		BuildConfig:         template.BuildConfig,
		Image:               &template.Image,
		User:                &template.User,
		RepositoryUrl:       template.RepositoryUrl,
		EnvVars:             template.EnvVars,
		GitProviderConfigId: template.GitProviderConfigId,
	}

	if newWorkspaceTemplate.Image == nil {
		newWorkspaceTemplate.Image = &apiServerConfig.DefaultWorkspaceImage
	}

	if newWorkspaceTemplate.User == nil {
		newWorkspaceTemplate.User = &apiServerConfig.DefaultWorkspaceUser
	}

	existingWorkspaceTemplateNames, err := getExistingWorkspaceTemplateNames(apiClient)
	if err != nil {
		return err
	}

	apiServerConfig, res, err = apiClient.ServerAPI.GetConfig(context.Background()).Execute()
	if err != nil {
		return apiclient_util.HandleErrorResponse(res, err)
	}

	workspaceDefaults := &views_util.WorkspaceTemplateDefaults{
		BuildChoice:          views_util.AUTOMATIC,
		Image:                &apiServerConfig.DefaultWorkspaceImage,
		ImageUser:            &apiServerConfig.DefaultWorkspaceUser,
		DevcontainerFilePath: create.DEVCONTAINER_FILEPATH,
	}

	createDto := []apiclient.CreateWorkspaceDTO{
		{
			Name: template.Name,
			Source: apiclient.CreateWorkspaceSourceDTO{
				Repository: apiclient.GitRepository{
					Url: template.RepositoryUrl,
				},
			},
			BuildConfig:         template.BuildConfig,
			EnvVars:             template.EnvVars,
			GitProviderConfigId: template.GitProviderConfigId,
		},
	}

	create.WorkspacesConfigurationChanged, err = create.RunWorkspaceConfiguration(&createDto, *workspaceDefaults, true)
	if err != nil {
		return err
	}

	submissionFormConfig := create.SubmissionFormParams{
		ChosenName:             &template.Name,
		SuggestedName:          template.Name,
		ExistingWorkspaceNames: existingWorkspaceTemplateNames,
		WorkspaceList:          &createDto,
		NameLabel:              "Workspace template",
		Defaults:               workspaceDefaults,
		ImportConfirmation:     util.Pointer(true),
	}

	err = create.RunSubmissionForm(submissionFormConfig)
	if err != nil {
		return err
	}

	if submissionFormConfig.ImportConfirmation != nil && *submissionFormConfig.ImportConfirmation {
		res, err = apiClient.WorkspaceTemplateAPI.SaveWorkspaceTemplate(ctx).WorkspaceTemplate(newWorkspaceTemplate).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		*workspaceTemplateList = append(*workspaceTemplateList, template)
		views.RenderInfoMessage(fmt.Sprintf("Workspace template %s imported successfully", newWorkspaceTemplate.Name))
		return nil
	}

	views.RenderInfoMessage(fmt.Sprintf("Workspace template %s import cancelled", newWorkspaceTemplate.Name))
	return nil
}
