// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaceconfig

import (
	"context"
	"fmt"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/target/selection"
	"github.com/spf13/cobra"
)

var workspaceConfigSetDefaultCmd = &cobra.Command{
	Use:   "set-default",
	Short: "Set workspace config info",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var workspaceConfigName string
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		if len(args) == 0 {
			workspaceConfigList, res, err := apiClient.WorkspaceConfigAPI.ListWorkspaceConfigs(ctx).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			workspaceConfig := selection.GetWorkspaceConfigFromPrompt(workspaceConfigList, 0, false, false, "Make Default")
			if workspaceConfig == nil {
				return nil
			}
			workspaceConfigName = workspaceConfig.Name
		} else {
			workspaceConfigName = args[0]
		}

		res, err := apiClient.WorkspaceConfigAPI.SetDefaultWorkspaceConfig(ctx, workspaceConfigName).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		views.RenderInfoMessage(fmt.Sprintf("Workspace config '%s' set as default", workspaceConfigName))
		return nil
	},
}

func init() {
}
