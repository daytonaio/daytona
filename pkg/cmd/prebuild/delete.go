// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package prebuild

import (
	"github.com/spf13/cobra"
)

var prebuildDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a prebuild configuration",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// var selectedProjectConfig *apiclient.ProjectConfig
		// var selectedProjectConfigName string

		// apiClient, err := apiclient_util.GetApiClient(nil)
		// if err != nil {
		// 	log.Fatal(err)
		// }

		// if len(args) == 0 {
		// 	projectConfigs, res, err := apiClient.ProjectConfigAPI.ListProjectConfigs(context.Background()).Execute()
		// 	if err != nil {
		// 		log.Fatal(apiclient_util.HandleErrorResponse(res, err))
		// 	}

		// 	if len(projectConfigs) == 0 {
		// 		views.RenderInfoMessage("No project configs found")
		// 		return
		// 	}

		// 	selectedProjectConfig = selection.GetProjectConfigFromPrompt(projectConfigs, 0, false, "Delete")
		// 	selectedProjectConfigName = *selectedProjectConfig.Name
		// } else {
		// 	selectedProjectConfigName = args[0]
		// }

		// res, err := apiClient.ProjectConfigAPI.DeleteProjectConfig(context.Background(), selectedProjectConfigName).Execute()
		// if err != nil {
		// 	log.Fatal(apiclient_util.HandleErrorResponse(res, err))
		// }

		// views.RenderInfoMessage("Project config deleted successfully")
	},
}
