// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspacetemplate

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	cmd_common "github.com/daytonaio/daytona/pkg/cmd/common"
	create_cmd "github.com/daytonaio/daytona/pkg/cmd/workspace/create"
	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/views"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/daytonaio/daytona/pkg/views/workspace/create"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:     "create",
	Short:   "Create a workspace template",
	Args:    cobra.MaximumNArgs(1),
	Aliases: cmd_common.GetAliases("create"),
	RunE: func(cmd *cobra.Command, args []string) error {
		var workspaceTemplate *apiclient.WorkspaceTemplate
		var workspaceTemplateName *string
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		gitProviders, res, err := apiClient.GitProviderAPI.ListGitProviders(ctx).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		if len(args) == 0 {
			workspaceTemplate, err = RunWorkspaceTemplateAddFlow(apiClient, gitProviders, ctx)
			if err != nil {
				return err
			}
			if workspaceTemplate == nil {
				return nil
			}
			workspaceTemplateName = &workspaceTemplate.Name
		} else {
			workspaceTemplateName, err = processCmdArgument(args[0], apiClient, ctx)
			if err != nil {
				return err
			}
		}

		if workspaceTemplateName == nil {
			return errors.New("workspace template name is required")
		}

		views.RenderInfoMessage(fmt.Sprintf("Workspace template %s added successfully", *workspaceTemplateName))
		return nil
	},
}

func RunWorkspaceTemplateAddFlow(apiClient *apiclient.APIClient, gitProviders []apiclient.GitProvider, ctx context.Context) (*apiclient.WorkspaceTemplate, error) {
	if cmd_common.CheckAnyWorkspaceConfigurationFlagSet(workspaceConfigurationFlags) {
		return nil, errors.New("please provide the repository URL in order to set up custom workspace template details through the CLI")
	}

	var createDtos []apiclient.CreateWorkspaceDTO
	existingWorkspaceTemplateNames, err := getExistingWorkspaceTemplateNames(apiClient)
	if err != nil {
		return nil, err
	}

	apiServerConfig, res, err := apiClient.ServerAPI.GetConfig(context.Background()).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	workspaceDefaults := &views_util.WorkspaceTemplateDefaults{
		BuildChoice:          views_util.AUTOMATIC,
		Image:                &apiServerConfig.DefaultWorkspaceImage,
		ImageUser:            &apiServerConfig.DefaultWorkspaceUser,
		DevcontainerFilePath: create.DEVCONTAINER_FILEPATH,
	}

	createDtos, err = create_cmd.GetWorkspacesCreationDataFromPrompt(ctx, create_cmd.WorkspacesDataPromptParams{
		UserGitProviders:    gitProviders,
		Manual:              *workspaceConfigurationFlags.Manual,
		MultiWorkspace:      false,
		SkipBranchSelection: true,
		ApiClient:           apiClient,
		Defaults:            workspaceDefaults,
	})

	if err != nil {
		if common.IsCtrlCAbort(err) {
			return nil, nil
		} else {
			return nil, err
		}
	}

	create.WorkspacesConfigurationChanged, err = create.RunWorkspaceConfiguration(&createDtos, *workspaceDefaults, false)
	if err != nil {
		return nil, err
	}

	if len(createDtos) == 0 {
		return nil, errors.New("no workspaces found")
	}

	if createDtos[0].Name == "" {
		return nil, errors.New("workspace template name is required")
	}

	initialSuggestion := createDtos[0].Name

	chosenName := create_cmd.GetSuggestedName(initialSuggestion, existingWorkspaceTemplateNames)

	submissionFormConfig := create.SubmissionFormParams{
		ChosenName:             &chosenName,
		SuggestedName:          chosenName,
		ExistingWorkspaceNames: existingWorkspaceTemplateNames,
		WorkspaceList:          &createDtos,
		NameLabel:              "Workspace Template",
		Defaults:               workspaceDefaults,
	}

	err = create.RunSubmissionForm(submissionFormConfig)
	if err != nil {
		return nil, err
	}

	createWorkspaceTemplate := apiclient.CreateWorkspaceTemplateDTO{
		Name:                chosenName,
		BuildConfig:         createDtos[0].BuildConfig,
		Image:               createDtos[0].Image,
		User:                createDtos[0].User,
		RepositoryUrl:       createDtos[0].Source.Repository.Url,
		EnvVars:             createDtos[0].EnvVars,
		GitProviderConfigId: createDtos[0].GitProviderConfigId,
	}

	res, err = apiClient.WorkspaceTemplateAPI.SaveWorkspaceTemplate(ctx).WorkspaceTemplate(createWorkspaceTemplate).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	workspaceTemplate := apiclient.WorkspaceTemplate{
		BuildConfig:         createWorkspaceTemplate.BuildConfig,
		Default:             false,
		EnvVars:             createWorkspaceTemplate.EnvVars,
		Name:                createWorkspaceTemplate.Name,
		Prebuilds:           nil,
		RepositoryUrl:       createWorkspaceTemplate.RepositoryUrl,
		GitProviderConfigId: createWorkspaceTemplate.GitProviderConfigId,
	}

	if createWorkspaceTemplate.Image != nil {
		workspaceTemplate.Image = *createWorkspaceTemplate.Image
	}

	if createWorkspaceTemplate.User != nil {
		workspaceTemplate.User = *createWorkspaceTemplate.User
	}

	if createWorkspaceTemplate.GitProviderConfigId == nil && *createWorkspaceTemplate.GitProviderConfigId == "" {
		gitProviderConfigId, res, err := apiClient.GitProviderAPI.FindGitProviderIdForUrl(ctx, url.QueryEscape(createWorkspaceTemplate.RepositoryUrl)).Execute()
		if err != nil {
			return nil, apiclient_util.HandleErrorResponse(res, err)
		}
		workspaceTemplate.GitProviderConfigId = &gitProviderConfigId
	}

	return &workspaceTemplate, nil
}

