// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package projectconfig

import (
	"context"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	workspace_util "github.com/daytonaio/daytona/pkg/cmd/workspace/util"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/workspace/create"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var projectConfigAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a project config",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		var projects []apiclient.CreateProjectDTO
		var existingProjectConfigNames []string
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		gitProviders, res, err := apiClient.GitProviderAPI.ListGitProviders(ctx).Execute()
		if err != nil {
			log.Fatal(apiclient_util.HandleErrorResponse(res, err))
		}

		apiServerConfig, res, err := apiClient.ServerAPI.GetConfig(context.Background()).Execute()
		if err != nil {
			log.Fatal(apiclient_util.HandleErrorResponse(res, err))
		}

		profileData, res, err := apiClient.ProfileAPI.GetProfileData(ctx).Execute()
		if err != nil {
			log.Fatal(apiclient_util.HandleErrorResponse(res, err))
		}

		existingProjectConfigs, res, err := apiClient.ProjectConfigAPI.ListProjectConfigs(context.Background()).Execute()
		if err != nil {
			log.Fatal(apiclient_util.HandleErrorResponse(res, err))
		}
		for _, pc := range existingProjectConfigs {
			existingProjectConfigNames = append(existingProjectConfigNames, *pc.Name)
		}

		projectDefaults := &create.ProjectDefaults{
			BuildChoice:          create.AUTOMATIC,
			Image:                apiServerConfig.DefaultProjectImage,
			ImageUser:            apiServerConfig.DefaultProjectUser,
			DevcontainerFilePath: create.DEVCONTAINER_FILEPATH,
		}

		projects, err = workspace_util.GetProjectsCreationDataFromPrompt(workspace_util.ProjectsDataPromptConfig{
			UserGitProviders:    gitProviders,
			Manual:              false,
			MultiProject:        false,
			SkipBranchSelection: true,
			ApiClient:           apiClient,
			Defaults:            projectDefaults,
		},
		)
		if err != nil {
			log.Fatal(err)
		}

		create.ProjectsConfigurationChanged, err = create.ConfigureProjects(&projects, *projectDefaults)
		if err != nil {
			log.Fatal(err)
		}

		if len(projects) == 0 {
			log.Fatal("no projects found")
		}

		if projects[0].NewConfig == nil {
			log.Fatal("project config is required")
		}

		if projects[0].NewConfig.Name == nil {
			log.Fatal("project config name is required")
		}

		initialSuggestion := *projects[0].NewConfig.Name

		suggestedName := workspace_util.GetSuggestedName(initialSuggestion, existingProjectConfigNames)

		chosenName := suggestedName

		submissionFormConfig := create.SubmissionFormConfig{
			ChosenName:    &chosenName,
			SuggestedName: suggestedName,
			ExistingNames: existingProjectConfigNames,
			ProjectList:   &projects,
			NameLabel:     "Project config",
			Defaults:      projectDefaults,
		}

		err = create.RunSubmissionForm(submissionFormConfig)
		if err != nil {
			log.Fatal(err)
		}

		newProjectConfig := apiclient.CreateProjectConfigDTO{
			Name:  &chosenName,
			Build: projects[0].NewConfig.Build,
			Image: projects[0].NewConfig.Image,
			User:  projects[0].NewConfig.User,
			Source: &apiclient.CreateProjectConfigSourceDTO{
				Repository: projects[0].NewConfig.Source.Repository,
			},
		}

		newProjectConfig.EnvVars = workspace_util.GetEnvVariables(&projects[0], profileData)

		res, err = apiClient.ProjectConfigAPI.SetProjectConfig(ctx).ProjectConfig(newProjectConfig).Execute()
		if err != nil {
			log.Fatal(apiclient_util.HandleErrorResponse(res, err))
		}

		views.RenderInfoMessage("Project config added successfully")
	},
}
