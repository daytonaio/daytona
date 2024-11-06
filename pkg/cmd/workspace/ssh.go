// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/workspace/common"
	workspace_common "github.com/daytonaio/daytona/pkg/cmd/workspace/common"
	"github.com/daytonaio/daytona/pkg/ide"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var sshOptions []string

var SshCmd = &cobra.Command{
	Use:     "ssh [WORKSPACE] [CMD...]",
	Short:   "SSH into a workspace using the terminal",
	Args:    cobra.ArbitraryArgs,
	GroupID: util.TARGET_GROUP,
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
		var ws *apiclient.WorkspaceDTO
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

			if len(workspaceList) == 0 {
				views_util.NotifyEmptyWorkspaceList(true)
				return nil
			}

			ws = selection.GetWorkspaceFromPrompt(workspaceList, "SSH Into")
			if ws == nil {
				return nil
			}
		} else {
			ws, err = apiclient_util.GetWorkspace(args[0], true)
			if err != nil {
				return err
			}
		}

		if ws.State == nil || ws.State.Uptime < 1 {
			wsRunningStatus, err := AutoStartWorkspace(*ws)
			if err != nil {
				return err
			}
			if !wsRunningStatus {
				return nil
			}
		}

		sshArgs := []string{}
		if len(args) > 1 {
			sshArgs = append(sshArgs, args[1:]...)
		}

		gpgKey, err := workspace_common.GetGitProviderGpgKey(apiClient, ctx, providerConfigId)
		if err != nil {
			log.Warn(err)
		}

		return ide.OpenTerminalSsh(activeProfile, ws.Id, gpgKey, sshOptions, sshArgs...)
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return common.GetWorkspaceNameCompletions()
	},
}

func init() {
	SshCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Automatically confirm any prompts")
	SshCmd.Flags().StringArrayVarP(&sshOptions, "option", "o", []string{}, "Specify SSH options in KEY=VALUE format.")
}
