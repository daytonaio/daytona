// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runner

import (
	"context"
	"errors"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/daytonaio/daytona/cmd/daytona/config"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	cmd_common "github.com/daytonaio/daytona/pkg/cmd/common"
	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/views"
	runner "github.com/daytonaio/daytona/pkg/views/server/runner/selection"
	views_util "github.com/daytonaio/daytona/pkg/views/util"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:     "delete [RUNNER]",
	Short:   "Delete a runner",
	Args:    cobra.RangeArgs(0, 1),
	Aliases: cmd_common.GetAliases("delete"),
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

		if selectedRunnerId == common.LOCAL_RUNNER_ID {
			return errors.New("to disable the local runner, use the 'daytona server configure' form")
		}

		var confirm bool

		if !yesFlag {
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						Title(fmt.Sprintf("Unregister runner %s?", selectedRunnerId)).
						Description("It is recommended that you remove all target configs, targets and workspaces associated with it.").
						Value(&confirm),
				),
			).WithTheme(views.GetCustomTheme())

			err := form.Run()
			if err != nil {
				return err
			}
		}

		res, err := apiClient.RunnerAPI.DeleteRunner(ctx, selectedRunnerId).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		views.RenderInfoMessageBold(fmt.Sprintf("Runner %s unregistered successfully", selectedRunnerId))
		return nil
	},
}

var yesFlag bool

func init() {
	deleteCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Confirm deletion without prompt")
}
