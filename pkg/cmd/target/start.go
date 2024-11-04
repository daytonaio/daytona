// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"context"
	"fmt"
	"time"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/target/selection"
	views_util "github.com/daytonaio/daytona/pkg/views/util"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type TargetState string

const (
	TARGET_STATE_RUNNING TargetState = "Running"
	TARGET_STATE_STOPPED TargetState = "Unavailable"
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

			err = StartTarget(apiClient, targetName)
			if err != nil {
				return err
			}

			views.RenderInfoMessage(fmt.Sprintf("Target '%s' started successfully", targetName))
		} else {
			for _, target := range selectedTargetsNames {
				err := StartTarget(apiClient, target)
				if err != nil {
					log.Errorf("Failed to start target %s: %v\n\n", target, err)
					continue
				}
				views.RenderInfoMessage(fmt.Sprintf("- Target '%s' started successfully", target))
			}
		}
		return nil
	},
	// FIXME: add after adding state to targets
	// ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// 	return getAllTargetsByState(TARGET_STATE_STOPPED)
	// },
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
		err := StartTarget(apiClient, target.Name)
		if err != nil {
			log.Errorf("Failed to start target %s: %v\n\n", target.Name, err)
			continue
		}

		views.RenderInfoMessage(fmt.Sprintf("- Target '%s' started successfully", target.Name))
	}
	return nil
}

// FIXME: add target completions
// func getWorkspaceNameCompletions(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
// 	ctx := context.Background()
// 	apiClient, err := apiclient_util.GetApiClient(nil)
// 	if err != nil {
// 		return nil, cobra.ShellCompDirectiveDefault
// 	}

// 	targetId := args[0]
// 	target, _, err := apiClient.TargetAPI.GetTarget(ctx, targetId).Execute()
// 	if err != nil {
// 		return nil, cobra.ShellCompDirectiveDefault
// 	}

// 	var choices []string
// 	for _, workspace := range target.Workspaces {
// 		choices = append(choices, workspace.Name)
// 	}
// 	return choices, cobra.ShellCompDirectiveDefault
// }

func getTargetNameCompletions() ([]string, cobra.ShellCompDirective) {
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
	for _, v := range targetList {
		choices = append(choices, v.Name)
	}

	return choices, cobra.ShellCompDirectiveNoFileComp
}

// FIXME: fix this after adding state to targets
// func getAllTargetsByState(state TargetState) ([]string, cobra.ShellCompDirective) {
// 	ctx := context.Background()
// 	apiClient, err := apiclient_util.GetApiClient(nil)
// 	if err != nil {
// 		return nil, cobra.ShellCompDirectiveNoFileComp
// 	}

// 	targetList, _, err := apiClient.TargetAPI.ListTargets(ctx).Execute()
// 	if err != nil {
// 		return nil, cobra.ShellCompDirectiveNoFileComp
// 	}

// 	var choices []string
// 	for _, target := range targetList {
// 		for _, workspace := range target.Workspaces {
// 			if workspace.State == nil {
// 				continue
// 			}
// 			if state == TARGET_STATE_RUNNING && workspace.State.Uptime != 0 {
// 				choices = append(choices, target.Name)
// 				break
// 			}
// 			if state == TARGET_STATE_STOPPED && workspace.State.Uptime == 0 {
// 				choices = append(choices, target.Name)
// 				break
// 			}
// 		}
// 	}

// 	return choices, cobra.ShellCompDirectiveNoFileComp
// }

func StartTarget(apiClient *apiclient.APIClient, targetId string) error {
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

	target, err := apiclient_util.GetTarget(targetId, false)
	if err != nil {
		return err
	}

	logsContext, stopLogs := context.WithCancel(context.Background())
	go apiclient_util.ReadTargetLogs(logsContext, activeProfile, target.Id, true, &from)
	res, err := apiClient.TargetAPI.StartTarget(ctx, targetId).Execute()
	if err != nil {
		stopLogs()
		return apiclient_util.HandleErrorResponse(res, err)
	}
	time.Sleep(100 * time.Millisecond)
	stopLogs()
	return nil
}
