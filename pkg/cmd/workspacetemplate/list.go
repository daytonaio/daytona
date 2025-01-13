// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspacetemplate

import (
	"context"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	cmd_common "github.com/daytonaio/daytona/pkg/cmd/common"
	"github.com/daytonaio/daytona/pkg/cmd/format"
	workspacetemplate_view "github.com/daytonaio/daytona/pkg/views/workspacetemplate/list"
	"github.com/spf13/cobra"
)

var workspaceTemplateListCmd = &cobra.Command{
	Use:     "list",
	Short:   "Lists workspace templates",
	Args:    cobra.NoArgs,
	Aliases: cmd_common.GetAliases("list"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		var specifyGitProviders bool

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		gitProviders, res, err := apiClient.GitProviderAPI.ListGitProviders(ctx).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		if len(gitProviders) > 1 {
			specifyGitProviders = true
		}

		apiServerConfig, res, err := apiClient.ServerAPI.GetConfig(context.Background()).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		workspaceTemplates, res, err := apiClient.WorkspaceTemplateAPI.ListWorkspaceTemplates(context.Background()).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		if format.FormatFlag != "" {
			formattedData := format.NewFormatter(workspaceTemplates)
			formattedData.Print()
			return nil
		}

		workspacetemplate_view.ListWorkspaceTemplates(workspaceTemplates, apiServerConfig, specifyGitProviders)
		return nil
	},
}

func init() {
	format.RegisterFormatFlag(workspaceTemplateListCmd)
}
