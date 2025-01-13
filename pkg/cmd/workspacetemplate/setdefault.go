// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspacetemplate

import (
	"context"
	"fmt"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	cmd_common "github.com/daytonaio/daytona/pkg/cmd/common"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/selection"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/spf13/cobra"
)

var workspaceTemplateSetDefaultCmd = &cobra.Command{
	Use:     "set-default",
	Short:   "Set workspace template info",
	Args:    cobra.MaximumNArgs(1),
	Aliases: cmd_common.GetAliases("set-default"),
	RunE: func(cmd *cobra.Command, args []string) error {
		var workspaceTemplateName string
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		if len(args) == 0 {
			workspaceTemplateList, res, err := apiClient.WorkspaceTemplateAPI.ListWorkspaceTemplates(ctx).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			if len(workspaceTemplateList) == 0 {
				views_util.NotifyEmptyWorkspaceTemplateList(true)
				return nil
			}

			workspaceTemplate := selection.GetWorkspaceTemplateFromPrompt(workspaceTemplateList, 0, false, false, "Make Default")
			if workspaceTemplate == nil {
				return nil
			}
			workspaceTemplateName = workspaceTemplate.Name
		} else {
			workspaceTemplateName = args[0]
		}

		res, err := apiClient.WorkspaceTemplateAPI.SetDefaultWorkspaceTemplate(ctx, workspaceTemplateName).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		views.RenderInfoMessage(fmt.Sprintf("Workspace template '%s' set as default", workspaceTemplateName))
		return nil
	},
}

func init() {
}
