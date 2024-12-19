// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runner

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/views"
	runner "github.com/daytonaio/daytona/pkg/views/server/runner/selection"
	views_util "github.com/daytonaio/daytona/pkg/views/util"

	"github.com/spf13/cobra"
)

var unregisterCmd = &cobra.Command{
	Use:   "unregister [RUNNER]",
	Short: "Unregister runner",
	Args:  cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var selectedRunnerId string

		ctx := context.Background()
		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		c, err := config.GetConfig()
		if err != nil {
			return err
		}

		if len(args) == 0 {
			activeProfile, err := c.GetActiveProfile()
			if err != nil {
				return err
			}

			runners, res, err := apiClient.RunnerAPI.ListRunners(ctx).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			if len(runners) == 0 {
				views_util.NotifyEmptyRunnerList(false)
				return nil
			}

			selectedRunner, err := runner.GetRunnerFromPrompt(runners, activeProfile.Name, "Unregister")
			if err != nil {
				if common.IsCtrlCAbort(err) {
					return nil
				} else {
					return err
				}
			}

			selectedRunnerId = selectedRunner.Id
		} else {
			selectedRunnerId = args[0]
		}

		res, err := apiClient.RunnerAPI.RemoveRunner(ctx, selectedRunnerId).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		views.RenderInfoMessageBold(fmt.Sprintf("Runenr %s unregistered successfully", selectedRunnerId))
		return nil
	},
}
