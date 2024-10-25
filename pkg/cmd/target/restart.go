// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/target/selection"
	"github.com/spf13/cobra"
)

var restartProjectFlag string

var RestartCmd = &cobra.Command{
	Use:     "restart [TARGET]",
	Short:   "Restart a target",
	Args:    cobra.RangeArgs(0, 1),
	GroupID: util.TARGET_GROUP,
	RunE: func(cmd *cobra.Command, args []string) error {
		var targetId string

		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		if len(args) == 0 {
			if restartProjectFlag != "" {
				err := cmd.Help()
				if err != nil {
					return err
				}
				return nil
			}

			targetList, res, err := apiClient.TargetAPI.ListTargets(ctx).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			target := selection.GetTargetFromPrompt(targetList, "Restart")
			if target == nil {
				return nil
			}
			targetId = target.Name
		} else {
			targetId = args[0]
		}

		err = RestartTarget(apiClient, targetId, restartProjectFlag)
		if err != nil {
			return err
		}
		if restartProjectFlag != "" {
			views.RenderInfoMessage(fmt.Sprintf("Project '%s' from target '%s' successfully restarted", restartProjectFlag, targetId))
		} else {
			views.RenderInfoMessage(fmt.Sprintf("Target '%s' successfully restarted", targetId))
		}
		return nil
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return getAllTargetsByState(TARGET_STATE_RUNNING)
	},
}

func init() {
	RestartCmd.Flags().StringVarP(&restartProjectFlag, "project", "p", "", "Restart a single project in the target (project name)")
}

func RestartTarget(apiClient *apiclient.APIClient, targetId, projectName string) error {
	err := StopTarget(apiClient, targetId, projectName)
	if err != nil {
		return err
	}
	return StartTarget(apiClient, targetId, projectName)
}
