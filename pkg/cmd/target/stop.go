// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"context"
	"fmt"
	"time"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	cmd_common "github.com/daytonaio/daytona/pkg/cmd/common"
	"github.com/daytonaio/daytona/pkg/views"
	logs_view "github.com/daytonaio/daytona/pkg/views/logs"
	"github.com/daytonaio/daytona/pkg/views/target/selection"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop [TARGET]",
	Short: "Stop a target",
	Args:  cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		timeFormat := time.Now().Format("2006-01-02 15:04:05")
		from, err := time.Parse("2006-01-02 15:04:05", timeFormat)
		if err != nil {
			return err
		}

		c, err := config.GetConfig()
		if err != nil {
			return err
		}

		activeProfile, err := c.GetActiveProfile()
		if err != nil {
			return err
		}

		if allFlag {
			return stopAllTargets(activeProfile, from)
		}

		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
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

			selectedTargets := selection.GetTargetsFromPrompt(targetList, "Stop")

			for _, target := range selectedTargets {
				err := StopTarget(apiClient, *target)
				if err != nil {
					log.Errorf("Failed to stop target %s: %v\n\n", target.Name, err)
					continue
				}

				logs_view.SetupLongestPrefixLength(util.ArrayMap(targetList, func(t apiclient.TargetDTO) string {
					return t.Name
				}))

				cmd_common.ReadTargetLogs(ctx, cmd_common.ReadLogParams{
					Id:                    target.Id,
					Label:                 &target.Name,
					ServerUrl:             activeProfile.Api.Url,
					ApiKey:                activeProfile.Api.Key,
					From:                  &from,
					SkipPrefixLengthSetup: true,
				})
				views.RenderInfoMessage(fmt.Sprintf("- Target '%s' successfully stopped", target.Name))
			}
		} else {
			targetId := args[0]

			target, _, err := apiclient_util.GetTarget(targetId)
			if err != nil {
				return err
			}

			err = StopTarget(apiClient, *target)
			if err != nil {
				return err
			}

			cmd_common.ReadTargetLogs(ctx, cmd_common.ReadLogParams{
				Id:        target.Id,
				Label:     &target.Name,
				ServerUrl: activeProfile.Api.Url,
				ApiKey:    activeProfile.Api.Key,
				From:      &from,
			})

			views.RenderInfoMessage(fmt.Sprintf("Target '%s' successfully stopped", targetId))
		}
		return nil
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return getAllTargetsByState(util.Pointer(apiclient.ResourceStateNameStarted))
	},
}

func init() {
	stopCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Stop all targets")
}

func stopAllTargets(activeProfile config.Profile, from time.Time) error {
	ctx := context.Background()
	apiClient, err := apiclient_util.GetApiClient(nil)
	if err != nil {
		return err
	}

	targetList, res, err := apiClient.TargetAPI.ListTargets(ctx).Execute()
	if err != nil {
		return apiclient_util.HandleErrorResponse(res, err)
	}

	for _, target := range targetList {
		err := StopTarget(apiClient, target)
		if err != nil {
			log.Errorf("Failed to stop target %s: %v\n\n", target.Name, err)
			continue
		}

		logs_view.SetupLongestPrefixLength(util.ArrayMap(targetList, func(t apiclient.TargetDTO) string {
			return t.Name
		}))

		cmd_common.ReadTargetLogs(ctx, cmd_common.ReadLogParams{
			Id:                    target.Id,
			Label:                 &target.Name,
			ServerUrl:             activeProfile.Api.Url,
			ApiKey:                activeProfile.Api.Key,
			From:                  &from,
			SkipPrefixLengthSetup: true,
		})
		views.RenderInfoMessage(fmt.Sprintf("- Target '%s' successfully stopped", target.Name))
	}
	return nil
}

func StopTarget(apiClient *apiclient.APIClient, target apiclient.TargetDTO) error {
	ctx := context.Background()

	if target.TargetConfig.ProviderInfo.AgentlessTarget != nil && *target.TargetConfig.ProviderInfo.AgentlessTarget {
		return agentlessTargetError(target.TargetConfig.ProviderInfo.Name)
	}

	err := views_util.WithInlineSpinner(fmt.Sprintf("Target '%s' is stopping", target.Name), func() error {
		res, err := apiClient.TargetAPI.StopTarget(ctx, target.Id).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		return cmd_common.AwaitTargetState(target.Id, apiclient.ResourceStateNameStarted)
	})
	if err != nil {
		return err
	}

	return nil
}
