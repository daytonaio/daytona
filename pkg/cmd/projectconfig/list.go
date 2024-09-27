// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package projectconfig

import (
	"context"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/format"
	"github.com/daytonaio/daytona/pkg/views"
	projectconfig_view "github.com/daytonaio/daytona/pkg/views/projectconfig/list"
	"github.com/spf13/cobra"
)

var projectConfigListCmd = &cobra.Command{
	Use:     "list",
	Short:   "Lists project configs",
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

		projectConfigs, res, err := apiClient.ProjectConfigAPI.ListProjectConfigs(context.Background()).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		if len(projectConfigs) == 0 {
			views.RenderInfoMessage("No project configs found. Add a new project config by running 'daytona project-config add'")
			return nil
		}

		if format.FormatFlag != "" {
			formattedData := format.NewFormatter(projectConfigs)
			formattedData.Print()
			return nil
		}

		projectconfig_view.ListProjectConfigs(projectConfigs, apiServerConfig, specifyGitProviders)
		return nil
	},
}

func init() {
	format.RegisterFormatFlag(projectConfigListCmd)
}
