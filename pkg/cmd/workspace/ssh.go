// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	workspace_util "github.com/daytonaio/daytona/pkg/cmd/workspace/util"
	"github.com/daytonaio/daytona/pkg/ide"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var sshOptions []string

var SshCmd = &cobra.Command{
	Use:     "ssh [WORKSPACE] [PROJECT] [CMD...]",
	Short:   "SSH into a project using the terminal",
	Args:    cobra.ArbitraryArgs,
	GroupID: util.WORKSPACE_GROUP,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := config.GetConfig()
		if err != nil {
			return err
		}

		activeProfile, err := c.GetActiveProfile()
		if err != nil {
			return err
		}

		ctx := context.Background()
		var workspace *apiclient.WorkspaceDTO
		var projectName string
		var providerConfigId *string

		apiClient, err := apiclient_util.GetApiClient(&activeProfile)
		if err != nil {
			return err
		}

		if len(args) == 0 {
			workspaceList, res, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			workspace = selection.GetWorkspaceFromPrompt(workspaceList, "SSH Into")
			if workspace == nil {
				return nil
			}
		} else {
			workspace, err = apiclient_util.GetWorkspace(args[0], true)
			if err != nil {
				return err
			}
		}

		if len(args) == 0 || len(args) == 1 {
			selectedProject, err := selectWorkspaceProject(workspace.Id, &activeProfile)
			if err != nil {
				return err
			}
			if selectedProject == nil {
				return nil
			}
			projectName = selectedProject.Name
			providerConfigId = selectedProject.GitProviderConfigId
		}

		if len(args) >= 2 {
			projectName = args[1]
			for _, project := range workspace.Projects {
				if project.Name == projectName {
					providerConfigId = project.GitProviderConfigId
					break
				}
			}
		}

		if !workspace_util.IsProjectRunning(workspace, projectName) {
			wsRunningStatus, err := AutoStartWorkspace(workspace.Name, projectName)
			if err != nil {
				return err
			}
			if !wsRunningStatus {
				return nil
			}
		}

		sshArgs := []string{}
		if len(args) > 2 {
			sshArgs = append(sshArgs, args[2:]...)
		}

		gpgKey, err := GetGitProviderGpgKey(apiClient, ctx, providerConfigId)
		if err != nil {
			log.Warn(err)
		}

		return ide.OpenTerminalSsh(activeProfile, workspace.Id, projectName, gpgKey, sshOptions, sshArgs...)
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

func init() {
	SshCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Automatically confirm any prompts")
	SshCmd.Flags().StringArrayVarP(&sshOptions, "option", "o", []string{}, "Specify SSH options in KEY=VALUE format.")
}
