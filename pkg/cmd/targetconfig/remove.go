// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targetconfig

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/targetconfig"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/spf13/cobra"
)

var yesFlag bool

var removeCmd = &cobra.Command{
	Use:     "remove [TARGET_CONFIG]",
	Short:   "Remove target config",
	Args:    cobra.RangeArgs(0, 1),
	Aliases: []string{"rm", "delete"},
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

			selectedTargetConfig, err := targetconfig.GetTargetConfigFromPrompt(targetConfigs, activeProfile.Name, nil, false, "Remove")
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

		res, err := apiClient.TargetConfigAPI.RemoveTargetConfig(ctx, selectedConfigName).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		views.RenderInfoMessageBold(fmt.Sprintf("Target config %s removed successfully", selectedConfigName))
		return nil
	},
}

func init() {
	removeCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Confirm deletion of all targets without prompt")
}
