// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"
	"os/exec"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/server"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var SshCmd = &cobra.Command{
	Use:   "ssh [WORKSPACE] [PROJECT]",
	Short: "SSH into a project using the terminal",
	Args:  cobra.RangeArgs(0, 2),
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
		var workspaceId string
		var projectName string

		apiClient, err := server.GetApiClient(&activeProfile)
		if err != nil {
			log.Fatal(err)
		}

		if len(args) == 0 {
			workspaceList, res, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Execute()
			if err != nil {
				log.Fatal(apiclient.HandleErrorResponse(res, err))
			}

			workspace := selection.GetWorkspaceFromPrompt(workspaceList, "ssh into")
			if workspace == nil {
				return
			}
			workspaceId = *workspace.Id
		} else {
			workspace, err := server.GetWorkspace(args[0])
			if err != nil {
				log.Fatal(err)
			}
			workspaceId = *workspace.Id
		}

		// Todo: make project_select_prompt view for 0 args
		if len(args) == 0 || len(args) == 1 {
			projectName, err = util.GetFirstWorkspaceProjectName(workspaceId, projectName, &activeProfile)
			if err != nil {
				log.Fatal(err)
			}
		}

		if len(args) == 2 {
			projectName = args[1]
		}

		err = config.EnsureSshConfigEntryAdded(activeProfile.Id, workspaceId, projectName)
		if err != nil {
			log.Fatal(err)
		}

		projectHostname := config.GetProjectHostname(activeProfile.Id, workspaceId, projectName)

		sshCommand := exec.Command("ssh", projectHostname)
		sshCommand.Stdin = cmd.InOrStdin()
		sshCommand.Stdout = cmd.OutOrStdout()
		sshCommand.Stderr = cmd.ErrOrStderr()

		err = sshCommand.Run()
		if err != nil {
			log.Fatal(err)
		}
	},
}
