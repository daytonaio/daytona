// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspacetemplate

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
	"github.com/daytonaio/daytona/pkg/views/selection"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:     "export",
	Aliases: []string{"exp"},
	Short:   "Export a workspace template",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var selectedWorkspaceTemplate *apiclient.WorkspaceTemplate
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		if allFlag {
			workspaceConfigs, res, err := apiClient.WorkspaceTemplateAPI.ListWorkspaceTemplates(ctx).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			if len(workspaceConfigs) == 0 {
				views_util.NotifyEmptyWorkspaceTemplateList(true)
				return nil
			}

			return exportWorkspaceTemplates(workspaceConfigs)
		}

		if len(args) == 0 {
			workspaceConfigs, res, err := apiClient.WorkspaceTemplateAPI.ListWorkspaceTemplates(ctx).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			if len(workspaceConfigs) == 0 {
				views_util.NotifyEmptyWorkspaceTemplateList(true)
				return nil
			}

			if format.FormatFlag != "" {
				format.UnblockStdOut()
			}

			selectedWorkspaceTemplate = selection.GetWorkspaceTemplateFromPrompt(workspaceConfigs, 0, false, false, "Export")
			if selectedWorkspaceTemplate == nil {
				return nil
			}

			if format.FormatFlag != "" {
				format.BlockStdOut()
			}
		} else {
			var res *http.Response
			selectedWorkspaceTemplate, res, err = apiClient.WorkspaceTemplateAPI.FindWorkspaceTemplate(ctx, args[0]).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}
		}

		return exportWorkspaceTemplates([]apiclient.WorkspaceTemplate{*selectedWorkspaceTemplate})
	},
}

func init() {
	exportCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Export all workspace templates")
	format.RegisterFormatFlag(exportCmd)
}

func exportWorkspaceTemplates(workspaceConfigs []apiclient.WorkspaceTemplate) error {
	if len(workspaceConfigs) == 0 {
		return nil
	}

	var pbFlag bool

	for i := range workspaceConfigs {
		workspaceConfigs[i].GitProviderConfigId = nil
		if workspaceConfigs[i].Prebuilds != nil {
			workspaceConfigs[i].Prebuilds = nil
			pbFlag = true
		}
	}

	var data []byte
	var err error

	if len(workspaceConfigs) == 1 {
		data, err = json.MarshalIndent(workspaceConfigs[0], "", "  ")
		views.RenderContainerLayout("Prebuilds have been removed from the template.")
	} else {
		data, err = json.MarshalIndent(workspaceConfigs, "", "  ")
		if pbFlag {
			views.RenderContainerLayout("Prebuilds have been removed from your templates.")
		}
	}

	if format.FormatFlag != "" {
		if len(workspaceConfigs) == 1 {
			formattedData := format.NewFormatter(workspaceConfigs[0])
			formattedData.Print()
		} else {
			formattedData := format.NewFormatter(workspaceConfigs)
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
