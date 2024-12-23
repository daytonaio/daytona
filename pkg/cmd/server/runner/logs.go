// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runner

import (
	"context"
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	cmd_common "github.com/daytonaio/daytona/pkg/cmd/common"
	"github.com/daytonaio/daytona/pkg/common"
	runner "github.com/daytonaio/daytona/pkg/views/server/runner/selection"
	views_util "github.com/daytonaio/daytona/pkg/views/util"

	"github.com/spf13/cobra"
)

var followFlag bool

var logsCmd = &cobra.Command{
	Use:     "logs [RUNNER_ID]",
	Short:   "View runner logs",
	Args:    cobra.MaximumNArgs(1),
	Aliases: []string{"log"},
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

		activeProfile, err := c.GetActiveProfile()
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

			selectedRunner, err := runner.GetRunnerFromPrompt(runners, activeProfile.Name, "View Logs")
			if err != nil {
				if common.IsCtrlCAbort(err) {
					return nil
				} else {
					return err
				}
			}

			selectedRunnerId = selectedRunner.Name
		} else {
			selectedRunnerId = args[0]
		}

		runner, res, err := apiClient.RunnerAPI.GetRunner(ctx, selectedRunnerId).Execute()
		if err != nil {
			if res.StatusCode == http.StatusNotFound {
				return fmt.Errorf("runner %s not found", selectedRunnerId)
			}
			return apiclient_util.HandleErrorResponse(res, err)
		}

		cmd_common.ReadRunnerLogs(ctx, cmd_common.ReadLogParams{
			Id:        runner.Id,
			Label:     &runner.Name,
			ServerUrl: activeProfile.Api.Url,
			ApiKey:    activeProfile.Api.Key,
			Index:     util.Pointer(0),
			Follow:    &followFlag,
		})
		return nil
	},
}

func init() {
	logsCmd.Flags().BoolVarP(&followFlag, "follow", "f", false, "Follow logs")
}
