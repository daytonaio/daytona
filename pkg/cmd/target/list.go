// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"context"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/format"
	list_view "github.com/daytonaio/daytona/pkg/views/target/list"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List targets",
	Args:    cobra.ExactArgs(0),
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		apiClient, err := apiclient.GetApiClient(nil)
		if err != nil {
			return err
		}

		targetList, res, err := apiClient.TargetAPI.ListTargets(ctx).ShowOptions(showOptions).Execute()

		if err != nil {
			return apiclient.HandleErrorResponse(res, err)
		}

		if format.FormatFlag != "" {
			formattedData := format.NewFormatter(targetList)
			formattedData.Print()
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

		list_view.ListTargets(targetList, activeProfile.Name)
		return nil
	},
}

func init() {
	listCmd.Flags().BoolVarP(&showOptions, "show-options", "v", false, "Show target options")
	format.RegisterFormatFlag(listCmd)
}
