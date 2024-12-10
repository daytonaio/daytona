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
	"github.com/daytonaio/daytona/pkg/cmd/format"
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

			return exportProjectConfigs(projectConfigs)
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

			if format.FormatFlag != "" {
				format.UnblockStdOut()
			}

			selectedProjectConfig = selection.GetProjectConfigFromPrompt(projectConfigs, 0, false, false, "Export")
			if selectedProjectConfig == nil {
				return nil
			}

			if format.FormatFlag != "" {
				format.BlockStdOut()
			}
		} else {
			var res *http.Response
			selectedProjectConfig, res, err = apiClient.ProjectConfigAPI.GetProjectConfig(ctx, args[0]).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}
		}

		return exportProjectConfigs([]apiclient.ProjectConfig{*selectedProjectConfig})
	},
}

func init() {
	projectConfigExportCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Export all project configs")
	format.RegisterFormatFlag(projectConfigExportCmd)
}

func exportProjectConfigs(projectConfigs []apiclient.ProjectConfig) error {
	if len(projectConfigs) == 0 {
		return nil
	}

	var pbFlag bool

	for i := range projectConfigs {
		projectConfigs[i].GitProviderConfigId = nil
		if projectConfigs[i].Prebuilds != nil {
			projectConfigs[i].Prebuilds = nil
			pbFlag = true
		}
	}

	var data []byte
	var err error

	if len(projectConfigs) == 1 {
		data, err = json.MarshalIndent(projectConfigs[0], "", "  ")
		views.RenderContainerLayout("Prebuilds have been removed from the config.")
	} else {
		data, err = json.MarshalIndent(projectConfigs, "", "  ")
		if pbFlag {
			views.RenderContainerLayout("Prebuilds have been removed from your configs.")
		}
	}

	if format.FormatFlag != "" {
		if len(projectConfigs) == 1 {
			formattedData := format.NewFormatter(projectConfigs[0])
			formattedData.Print()
		} else {
			formattedData := format.NewFormatter(projectConfigs)
			formattedData.Print()
		}
		return nil
	}

	if err != nil {
		return err
	}

	fmt.Println(string(data))

	if err := clipboard.WriteAll(string(data)); err == nil {
		views.RenderContainerLayout(views.GetInfoMessage("The config(s) have been copied to your clipboard."))
	} else {
		views.RenderContainerLayout(views.GetInfoMessage("Could not copy the config(s) to your clipboard."))
	}

	return nil
}
