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
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/target/selection"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var stopWorkspaceFlag string

var StopCmd = &cobra.Command{
	Use:     "stop [TARGET]",
	Short:   "Stop a target",
	GroupID: util.TARGET_GROUP,
	Args:    cobra.RangeArgs(0, 1),
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
			if stopWorkspaceFlag != "" {
				return cmd.Help()
			}
			targetList, res, err := apiClient.TargetAPI.ListTargets(ctx).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			selectedTargets := selection.GetTargetsFromPrompt(targetList, "Stop")

			for _, target := range selectedTargets {
				err := StopTarget(apiClient, target.Name, "")
				if err != nil {
					log.Errorf("Failed to stop target %s: %v\n\n", target.Name, err)
					continue
				}

				workspaceNames := util.ArrayMap(target.Workspaces, func(w apiclient.Workspace) string {
					return w.Name
				})
				apiclient_util.ReadTargetLogs(ctx, activeProfile, target.Id, workspaceNames, false, true, &from)
				views.RenderInfoMessage(fmt.Sprintf("- Target '%s' successfully stopped", target.Name))
			}
		} else {
			targetId := args[0]
			var workspaceNames []string

			err = StopTarget(apiClient, targetId, stopWorkspaceFlag)
			if err != nil {
				return err
			}

			target, err := apiclient_util.GetTarget(targetId, false)
			if err != nil {
				return err
			}

			if startWorkspaceFlag != "" {
				workspaceNames = append(workspaceNames, stopWorkspaceFlag)
			} else {
				workspaceNames = util.ArrayMap(target.Workspaces, func(w apiclient.Workspace) string {
					return w.Name
				})
			}

			apiclient_util.ReadTargetLogs(ctx, activeProfile, target.Id, workspaceNames, false, true, &from)

			if stopWorkspaceFlag != "" {
				views.RenderInfoMessage(fmt.Sprintf("Workspace '%s' from target '%s' successfully stopped", stopWorkspaceFlag, targetId))
			} else {
				views.RenderInfoMessage(fmt.Sprintf("Target '%s' successfully stopped", targetId))
			}
		}
		return nil
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return getAllTargetsByState(TARGET_STATE_RUNNING)
	},
}

func init() {
	StopCmd.Flags().StringVarP(&stopWorkspaceFlag, "workspace", "w", "", "Stop a single workspace in the target (workspace name)")
	StopCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Stop all targets")
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
		err := StopTarget(apiClient, target.Name, "")
		if err != nil {
			log.Errorf("Failed to stop target %s: %v\n\n", target.Name, err)
			continue
		}

		workspaceNames := util.ArrayMap(target.Workspaces, func(w apiclient.Workspace) string {
			return w.Name
		})

		apiclient_util.ReadTargetLogs(ctx, activeProfile, target.Id, workspaceNames, false, true, &from)
		views.RenderInfoMessage(fmt.Sprintf("- Target '%s' successfully stopped", target.Name))
	}
	return nil
}

func StopTarget(apiClient *apiclient.APIClient, targetId, workspaceName string) error {
	ctx := context.Background()
	var message string
	var stopFunc func() error

	if workspaceName == "" {
		message = fmt.Sprintf("Target '%s' is stopping", targetId)
		stopFunc = func() error {
			res, err := apiClient.TargetAPI.StopTarget(ctx, targetId).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}
			return nil
		}
	} else {
		message = fmt.Sprintf("Workspace '%s' from target '%s' is stopping", workspaceName, targetId)
		stopFunc = func() error {
			res, err := apiClient.TargetAPI.StopWorkspace(ctx, targetId, workspaceName).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}
			return nil
		}
	}

	err := views_util.WithInlineSpinner(message, stopFunc)
	if err != nil {
		return err
	}

	return nil
}
