// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaceconfig

import (
	"context"
	"net/http"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/format"
	"github.com/daytonaio/daytona/pkg/views/target/selection"
	"github.com/daytonaio/daytona/pkg/views/workspaceconfig/info"
	"github.com/spf13/cobra"
)

var workspaceConfigInfoCmd = &cobra.Command{
	Use:     "info",
	Short:   "Show workspace config info",
	Aliases: []string{"view", "inspect"},
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		apiServerConfig, res, err := apiClient.ServerAPI.GetConfig(context.Background()).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		var workspaceConfig *apiclient.WorkspaceConfig

		if len(args) == 0 {
			workspaceConfigList, res, err := apiClient.WorkspaceConfigAPI.ListWorkspaceConfigs(ctx).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			if format.FormatFlag != "" {
				format.UnblockStdOut()
			}

			workspaceConfig = selection.GetWorkspaceConfigFromPrompt(workspaceConfigList, 0, false, false, "View")
			if format.FormatFlag != "" {
				format.BlockStdOut()
			}

		} else {
			var res *http.Response
			workspaceConfig, res, err = apiClient.WorkspaceConfigAPI.GetWorkspaceConfig(ctx, args[0]).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}
		}

		if workspaceConfig == nil {
			return nil
		}

		if format.FormatFlag != "" {
			formattedData := format.NewFormatter(workspaceConfig)
			formattedData.Print()
			return nil
		}

		info.Render(workspaceConfig, apiServerConfig, false)
		return nil
	},
}

func init() {
	format.RegisterFormatFlag(workspaceConfigInfoCmd)
}
