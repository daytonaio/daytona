// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package projectconfig

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	target_util "github.com/daytonaio/daytona/pkg/cmd/target/util"
	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/target/create"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/spf13/cobra"
)

var projectConfigAddCmd = &cobra.Command{
	Use:     "add",
	Aliases: []string{"new", "create"},
	Short:   "Add a project config",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var projectConfig *apiclient.ProjectConfig
		var projectConfigName *string
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
			projectConfig, err = RunProjectConfigAddFlow(apiClient, gitProviders, ctx)
			if err != nil {
				return err
			}
			if projectConfig == nil {
				return nil
			}
			projectConfigName = &projectConfig.Name
		} else {
			projectConfigName, err = processCmdArgument(args[0], apiClient, ctx)
			if err != nil {
				return err
			}
		}

		if projectConfigName == nil {
			return errors.New("project config name is required")
		}

		views.RenderInfoMessage(fmt.Sprintf("Project config %s added successfully", *projectConfigName))
		return nil
	},
}

func RunProjectConfigAddFlow(apiClient *apiclient.APIClient, gitProviders []apiclient.GitProvider, ctx context.Context) (*apiclient.ProjectConfig, error) {
	if target_util.CheckAnyProjectConfigurationFlagSet(projectConfigurationFlags) {
		return nil, errors.New("please provide the repository URL in order to set up custom project config details through the CLI")
	}

	var createDtos []apiclient.CreateProjectDTO
	existingProjectConfigNames, err := GetExistingProjectConfigNames(apiClient)
	if err != nil {
		return nil, err
	}

	apiServerConfig, res, err := apiClient.ServerAPI.GetConfig(context.Background()).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	projectDefaults := &views_util.ProjectConfigDefaults{
		BuildChoice:          views_util.AUTOMATIC,
		Image:                &apiServerConfig.DefaultProjectImage,
		ImageUser:            &apiServerConfig.DefaultProjectUser,
		DevcontainerFilePath: create.DEVCONTAINER_FILEPATH,
	}

	createDtos, err = target_util.GetProjectsCreationDataFromPrompt(target_util.ProjectsDataPromptConfig{
		UserGitProviders:    gitProviders,
		Manual:              *projectConfigurationFlags.Manual,
		MultiProject:        false,
		SkipBranchSelection: true,
		ApiClient:           apiClient,
		Defaults:            projectDefaults,
	})

	if err != nil {
		if common.IsCtrlCAbort(err) {
			return nil, nil
		} else {
			return nil, err
		}
	}

	create.ProjectsConfigurationChanged, err = create.RunProjectConfiguration(&createDtos, *projectDefaults, false)
	if err != nil {
		return nil, err
	}

	if len(createDtos) == 0 {
		return nil, errors.New("no projects found")
	}

	if createDtos[0].Name == "" {
		return nil, errors.New("project config name is required")
	}

	initialSuggestion := createDtos[0].Name

	chosenName := target_util.GetSuggestedName(initialSuggestion, existingProjectConfigNames)

	submissionFormConfig := create.SubmissionFormConfig{
		ChosenName:    &chosenName,
		SuggestedName: chosenName,
		ExistingNames: existingProjectConfigNames,
		ProjectList:   &createDtos,
		NameLabel:     "Project config",
		Defaults:      projectDefaults,
	}

	pcFlag := false
	err = create.RunSubmissionForm(submissionFormConfig, &pcFlag)
	if err != nil {
		return nil, err
	}

	createProjectConfig := apiclient.CreateProjectConfigDTO{
		Name:                chosenName,
		BuildConfig:         createDtos[0].BuildConfig,
		Image:               createDtos[0].Image,
		User:                createDtos[0].User,
		RepositoryUrl:       createDtos[0].Source.Repository.Url,
		EnvVars:             createDtos[0].EnvVars,
		GitProviderConfigId: createDtos[0].GitProviderConfigId,
	}

	res, err = apiClient.ProjectConfigAPI.SetProjectConfig(ctx).ProjectConfig(createProjectConfig).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	projectConfig := apiclient.ProjectConfig{
		BuildConfig:         createProjectConfig.BuildConfig,
		Default:             false,
		EnvVars:             createProjectConfig.EnvVars,
		Name:                createProjectConfig.Name,
		Prebuilds:           nil,
		RepositoryUrl:       createProjectConfig.RepositoryUrl,
		GitProviderConfigId: createProjectConfig.GitProviderConfigId,
	}

	if createProjectConfig.Image != nil {
		projectConfig.Image = *createProjectConfig.Image
	}

	if createProjectConfig.User != nil {
		projectConfig.User = *createProjectConfig.User
	}

	if createProjectConfig.GitProviderConfigId == nil && *createProjectConfig.GitProviderConfigId == "" {
		gitProviderConfigId, res, err := apiClient.GitProviderAPI.GetGitProviderIdForUrl(ctx, url.QueryEscape(createProjectConfig.RepositoryUrl)).Execute()
		if err != nil {
			return nil, apiclient_util.HandleErrorResponse(res, err)
		}
		projectConfig.GitProviderConfigId = &gitProviderConfigId
	}

	return &projectConfig, nil
}

