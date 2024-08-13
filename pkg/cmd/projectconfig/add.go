// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package projectconfig

import (
	"context"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	workspace_util "github.com/daytonaio/daytona/pkg/cmd/workspace/util"
	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/workspace/create"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var projectConfigAddCmd = &cobra.Command{
	Use:     "add",
	Aliases: []string{"new"},
	Short:   "Add a project config",
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		var projects []apiclient.CreateProjectConfigDTO
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

		existingProjectConfigs, res, err := apiClient.ProjectConfigAPI.ListProjectConfigs(context.Background()).Execute()
		if err != nil {
			log.Fatal(apiclient_util.HandleErrorResponse(res, err))
		}
		for _, pc := range existingProjectConfigs {
			existingProjectConfigNames = append(existingProjectConfigNames, pc.Name)
		}

		projectDefaults := &create.ProjectConfigDefaults{
			BuildChoice:          create.AUTOMATIC,
			Image:                &apiServerConfig.DefaultProjectImage,
			ImageUser:            &apiServerConfig.DefaultProjectUser,
			DevcontainerFilePath: create.DEVCONTAINER_FILEPATH,
		}

		projects, err = workspace_util.GetProjectsCreationDataFromPrompt(workspace_util.ProjectsDataPromptConfig{
			UserGitProviders:    gitProviders,
			Manual:              manualFlag,
			MultiProject:        false,
			SkipBranchSelection: true,
			ApiClient:           apiClient,
			Defaults:            projectDefaults,
		},
		)
		if err != nil {
			if common.IsCtrlCAbort(err) {
				return
			} else {
				log.Fatal(err)
			}
		}

		create.ProjectsConfigurationChanged, err = create.RunProjectConfiguration(&projects, *projectDefaults)
		if err != nil {
			log.Fatal(err)
		}

		if len(projects) == 0 {
			log.Fatal("no projects found")
		}

		if projects[0].Name == "" {
			log.Fatal("project config name is required")
		}

		initialSuggestion := projects[0].Name

		chosenName := workspace_util.GetSuggestedName(initialSuggestion, existingProjectConfigNames)

		submissionFormConfig := create.SubmissionFormConfig{
			ChosenName:    &chosenName,
			SuggestedName: chosenName,
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
			Name:        chosenName,
			BuildConfig: projects[0].BuildConfig,
			Image:       projects[0].Image,
			User:        projects[0].User,
			Source: apiclient.CreateProjectConfigSourceDTO{
				Repository: projects[0].Source.Repository,
			},
		}

		newProjectConfig.EnvVars = *workspace_util.GetEnvVariables(&projects[0], nil)

		res, err = apiClient.ProjectConfigAPI.SetProjectConfig(ctx).ProjectConfig(newProjectConfig).Execute()
		if err != nil {
			log.Fatal(apiclient_util.HandleErrorResponse(res, err))
		}

		views.RenderInfoMessage("Project config added successfully")
	},
}

var manualFlag bool

func init() {
	projectConfigAddCmd.Flags().BoolVar(&manualFlag, "manual", false, "Manually enter the Git repository")
}
