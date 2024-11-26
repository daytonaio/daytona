// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package projectconfig

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/atotto/clipboard"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
	"github.com/spf13/cobra"
)

var projectConfigExportCmd = &cobra.Command{
	Use:     "export",
	Aliases: []string{"exp"},
	Short:   "Export a project config",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var selectedProjectConfig *apiclient.ProjectConfig
		var output string
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		if allFlag {
			projectConfigs, res, err := apiClient.ProjectConfigAPI.ListProjectConfigs(ctx).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			if len(projectConfigs) == 0 {
				views_util.NotifyEmptyProjectConfigList(true)
				return nil
			}

			for i := range projectConfigs {
				projectConfigs[i].GitProviderConfigId = nil
			}

			data, err := json.MarshalIndent(projectConfigs, "", "  ")
			if err != nil {
				return err
			}

			fmt.Println(string(data))

			if err := clipboard.WriteAll(string(data)); err == nil {
				output = "The configs have been copied to your clipboard."
			} else {
				output = "Could not copy the configs to your clipboard."
			}
			views.RenderContainerLayout(views.GetInfoMessage(output))

			return nil
		}

		if len(args) == 0 {
			projectConfigs, res, err := apiClient.ProjectConfigAPI.ListProjectConfigs(ctx).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			if len(projectConfigs) == 0 {
				views_util.NotifyEmptyProjectConfigList(true)
				return nil
			}

			selectedProjectConfig = selection.GetProjectConfigFromPrompt(projectConfigs, 0, false, false, "Export")
			if selectedProjectConfig == nil {
				return nil
			}
		} else {
			var res *http.Response
			selectedProjectConfig, res, err = apiClient.ProjectConfigAPI.GetProjectConfig(ctx, args[0]).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}
		}

		selectedProjectConfig.GitProviderConfigId = nil

		data, err := json.MarshalIndent(selectedProjectConfig, "", "  ")
		if err != nil {
			return err
		}

		fmt.Println(string(data))

		if err := clipboard.WriteAll(string(data)); err == nil {
			output = "The config has been copied to your clipboard."
		} else {
			output = "Could not copy the config to your clipboard."
		}
		views.RenderContainerLayout(views.GetInfoMessage(output))

		return nil
	},
}

func init() {
	projectConfigExportCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Export all project configs")
}
