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
	Use:   "export",
	Short: "Export a workspace template",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var selectedWorkspaceTemplate *apiclient.WorkspaceTemplate
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		if allFlag {
			templates, res, err := apiClient.WorkspaceTemplateAPI.ListWorkspaceTemplates(ctx).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			if len(templates) == 0 {
				views_util.NotifyEmptyWorkspaceTemplateList(true)
				return nil
			}

			return exportWorkspaceTemplates(templates)
		}

		if len(args) == 0 {
			templates, res, err := apiClient.WorkspaceTemplateAPI.ListWorkspaceTemplates(ctx).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			if len(templates) == 0 {
				views_util.NotifyEmptyWorkspaceTemplateList(true)
				return nil
			}

			if format.FormatFlag != "" {
				format.UnblockStdOut()
			}

			selectedWorkspaceTemplate = selection.GetWorkspaceTemplateFromPrompt(templates, 0, false, false, "Export")
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

func exportWorkspaceTemplates(templates []apiclient.WorkspaceTemplate) error {
	if len(templates) == 0 {
		return nil
	}

	var pbFlag bool

	for i := range templates {
		templates[i].GitProviderConfigId = nil
		if templates[i].Prebuilds != nil {
			templates[i].Prebuilds = nil
			pbFlag = true
		}
	}

	data, err := json.MarshalIndent(templates, "", "  ")
	if pbFlag {
		views.RenderContainerLayout("Prebuilds have been removed from the export.")
	}
	if err != nil {
		return err
	}

	if format.FormatFlag != "" {
		if len(templates) == 1 {
			formattedData := format.NewFormatter(templates[0])
			formattedData.Print()
		} else {
			formattedData := format.NewFormatter(templates)
			formattedData.Print()
		}
		return nil
	}

	fmt.Println(string(data))

	if err := clipboard.WriteAll(string(data)); err == nil {
		views.RenderContainerLayout(views.GetInfoMessage("The export has been copied to your clipboard."))
	} else {
		views.RenderContainerLayout(views.GetInfoMessage("Could not copy the export to your clipboard."))
	}

	return nil
}
