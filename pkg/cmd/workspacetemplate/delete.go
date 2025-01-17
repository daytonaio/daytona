// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspacetemplate

import (
	"context"
	"fmt"

	"github.com/charmbracelet/huh"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	cmd_common "github.com/daytonaio/daytona/pkg/cmd/common"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/selection"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var allFlag bool
var yesFlag bool
var forceFlag bool

var deleteCmd = &cobra.Command{
	Use:     "delete",
	Short:   "Delete a workspace template",
	Args:    cobra.MaximumNArgs(1),
	Aliases: cmd_common.GetAliases("delete"),
	RunE: func(cmd *cobra.Command, args []string) error {
		var selectedWorkspaceTemplate *apiclient.WorkspaceTemplate
		var selectedWorkspaceTemplateName string

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		if allFlag {
			if !yesFlag {
				form := huh.NewForm(
					huh.NewGroup(
						huh.NewConfirm().
							Title("Delete all workspace templates?").
							Description("Are you sure you want to delete all workspace templates?").
							Value(&yesFlag),
					),
				).WithTheme(views.GetCustomTheme())

				err := form.Run()
				if err != nil {
					return err
				}

				if !yesFlag {
					fmt.Println("Operation canceled.")
					return nil
				}
			}

			workspaceTemplates, res, err := apiClient.WorkspaceTemplateAPI.ListWorkspaceTemplates(context.Background()).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			if len(workspaceTemplates) == 0 {
				views_util.NotifyEmptyWorkspaceTemplateList(false)
				return nil
			}

			for _, workspaceTemplate := range workspaceTemplates {
				selectedWorkspaceTemplateName = workspaceTemplate.Name
				res, err := apiClient.WorkspaceTemplateAPI.DeleteWorkspaceTemplate(context.Background(), selectedWorkspaceTemplateName).Execute()
				if err != nil {
					log.Error(apiclient_util.HandleErrorResponse(res, err))
					continue
				}
				views.RenderInfoMessage("Deleted workspace template: " + selectedWorkspaceTemplateName)
			}
			return nil
		}

		if len(args) == 0 {
			workspaceTemplates, res, err := apiClient.WorkspaceTemplateAPI.ListWorkspaceTemplates(context.Background()).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			if len(workspaceTemplates) == 0 {
				views.RenderInfoMessage("No workspace templates found")
				return nil
			}

			selectedWorkspaceTemplate = selection.GetWorkspaceTemplateFromPrompt(workspaceTemplates, 0, false, false, "Delete")
			if selectedWorkspaceTemplate == nil {
				return nil
			}
			selectedWorkspaceTemplateName = selectedWorkspaceTemplate.Name
		} else {
			selectedWorkspaceTemplateName = args[0]
		}

		res, err := apiClient.WorkspaceTemplateAPI.DeleteWorkspaceTemplate(context.Background(), selectedWorkspaceTemplateName).Force(forceFlag).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		views.RenderInfoMessage("Workspace template deleted successfully")
		return nil
	},
}

func init() {
	deleteCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Delete all workspace templates")
	deleteCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Confirm deletion without prompt")
	deleteCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Force delete prebuild")
}
