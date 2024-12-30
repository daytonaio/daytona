// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"errors"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
	"github.com/spf13/cobra"
)

var followFlag bool
var workspaceFlag bool

var logsCmd = &cobra.Command{
	Use:     "logs [WORKSPACE] [PROJECT_NAME]",
	Short:   "View logs for a workspace/project",
	Args:    cobra.RangeArgs(0, 2),
	GroupID: util.WORKSPACE_GROUP,
	Aliases: []string{"lg", "log"},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := config.GetConfig()
		if err != nil {
			return err
		}

		activeProfile, err := c.GetActiveProfile()
		if err != nil {
			return err
		}

		var workspace *apiclient.WorkspaceDTO
		apiClient, err := apiclient_util.GetApiClient(&activeProfile)
		if err != nil {
			return err
		}

		var (
			showWorkspaceLogs = true
			projectNames      []string
		)

		if len(args) == 0 {
			workspaceList, res, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}
			if len(workspaceList) == 0 {
				views.RenderInfoMessage("The workspace list is empty. Start off by running 'daytona create'.")
				return nil
			}
			workspace = selection.GetWorkspaceFromPrompt(workspaceList, "Get Logs For")
		} else {
			workspace, err = apiclient_util.GetWorkspace(args[0], false)
			if err != nil {
				return err
			}
		}

		if workspace == nil {
			return errors.New("workspace not found")
		} else if len(workspace.Projects) == 0 {
			return errors.New("no projects found in workspace")
		}

		if len(args) == 2 {
			projects := util.ArrayMap(workspace.Projects, func(p apiclient.Project) string {
				return p.Name
			})
			var found bool
			for _, project := range projects {
				if project == args[1] {
					found = true
					break
				}
			}
			if !found {
				return errors.New("project not found in workspace")
			}
			projectNames = append(projectNames, args[1])
			if workspaceFlag {
				showWorkspaceLogs = true
			} else {
				showWorkspaceLogs = false
			}
		} else if !workspaceFlag {
			projectNames = util.ArrayMap(workspace.Projects, func(p apiclient.Project) string {
				return p.Name
			})
		}

		apiclient_util.ReadWorkspaceLogs(ctx, activeProfile, workspace.Id, projectNames, followFlag, showWorkspaceLogs, nil)

		return nil
	},
}

func init() {
	logsCmd.Flags().BoolVarP(&followFlag, "follow", "f", false, "Follow logs")
	logsCmd.Flags().BoolVarP(&workspaceFlag, "workspace", "w", false, "View workspace logs")
}
