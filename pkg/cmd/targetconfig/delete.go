// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targetconfig

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	cmd_common "github.com/daytonaio/daytona/pkg/cmd/common"
	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/targetconfig"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/spf13/cobra"
)

var yesFlag bool

var deleteCmd = &cobra.Command{
	Use:     "delete [TARGET_CONFIG]",
	Short:   "Deletes a target config",
	Args:    cobra.RangeArgs(0, 1),
	Aliases: cmd_common.GetAliases("delete"),
	RunE: func(cmd *cobra.Command, args []string) error {
		var selectedConfigName string

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

			targetConfigs, res, err := apiClient.TargetConfigAPI.ListTargetConfigs(ctx).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			if len(targetConfigs) == 0 {
				views_util.NotifyEmptyTargetConfigList(false)
				return nil
			}

			selectedTargetConfig, err := targetconfig.GetTargetConfigFromPrompt(targetConfigs, activeProfile.Name, nil, false, "Delete")
			if err != nil {
				if common.IsCtrlCAbort(err) {
					return nil
				} else {
					return err
				}
			}

			selectedConfigName = selectedTargetConfig.Name
		} else {
			selectedConfigName = args[0]
		}

		res, err := apiClient.TargetConfigAPI.DeleteTargetConfig(ctx, selectedConfigName).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		views.RenderInfoMessageBold(fmt.Sprintf("Target config %s removed successfully", selectedConfigName))
		return nil
	},
}

func init() {
	deleteCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Confirm deletion of all targets without prompt")
}
