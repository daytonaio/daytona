// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package projectconfig

import (
	"context"
	"fmt"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/target/selection"
	"github.com/spf13/cobra"
)

var projectConfigSetDefaultCmd = &cobra.Command{
	Use:   "set-default",
	Short: "Set project config info",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var projectConfigName string
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		if len(args) == 0 {
			projectConfigList, res, err := apiClient.ProjectConfigAPI.ListProjectConfigs(ctx).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			projectConfig := selection.GetProjectConfigFromPrompt(projectConfigList, 0, false, false, "Make Default")
			if projectConfig == nil {
				return nil
			}
			projectConfigName = projectConfig.Name
		} else {
			projectConfigName = args[0]
		}

		res, err := apiClient.ProjectConfigAPI.SetDefaultProjectConfig(ctx, projectConfigName).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		views.RenderInfoMessage(fmt.Sprintf("Project config '%s' set as default", projectConfigName))
		return nil
	},
}

func init() {
}
