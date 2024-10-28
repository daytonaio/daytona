// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaceconfig

import (
	"context"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/format"
	workspaceconfig_view "github.com/daytonaio/daytona/pkg/views/workspaceconfig/list"
	"github.com/spf13/cobra"
)

var workspaceConfigListCmd = &cobra.Command{
	Use:     "list",
	Short:   "Lists workspace configs",
	Aliases: []string{"ls"},
	Args:    cobra.NoArgs,
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

		workspaceConfigs, res, err := apiClient.WorkspaceConfigAPI.ListWorkspaceConfigs(context.Background()).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		if format.FormatFlag != "" {
			formattedData := format.NewFormatter(workspaceConfigs)
			formattedData.Print()
			return nil
		}

		workspaceconfig_view.ListWorkspaceConfigs(workspaceConfigs, apiServerConfig, specifyGitProviders)
		return nil
	},
}

func init() {
	format.RegisterFormatFlag(workspaceConfigListCmd)
}