func processCmdArgument(argument string, apiClient *apiclient.APIClient, ctx context.Context) (*string, error) {
	if *projectConfigurationFlags.Builder != "" && *projectConfigurationFlags.Builder != views_util.DEVCONTAINER && *projectConfigurationFlags.DevcontainerPath != "" {
		return nil, fmt.Errorf("can't set devcontainer file path if builder is not set to %s", views_util.DEVCONTAINER)
	}

	apiServerConfig, res, err := apiClient.ServerAPI.GetConfig(context.Background()).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	existingProjectConfigNames, err := GetExistingProjectConfigNames(apiClient)
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

	projectConfigurationFlags.GitProviderConfig, err = target_util.GetGitProviderConfigIdFromFlag(ctx, apiClient, projectConfigurationFlags.GitProviderConfig)
	if err != nil {
		return nil, err
	}

	project, err := target_util.GetCreateProjectDtoFromFlags(projectConfigurationFlags)
	if err != nil {
		return nil, err
	}

	var name string
	if nameFlag != "" {
		name = nameFlag
	} else {
		projectName := target_util.GetProjectNameFromRepo(repoUrl)
		name = target_util.GetSuggestedName(projectName, existingProjectConfigNames)
	}

	if project.GitProviderConfigId == nil || *project.GitProviderConfigId == "" {
		gitProviderConfigId, res, err := apiClient.GitProviderAPI.GetGitProviderIdForUrl(ctx, url.QueryEscape(repoUrl)).Execute()
		if err != nil {
			return nil, apiclient_util.HandleErrorResponse(res, err)
		}
		*project.GitProviderConfigId = gitProviderConfigId
	}

	newProjectConfig := apiclient.CreateProjectConfigDTO{
		Name:                name,
		BuildConfig:         project.BuildConfig,
		Image:               project.Image,
		User:                project.User,
		RepositoryUrl:       repoUrl,
		EnvVars:             project.EnvVars,
		GitProviderConfigId: project.GitProviderConfigId,
	}

	if newProjectConfig.Image == nil {
		newProjectConfig.Image = &apiServerConfig.DefaultProjectImage
	}

	if newProjectConfig.User == nil {
		newProjectConfig.User = &apiServerConfig.DefaultProjectUser
	}

	res, err = apiClient.ProjectConfigAPI.SetProjectConfig(ctx).ProjectConfig(newProjectConfig).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	return &newProjectConfig.Name, nil
}

func GetExistingProjectConfigNames(apiClient *apiclient.APIClient) ([]string, error) {
	var existingProjectConfigNames []string

	existingProjectConfigs, res, err := apiClient.ProjectConfigAPI.ListProjectConfigs(context.Background()).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	for _, pc := range existingProjectConfigs {
		existingProjectConfigNames = append(existingProjectConfigNames, pc.Name)
	}

	return existingProjectConfigNames, nil
}

var nameFlag string

var projectConfigurationFlags = target_util.ProjectConfigurationFlags{
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
	projectConfigAddCmd.Flags().StringVar(&nameFlag, "name", "", "Specify the project config name")
	target_util.AddProjectConfigurationFlags(projectConfigAddCmd, projectConfigurationFlags, false)
}
