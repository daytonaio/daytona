// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targetconfig

import (
	"context"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/format"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/targetconfig"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List target configs",
	Args:    cobra.NoArgs,
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		targetConfigs, res, err := apiClient.TargetConfigAPI.ListTargetConfigs(ctx).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		if len(targetConfigs) == 0 {
			views.RenderInfoMessageBold("No target configs found")
			views.RenderInfoMessage("Use 'daytona target-config set' to add a target config")
			return nil
		}

		if format.FormatFlag != "" {
			formattedData := format.NewFormatter(targetConfigs)
			formattedData.Print()
			return nil
		}

		targetconfig.ListTargetConfigs(targetConfigs)
		return nil
	},
}

func init() {
	format.RegisterFormatFlag(listCmd)
}
