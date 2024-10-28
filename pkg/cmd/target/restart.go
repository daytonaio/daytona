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
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/spf13/cobra"
)

var restartWorkspaceFlag string

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
			if restartWorkspaceFlag != "" {
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

			if len(targetList) == 0 {
				views_util.NotifyEmptyTargetList(true)
				return nil
			}

			target := selection.GetTargetFromPrompt(targetList, "Restart")
			if target == nil {
				return nil
			}
			targetId = target.Name
		} else {
			targetId = args[0]
		}

		err = RestartTarget(apiClient, targetId, restartWorkspaceFlag)
		if err != nil {
			return err
		}
		if restartWorkspaceFlag != "" {
			views.RenderInfoMessage(fmt.Sprintf("Workspace '%s' from target '%s' successfully restarted", restartWorkspaceFlag, targetId))
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
	RestartCmd.Flags().StringVarP(&restartWorkspaceFlag, "workspace", "w", "", "Restart a single workspace in the target (workspace name)")
}

func RestartTarget(apiClient *apiclient.APIClient, targetId, workspaceName string) error {
	err := StopTarget(apiClient, targetId, workspaceName)
	if err != nil {
		return err
	}
	return StartTarget(apiClient, targetId, workspaceName)
}
