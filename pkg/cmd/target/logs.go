// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"context"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	cmd_common "github.com/daytonaio/daytona/pkg/cmd/common"
	"github.com/daytonaio/daytona/pkg/views/target/selection"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:     "logs [TARGET]",
	Short:   "View the logs of a target",
	Args:    cobra.RangeArgs(0, 1),
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

		var target *apiclient.TargetDTO
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
				views_util.NotifyEmptyTargetList(true)
				return nil
			}
			target = selection.GetTargetFromPrompt(targetList, false, "Get Logs For")
			if target == nil {
				return nil
			}
		} else {
			target, _, err = apiclient_util.GetTarget(args[0])
			if err != nil {
				return err
			}
		}

		cmd_common.ReadTargetLogs(ctx, cmd_common.ReadLogParams{
			Id:        target.Id,
			Label:     &target.Name,
			ServerUrl: activeProfile.Api.Url,
			ApiKey:    activeProfile.Api.Key,
			Follow:    &followFlag,
		})

		return nil
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return getAllTargetsByState(nil)
	},
}

var followFlag bool

func init() {
	logsCmd.Flags().BoolVarP(&followFlag, "follow", "f", false, "Follow logs")
}
