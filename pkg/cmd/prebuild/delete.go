// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package prebuild

import (
	"context"
	"net/http"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/common"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/selection"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/spf13/cobra"
)

var forceFlag bool

var deleteCmd = &cobra.Command{
	Use:     "delete [WORKSPACE_CONFIG] [PREBUILD]",
	Short:   "Delete a prebuild configuration",
	Args:    cobra.MaximumNArgs(2),
	Aliases: common.GetAliases("delete"),
	RunE: func(cmd *cobra.Command, args []string) error {
		var selectedPrebuild *apiclient.PrebuildDTO
		var selectedPrebuildId string
		var selectedWorkspaceTemplateName string

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		if len(args) < 2 {
			var prebuilds []apiclient.PrebuildDTO
			var res *http.Response

			if len(args) == 1 {
				selectedWorkspaceTemplateName = args[0]
				prebuilds, res, err = apiClient.PrebuildAPI.ListPrebuildsForWorkspaceTemplate(context.Background(), selectedWorkspaceTemplateName).Execute()
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
				views_util.NotifyEmptyPrebuildList(false)
				return nil
			}

			selectedPrebuild = selection.GetPrebuildFromPrompt(prebuilds, "Delete")
			if selectedPrebuild == nil {
				return nil
			}
			selectedPrebuildId = selectedPrebuild.Id
			selectedWorkspaceTemplateName = selectedPrebuild.WorkspaceTemplateName
		} else {
			selectedWorkspaceTemplateName = args[0]
			selectedPrebuildId = args[1]
		}

		res, err := apiClient.PrebuildAPI.DeletePrebuild(context.Background(), selectedWorkspaceTemplateName, selectedPrebuildId).Force(forceFlag).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		views.RenderInfoMessage("Prebuild deleted successfully")

		return nil
	},
}

func init() {
	deleteCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Force delete prebuild")
}
