// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/views"
	target_view "github.com/daytonaio/daytona/pkg/views/target"
	"github.com/spf13/cobra"
)

var targetSetDefaultCmd = &cobra.Command{
	Use:   "set-default [TARGET_NAME]",
	Short: "Set target to be used by default",
	Args:  cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var targetName string
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

			c, err := config.GetConfig()
			if err != nil {
				return err
			}

			activeProfile, err := c.GetActiveProfile()
			if err != nil {
				return err
			}

			selectedTarget, err := target_view.GetTargetFromPrompt(targetList, activeProfile.Name, nil, false, "Make Default")
			if err != nil {
				if common.IsCtrlCAbort(err) {
					return nil
				} else {
					return err
				}
			}

			if selectedTarget == nil {
				return nil
			}

			targetName = selectedTarget.Name
		} else {
			targetName = args[0]
		}

		res, err := apiClient.TargetAPI.SetDefaultTarget(ctx, targetName).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		views.RenderInfoMessage(fmt.Sprintf("Target '%s' set as default", targetName))
		return nil
	},
}
