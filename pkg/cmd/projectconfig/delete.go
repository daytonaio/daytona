// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package projectconfig

import (
	"context"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var projectConfigDeleteCmd = &cobra.Command{
	Use:     "delete",
	Aliases: []string{"remove", "rm"},
	Short:   "Delete a project config",
	Args:    cobra.RangeArgs(0, 2),
	Run: func(cmd *cobra.Command, args []string) {
		var selectedProjectConfig *apiclient.ProjectConfig
		var selectedProjectConfigName string

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		if len(args) == 0 {
			projectConfigs, res, err := apiClient.ProjectConfigAPI.ListProjectConfigs(context.Background()).Execute()
			if err != nil {
				log.Fatal(apiclient_util.HandleErrorResponse(res, err))
			}

			if len(projectConfigs) == 0 {
				views.RenderInfoMessage("No project configs found")
				return
			}

			selectedProjectConfig = selection.GetProjectConfigFromPrompt(projectConfigs, 0, false, "Delete")
			selectedProjectConfigName = *selectedProjectConfig.Name
		} else {
			selectedProjectConfigName = args[0]
		}

		res, err := apiClient.ProjectConfigAPI.DeleteProjectConfig(context.Background(), selectedProjectConfigName).Execute()
		if err != nil {
			log.Fatal(apiclient_util.HandleErrorResponse(res, err))
		}

		views.RenderInfoMessage("Project config deleted successfully")
	},
}
