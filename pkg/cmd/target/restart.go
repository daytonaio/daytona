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

var restartCmd = &cobra.Command{
	Use:   "restart [TARGET]",
	Short: "Restart a target",
	Args:  cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var target *apiclient.TargetDTO
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

			target = selection.GetTargetFromPrompt(targetList, false, "Restart")
			if target == nil {
				return nil
			}
		} else {
			target, _, err = apiclient_util.GetTarget(args[0])
			if err != nil {
				return err
			}
		}

		err = StartTarget(apiClient, *target, true)
		if err != nil {
			return err
		}
		views.RenderInfoMessage(fmt.Sprintf("Target '%s' successfully restarted", target.Name))
		return nil
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return getAllTargetsByState(util.Pointer(apiclient.ResourceStateNameStarted))
	},
}
