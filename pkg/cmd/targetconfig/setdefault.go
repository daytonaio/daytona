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

var setDefaultCmd = &cobra.Command{
	Use:   "set-default [CONFIG_NAME]",
	Short: "Set default target config",
	Args:  cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var configName string
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		if len(args) == 0 {
			targetConfigs, res, err := apiClient.TargetConfigAPI.ListTargetConfigs(ctx).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			if len(targetConfigs) == 0 {
				views_util.NotifyEmptyTargetConfigList(true)
				return nil
			}

			c, err := config.GetConfig()
			if err != nil {
				return err
			}

			activeProfile, err := c.GetActiveProfile()
			if err != nil {
				return err
			}

			selectedTargetConfig, err := targetconfig.GetTargetConfigFromPrompt(targetConfigs, activeProfile.Name, nil, false, "Make Default")
			if err != nil {
				if common.IsCtrlCAbort(err) {
					return nil
				} else {
					return err
				}
			}

			if selectedTargetConfig == nil {
				return nil
			}

			configName = selectedTargetConfig.Name
		} else {
			configName = args[0]
		}

		res, err := apiClient.TargetConfigAPI.SetDefaultTargetConfig(ctx, configName).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		views.RenderInfoMessage(fmt.Sprintf("Target config '%s' set as default", configName))
		return nil
	},
}
