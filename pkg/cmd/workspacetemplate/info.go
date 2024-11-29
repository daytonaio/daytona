// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspacetemplate

import (
	"context"
	"net/http"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/format"
	"github.com/daytonaio/daytona/pkg/views/selection"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/daytonaio/daytona/pkg/views/workspacetemplate/info"
	"github.com/spf13/cobra"
)

var workspaceTemplateInfoCmd = &cobra.Command{
	Use:     "info",
	Short:   "Show workspace template info",
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

		var workspaceTemplate *apiclient.WorkspaceTemplate

		if len(args) == 0 {
			workspaceTemplateList, res, err := apiClient.WorkspaceTemplateAPI.ListWorkspaceTemplates(ctx).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			if len(workspaceTemplateList) == 0 {
				views_util.NotifyEmptyWorkspaceTemplateList(true)
				return nil
			}

			if format.FormatFlag != "" {
				format.UnblockStdOut()
			}

			workspaceTemplate = selection.GetWorkspaceTemplateFromPrompt(workspaceTemplateList, 0, false, false, "View")
			if format.FormatFlag != "" {
				format.BlockStdOut()
			}

		} else {
			var res *http.Response
			workspaceTemplate, res, err = apiClient.WorkspaceTemplateAPI.GetWorkspaceTemplate(ctx, args[0]).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}
		}

		if workspaceTemplate == nil {
			return nil
		}

		if format.FormatFlag != "" {
			formattedData := format.NewFormatter(workspaceTemplate)
			formattedData.Print()
			return nil
		}

		info.Render(workspaceTemplate, apiServerConfig, false)
		return nil
	},
}

func init() {
	format.RegisterFormatFlag(workspaceTemplateInfoCmd)
}