func processCmdArgument(argument string, apiClient *apiclient.APIClient, ctx context.Context) (*string, error) {
	if *workspaceConfigurationFlags.Builder != "" && *workspaceConfigurationFlags.Builder != views_util.DEVCONTAINER && *workspaceConfigurationFlags.DevcontainerPath != "" {
		return nil, fmt.Errorf("can't set devcontainer file path if builder is not set to %s", views_util.DEVCONTAINER)
	}

	apiServerConfig, res, err := apiClient.ServerAPI.GetConfig(context.Background()).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	existingWorkspaceTemplateNames, err := getExistingWorkspaceTemplateNames(apiClient)
	if err != nil {
		return nil, err
	}

	repoUrl, err := util.GetValidatedUrl(argument)
	if err != nil {
		return nil, err
	}

	_, res, err = apiClient.GitProviderAPI.GetGitContext(ctx).Repository(apiclient.GetRepositoryContext{
		Url: repoUrl,
	}).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	workspaceConfigurationFlags.GitProviderConfig, err = create_cmd.GetGitProviderConfigIdFromFlag(ctx, apiClient, workspaceConfigurationFlags.GitProviderConfig)
	if err != nil {
		return nil, err
	}

	workspace, err := create_cmd.GetCreateWorkspaceDtoFromFlags(workspaceConfigurationFlags)
	if err != nil {
		return nil, err
	}

	var name string
	if nameFlag != "" {
		name = nameFlag
	} else {
		workspaceName := create_cmd.GetWorkspaceNameFromRepo(repoUrl)
		name = create_cmd.GetSuggestedName(workspaceName, existingWorkspaceTemplateNames)
	}

	if workspace.GitProviderConfigId == nil || *workspace.GitProviderConfigId == "" {
		gitProviderConfigId, res, err := apiClient.GitProviderAPI.FindGitProviderIdForUrl(ctx, url.QueryEscape(repoUrl)).Execute()
		if err != nil {
			return nil, apiclient_util.HandleErrorResponse(res, err)
		}
		*workspace.GitProviderConfigId = gitProviderConfigId
	}

	newWorkspaceTemplate := apiclient.CreateWorkspaceTemplateDTO{
		Name:                name,
		BuildConfig:         workspace.BuildConfig,
		Image:               workspace.Image,
		User:                workspace.User,
		RepositoryUrl:       repoUrl,
		EnvVars:             workspace.EnvVars,
		GitProviderConfigId: workspace.GitProviderConfigId,
	}

	if newWorkspaceTemplate.Image == nil {
		newWorkspaceTemplate.Image = &apiServerConfig.DefaultWorkspaceImage
	}

	if newWorkspaceTemplate.User == nil {
		newWorkspaceTemplate.User = &apiServerConfig.DefaultWorkspaceUser
	}

	res, err = apiClient.WorkspaceTemplateAPI.SaveWorkspaceTemplate(ctx).WorkspaceTemplate(newWorkspaceTemplate).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	return &newWorkspaceTemplate.Name, nil
}

func getExistingWorkspaceTemplateNames(apiClient *apiclient.APIClient) ([]string, error) {
	var existingWorkspaceTemplateNames []string

	existingWorkspaceTemplates, res, err := apiClient.WorkspaceTemplateAPI.ListWorkspaceTemplates(context.Background()).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	for _, wt := range existingWorkspaceTemplates {
		existingWorkspaceTemplateNames = append(existingWorkspaceTemplateNames, wt.Name)
	}

	return existingWorkspaceTemplateNames, nil
}

var nameFlag string

var workspaceConfigurationFlags = cmd_common.WorkspaceConfigurationFlags{
	Builder:           new(views_util.BuildChoice),
	CustomImage:       new(string),
	CustomImageUser:   new(string),
	Branches:          new([]string),
	DevcontainerPath:  new(string),
	EnvVars:           new([]string),
	Manual:            new(bool),
	GitProviderConfig: new(string),
	Labels:            new([]string),
}

func init() {
	createCmd.Flags().StringVar(&nameFlag, "name", "", "Specify the workspace template name")
	cmd_common.AddWorkspaceConfigurationFlags(createCmd, workspaceConfigurationFlags, false)
}
