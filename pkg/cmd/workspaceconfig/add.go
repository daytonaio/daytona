// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaceconfig

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

var workspaceAddCmd = &cobra.Command{
	Use:     "add",
	Aliases: []string{"new", "create"},
	Short:   "Add a workspace config",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var workspaceConfig *apiclient.WorkspaceConfig
		var workspaceConfigName *string
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
			workspaceConfig, err = RunWorkspaceConfigAddFlow(apiClient, gitProviders, ctx)
			if err != nil {
				return err
			}
			if workspaceConfig == nil {
				return nil
			}
			workspaceConfigName = &workspaceConfig.Name
		} else {
			workspaceConfigName, err = processCmdArgument(args[0], apiClient, ctx)
			if err != nil {
				return err
			}
		}

		if workspaceConfigName == nil {
			return errors.New("workspace config name is required")
		}

		views.RenderInfoMessage(fmt.Sprintf("Workspace config %s added successfully", *workspaceConfigName))
		return nil
	},
}

func RunWorkspaceConfigAddFlow(apiClient *apiclient.APIClient, gitProviders []apiclient.GitProvider, ctx context.Context) (*apiclient.WorkspaceConfig, error) {
	if cmd_common.CheckAnyWorkspaceConfigurationFlagSet(workspaceConfigurationFlags) {
		return nil, errors.New("please provide the repository URL in order to set up custom workspace config details through the CLI")
	}

	var createDtos []apiclient.CreateWorkspaceDTO
	existingWorkspaceConfigNames, err := getExistingWorkspaceConfigNames(apiClient)
	if err != nil {
		return nil, err
	}

	apiServerConfig, res, err := apiClient.ServerAPI.GetConfig(context.Background()).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	workspaceDefaults := &views_util.WorkspaceConfigDefaults{
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

	create.WorkspacesConfigurationChanged, err = create.RunWorkspaceConfiguration(&createDtos, *workspaceDefaults)
	if err != nil {
		return nil, err
	}

	if len(createDtos) == 0 {
		return nil, errors.New("no workspaces found")
	}

	if createDtos[0].Name == "" {
		return nil, errors.New("workspace config name is required")
	}

	initialSuggestion := createDtos[0].Name

	chosenName := create_cmd.GetSuggestedName(initialSuggestion, existingWorkspaceConfigNames)

	submissionFormConfig := create.SubmissionFormParams{
		ChosenName:             &chosenName,
		SuggestedName:          chosenName,
		ExistingWorkspaceNames: existingWorkspaceConfigNames,
		WorkspaceList:          &createDtos,
		NameLabel:              "Workspace config",
		Defaults:               workspaceDefaults,
	}

	err = create.RunSubmissionForm(submissionFormConfig)
	if err != nil {
		return nil, err
	}

	createWorkspaceConfig := apiclient.CreateWorkspaceConfigDTO{
		Name:                chosenName,
		BuildConfig:         createDtos[0].BuildConfig,
		Image:               createDtos[0].Image,
		User:                createDtos[0].User,
		RepositoryUrl:       createDtos[0].Source.Repository.Url,
		EnvVars:             createDtos[0].EnvVars,
		GitProviderConfigId: createDtos[0].GitProviderConfigId,
	}

	res, err = apiClient.WorkspaceConfigAPI.SetWorkspaceConfig(ctx).WorkspaceConfig(createWorkspaceConfig).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	workspaceConfig := apiclient.WorkspaceConfig{
		BuildConfig:         createWorkspaceConfig.BuildConfig,
		Default:             false,
		EnvVars:             createWorkspaceConfig.EnvVars,
		Name:                createWorkspaceConfig.Name,
		Prebuilds:           nil,
		RepositoryUrl:       createWorkspaceConfig.RepositoryUrl,
		GitProviderConfigId: createWorkspaceConfig.GitProviderConfigId,
	}

	if createWorkspaceConfig.Image != nil {
		workspaceConfig.Image = *createWorkspaceConfig.Image
	}

	if createWorkspaceConfig.User != nil {
		workspaceConfig.User = *createWorkspaceConfig.User
	}

	if createWorkspaceConfig.GitProviderConfigId == nil && *createWorkspaceConfig.GitProviderConfigId == "" {
		gitProviderConfigId, res, err := apiClient.GitProviderAPI.GetGitProviderIdForUrl(ctx, url.QueryEscape(createWorkspaceConfig.RepositoryUrl)).Execute()
		if err != nil {
			return nil, apiclient_util.HandleErrorResponse(res, err)
		}
		workspaceConfig.GitProviderConfigId = &gitProviderConfigId
	}

	return &workspaceConfig, nil
}

func processCmdArgument(argument string, apiClient *apiclient.APIClient, ctx context.Context) (*string, error) {
	if *workspaceConfigurationFlags.Builder != "" && *workspaceConfigurationFlags.Builder != views_util.DEVCONTAINER && *workspaceConfigurationFlags.DevcontainerPath != "" {
		return nil, fmt.Errorf("can't set devcontainer file path if builder is not set to %s", views_util.DEVCONTAINER)
	}

	apiServerConfig, res, err := apiClient.ServerAPI.GetConfig(context.Background()).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	existingWorkspaceConfigNames, err := getExistingWorkspaceConfigNames(apiClient)
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
		name = create_cmd.GetSuggestedName(workspaceName, existingWorkspaceConfigNames)
	}

	if workspace.GitProviderConfigId == nil || *workspace.GitProviderConfigId == "" {
		gitProviderConfigId, res, err := apiClient.GitProviderAPI.GetGitProviderIdForUrl(ctx, url.QueryEscape(repoUrl)).Execute()
		if err != nil {
			return nil, apiclient_util.HandleErrorResponse(res, err)
		}
		*workspace.GitProviderConfigId = gitProviderConfigId
	}

	newWorkspaceConfig := apiclient.CreateWorkspaceConfigDTO{
		Name:                name,
		BuildConfig:         workspace.BuildConfig,
		Image:               workspace.Image,
		User:                workspace.User,
		RepositoryUrl:       repoUrl,
		EnvVars:             workspace.EnvVars,
		GitProviderConfigId: workspace.GitProviderConfigId,
	}

	if newWorkspaceConfig.Image == nil {
		newWorkspaceConfig.Image = &apiServerConfig.DefaultWorkspaceImage
	}

	if newWorkspaceConfig.User == nil {
		newWorkspaceConfig.User = &apiServerConfig.DefaultWorkspaceUser
	}

	res, err = apiClient.WorkspaceConfigAPI.SetWorkspaceConfig(ctx).WorkspaceConfig(newWorkspaceConfig).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	return &newWorkspaceConfig.Name, nil
}

func getExistingWorkspaceConfigNames(apiClient *apiclient.APIClient) ([]string, error) {
	var existingWorkspaceConfigNames []string

	existingWorkspaceConfigs, res, err := apiClient.WorkspaceConfigAPI.ListWorkspaceConfigs(context.Background()).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	for _, wc := range existingWorkspaceConfigs {
		existingWorkspaceConfigNames = append(existingWorkspaceConfigNames, wc.Name)
	}

	return existingWorkspaceConfigNames, nil
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
}

func init() {
	workspaceAddCmd.Flags().StringVar(&nameFlag, "name", "", "Specify the workspace config name")
	cmd_common.AddWorkspaceConfigurationFlags(workspaceAddCmd, workspaceConfigurationFlags, false)
}
