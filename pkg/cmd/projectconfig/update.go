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
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var projectConfigUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a project config",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		var projectConfig *apiclient.ProjectConfig
		var projects []apiclient.CreateProjectDTO
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		projectConfigList, res, err := apiClient.ProjectConfigAPI.ListProjectConfigs(ctx).Execute()
		if err != nil {
			log.Fatal(apiclient_util.HandleErrorResponse(res, err))
		}

		projectConfig = selection.GetProjectConfigFromPrompt(projectConfigList, 0, false, "Update")

		if projectConfig == nil || projectConfig.Name == nil {
			log.Fatal("project config not found")
		}

		projects = append(projects, apiclient.CreateProjectDTO{
			NewConfig: &apiclient.CreateProjectConfigDTO{
				Name: projectConfig.Name,
				Source: &apiclient.CreateProjectConfigSourceDTO{
					Repository: projectConfig.Repository,
				},
				Build:   projectConfig.Build,
				EnvVars: &map[string]string{},
			},
		})

		profileData, res, err := apiClient.ProfileAPI.GetProfileData(ctx).Execute()
		if err != nil {
			log.Fatal(apiclient_util.HandleErrorResponse(res, err))
		}

		projectDefaults := &create.ProjectDefaults{
			BuildChoice: create.AUTOMATIC,
		}

		if projectConfig.Image != nil {
			projectDefaults.Image = projectConfig.Image
		}

		if projectConfig.User != nil {
			projectDefaults.ImageUser = projectConfig.User
		}

		if projectConfig.Build != nil && projectConfig.Build.Devcontainer != nil && projectConfig.Build.Devcontainer.FilePath != nil {
			projectDefaults.DevcontainerFilePath = *projectConfig.Build.Devcontainer.FilePath
		}

		create.ProjectsConfigurationChanged, err = create.ConfigureProjects(&projects, *projectDefaults)
		if err != nil {
			log.Fatal(err)
		}

		newProjectConfig := apiclient.CreateProjectConfigDTO{
			Name:  projectConfig.Name,
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

		views.RenderInfoMessage("Project config updated successfully")
	},
}
