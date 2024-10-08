// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"context"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/format"
	"github.com/daytonaio/daytona/pkg/views"
	list_view "github.com/daytonaio/daytona/pkg/views/target/list"
	"github.com/spf13/cobra"
)

var targetListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List targets",
	Args:    cobra.NoArgs,
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		targetList, res, err := apiClient.TargetAPI.ListTargets(ctx).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		if len(targetList) == 0 {
			views.RenderInfoMessageBold("No targets found")
			views.RenderInfoMessage("Use 'daytona target set' to add a target")
			return nil
		}

		if format.FormatFlag != "" {
			formattedData := format.NewFormatter(targetList)
			formattedData.Print()
			return nil
		}

		list_view.ListTargets(targetList)
		return nil
	},
}

func init() {
	format.RegisterFormatFlag(targetListCmd)
}
