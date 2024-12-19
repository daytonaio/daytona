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
	"github.com/daytonaio/daytona/pkg/views/target/selection"
	views_util "github.com/daytonaio/daytona/pkg/views/util"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var allFlag bool

var startCmd = &cobra.Command{
	Use:   "start [TARGET]",
	Short: "Start a target",
	Args:  cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var selectedTargetsNames []string

		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		if allFlag {
			return startAllTargets()
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
			selectedTargets := selection.GetTargetsFromPrompt(targetList, "Start")
			for _, targets := range selectedTargets {
				selectedTargetsNames = append(selectedTargetsNames, targets.Name)
			}
		} else {
			selectedTargetsNames = append(selectedTargetsNames, args[0])
		}

		if len(selectedTargetsNames) == 1 {
			targetName := selectedTargetsNames[0]

			target, _, err := apiclient_util.GetTarget(targetName)
			if err != nil {
				return err
			}

			err = StartTarget(apiClient, *target)
			if err != nil {
				return err
			}

			views.RenderInfoMessage(fmt.Sprintf("Target '%s' started successfully", targetName))
		} else {
			for _, targetName := range selectedTargetsNames {
				target, _, err := apiclient_util.GetTarget(targetName)
				if err != nil {
					return err
				}

				err = StartTarget(apiClient, *target)
				if err != nil {
					log.Errorf("Failed to start target %s: %v\n\n", targetName, err)
					continue
				}
				views.RenderInfoMessage(fmt.Sprintf("- Target '%s' started successfully", targetName))
			}
		}
		return nil
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return getAllTargetsByState(util.Pointer(apiclient.ResourceStateNameStopped))
	},
}

func init() {
	startCmd.PersistentFlags().BoolVarP(&allFlag, "all", "a", false, "Start all targets")
	startCmd.PersistentFlags().BoolVarP(&yesFlag, "yes", "y", false, "Automatically confirm any prompts")
}

func startAllTargets() error {
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
		err := StartTarget(apiClient, target)
		if err != nil {
			log.Errorf("Failed to start target %s: %v\n\n", target.Name, err)
			continue
		}

		views.RenderInfoMessage(fmt.Sprintf("- Target '%s' started successfully", target.Name))
	}
	return nil
}

func getAllTargetsByState(state *apiclient.ModelsResourceStateName) ([]string, cobra.ShellCompDirective) {
	ctx := context.Background()
	apiClient, err := apiclient_util.GetApiClient(nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	targetList, _, err := apiClient.TargetAPI.ListTargets(ctx).Execute()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var choices []string
	for _, target := range targetList {
		if state == nil {
			choices = append(choices, target.Name)
			break
		} else {
			if *state == apiclient.ResourceStateNameStarted {
				choices = append(choices, target.Name)
				break
			}
			if *state == apiclient.ResourceStateNameStopped {
				choices = append(choices, target.Name)
				break
			}
		}
	}

	return choices, cobra.ShellCompDirectiveNoFileComp
}

func StartTarget(apiClient *apiclient.APIClient, target apiclient.TargetDTO) error {
	ctx := context.Background()
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

	if target.TargetConfig.ProviderInfo.AgentlessTarget != nil && *target.TargetConfig.ProviderInfo.AgentlessTarget {
		return agentlessTargetError(target.TargetConfig.ProviderInfo.Name)
	}

	logsContext, stopLogs := context.WithCancel(context.Background())
	go cmd_common.ReadTargetLogs(logsContext, cmd_common.ReadLogParams{
		Id:        target.Id,
		Label:     &target.Name,
		ServerUrl: activeProfile.Api.Url,
		ApiKey:    activeProfile.Api.Key,
		Follow:    util.Pointer(true),
		From:      &from,
	})

	res, err := apiClient.TargetAPI.StartTarget(ctx, target.Id).Execute()
	if err != nil {
		stopLogs()
		return apiclient_util.HandleErrorResponse(res, err)
	}

	err = cmd_common.AwaitTargetState(target.Id, apiclient.ResourceStateNameStarted)

	// Ensure reading remaining logs is completed
	time.Sleep(100 * time.Millisecond)

	stopLogs()
	return err
}

func agentlessTargetError(providerName string) error {
	return fmt.Errorf("%s does not require target state management; you may continue without starting or stopping it", providerName)
}
