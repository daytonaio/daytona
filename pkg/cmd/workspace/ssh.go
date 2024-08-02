// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	workspace_util "github.com/daytonaio/daytona/pkg/cmd/workspace/util"
	"github.com/daytonaio/daytona/pkg/ide"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var SshCmd = &cobra.Command{
	Use:     "ssh [WORKSPACE] [PROJECT]",
	Short:   "SSH into a project using the terminal",
	Args:    cobra.RangeArgs(0, 2),
	GroupID: util.WORKSPACE_GROUP,
	Run: func(cmd *cobra.Command, args []string) {
		c, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		activeProfile, err := c.GetActiveProfile()
		if err != nil {
			log.Fatal(err)
		}

		ctx := context.Background()
		var workspace *apiclient.WorkspaceDTO
		var projectName string

		apiClient, err := apiclient_util.GetApiClient(&activeProfile)
		if err != nil {
			log.Fatal(err)
		}

		if len(args) == 0 {
			workspaceList, res, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Execute()
			if err != nil {
				log.Fatal(apiclient_util.HandleErrorResponse(res, err))
			}

			workspace = selection.GetWorkspaceFromPrompt(workspaceList, "SSH Into")
			if workspace == nil {
				return
			}
		} else {
			workspace, err = apiclient_util.GetWorkspace(args[0])
			if err != nil {
				log.Fatal(err)
			}
		}

		if len(args) == 0 || len(args) == 1 {
			selectedProject, err := selectWorkspaceProject(*workspace.Id, &activeProfile)
			if err != nil {
				log.Fatal(err)
			}
			if selectedProject == nil {
				return
			}
			projectName = *selectedProject.Name
		}

		if len(args) == 2 {
			projectName = args[1]
		}

		if !workspace_util.IsProjectRunning(workspace, projectName) {
			views.RenderInfoMessage(fmt.Sprintf("Project '%s' from workspace '%s' is not in running state", projectName, *workspace.Name))
			return
		}

		err = ide.OpenTerminalSsh(activeProfile, *workspace.Id, projectName)
		if err != nil {
			log.Fatal(err)
		}
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) >= 2 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		if len(args) == 1 {
			return getProjectNameCompletions(cmd, args, toComplete)
		}

		return getWorkspaceNameCompletions()
	},
}
