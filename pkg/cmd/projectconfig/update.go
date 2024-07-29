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
		var projects []apiclient.CreateProjectConfigDTO
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

		projects = append(projects, apiclient.CreateProjectConfigDTO{
			Name: projectConfig.Name,
			Source: &apiclient.CreateProjectConfigSourceDTO{
				Repository: projectConfig.Repository,
			},
			BuildConfig: projectConfig.BuildConfig,
			EnvVars:     &map[string]string{},
		})

		projectDefaults := &create.ProjectDefaults{
			BuildChoice: create.AUTOMATIC,
		}

		if projectConfig.Image != nil {
			projectDefaults.Image = projectConfig.Image
		}

		if projectConfig.User != nil {
			projectDefaults.ImageUser = projectConfig.User
		}

		if projectConfig.BuildConfig != nil && projectConfig.BuildConfig.Devcontainer != nil && projectConfig.BuildConfig.Devcontainer.FilePath != nil {
			projectDefaults.DevcontainerFilePath = *projectConfig.BuildConfig.Devcontainer.FilePath
		}

		create.ProjectsConfigurationChanged, err = create.ConfigureProjects(&projects, *projectDefaults)
		if err != nil {
			log.Fatal(err)
		}

		newProjectConfig := apiclient.CreateProjectConfigDTO{
			Name:        projectConfig.Name,
			BuildConfig: projects[0].BuildConfig,
			Image:       projects[0].Image,
			User:        projects[0].User,
			Source: &apiclient.CreateProjectConfigSourceDTO{
				Repository: projects[0].Source.Repository,
			},
		}

		newProjectConfig.EnvVars = workspace_util.GetEnvVariables(&projects[0], nil)

		res, err = apiClient.ProjectConfigAPI.SetProjectConfig(ctx).ProjectConfig(newProjectConfig).Execute()
		if err != nil {
			log.Fatal(apiclient_util.HandleErrorResponse(res, err))
		}

		views.RenderInfoMessage("Project config updated successfully")
	},
}
