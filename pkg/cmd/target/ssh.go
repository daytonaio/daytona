// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"context"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/common"
	"github.com/daytonaio/daytona/pkg/ide"
	ide_views "github.com/daytonaio/daytona/pkg/views/ide"
	"github.com/daytonaio/daytona/pkg/views/target/selection"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/spf13/cobra"
)

var sshOptions []string

var sshCmd = &cobra.Command{
	Use:   "ssh [TARGET] [CMD...]",
	Short: "SSH into a target using the terminal",
	Args:  cobra.ArbitraryArgs,
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
		var tg *apiclient.TargetDTO

		apiClient, err := apiclient_util.GetApiClient(&activeProfile)
		if err != nil {
			return err
		}

		if len(args) == 0 {
			targetList, res, err := apiClient.TargetAPI.ListTargets(ctx).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			if len(targetList) == 0 {
				views_util.NotifyEmptyWorkspaceList(true)
				return nil
			}

			tg = selection.GetTargetFromPrompt(targetList, false, "SSH Into")
			if tg == nil {
				return nil
			}
		} else {
			tg, _, err = apiclient_util.GetTarget(args[0])
			if err != nil {
				return err
			}
		}

		if tg.TargetConfig.ProviderInfo.AgentlessTarget != nil && *tg.TargetConfig.ProviderInfo.AgentlessTarget {
			return agentlessTargetError(tg.TargetConfig.ProviderInfo.Name)
		}

		if tg.State.Name == apiclient.ResourceStateNameStopped {
			tgRunningStatus, err := autoStartTarget(*tg)
			if err != nil {
				return err
			}
			if !tgRunningStatus {
				return nil
			}
		}

		sshArgs := []string{}
		if len(args) > 1 {
			sshArgs = append(sshArgs, args[1:]...)
		}

		return ide.OpenTerminalSsh(activeProfile, tg.Id, nil, sshOptions, sshArgs...)
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return common.GetWorkspaceNameCompletions()
	},
}

func init() {
	sshCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Automatically confirm any prompts")
	sshCmd.Flags().StringArrayVarP(&sshOptions, "option", "o", []string{}, "Specify SSH options in KEY=VALUE format.")
}

func autoStartTarget(target apiclient.TargetDTO) (bool, error) {
	if !yesFlag {
		if !ide_views.RunStartTargetForm(target.Name) {
			return false, nil
		}
	}

	apiClient, err := apiclient_util.GetApiClient(nil)
	if err != nil {
		return false, err
	}

	err = StartTarget(apiClient, target)
	if err != nil {
		return false, err
	}

	return true, nil
}
