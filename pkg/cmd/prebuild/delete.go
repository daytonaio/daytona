// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package prebuild

import (
	"context"
	"net/http"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
	"github.com/spf13/cobra"
)

var forceFlag bool

var prebuildDeleteCmd = &cobra.Command{
	Use:     "delete [PROJECT_CONFIG] [PREBUILD]",
	Short:   "Delete a prebuild configuration",
	Aliases: []string{"remove", "rm"},
	Args:    cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		var selectedPrebuild *apiclient.PrebuildDTO
		var selectedPrebuildId string
		var selectedProjectConfigName string

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		if len(args) < 2 {
			var prebuilds []apiclient.PrebuildDTO
			var res *http.Response

			if len(args) == 1 {
				selectedProjectConfigName = args[0]
				prebuilds, res, err = apiClient.PrebuildAPI.ListPrebuildsForProjectConfig(context.Background(), selectedProjectConfigName).Execute()
				if err != nil {
					return apiclient_util.HandleErrorResponse(res, err)
				}
			} else {
				prebuilds, res, err = apiClient.PrebuildAPI.ListPrebuilds(context.Background()).Execute()
				if err != nil {
					return apiclient_util.HandleErrorResponse(res, err)
				}
			}

			if len(prebuilds) == 0 {
				views.RenderInfoMessage("No prebuilds found")
				return nil
			}

			selectedPrebuild = selection.GetPrebuildFromPrompt(prebuilds, "Delete")
			if selectedPrebuild == nil {
				return nil
			}
			selectedPrebuildId = selectedPrebuild.Id
			selectedProjectConfigName = selectedPrebuild.ProjectConfigName
		} else {
			selectedProjectConfigName = args[0]
			selectedPrebuildId = args[1]
		}

		res, err := apiClient.PrebuildAPI.DeletePrebuild(context.Background(), selectedProjectConfigName, selectedPrebuildId).Force(forceFlag).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		views.RenderInfoMessage("Prebuild deleted successfully")

		return nil
	},
}

func init() {
	prebuildDeleteCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Force delete prebuild")
}
