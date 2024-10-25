// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"context"

	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/format"
	"github.com/daytonaio/daytona/pkg/views/target/info"
	"github.com/daytonaio/daytona/pkg/views/target/selection"
	"github.com/spf13/cobra"
)

var InfoCmd = &cobra.Command{
	Use:     "info [TARGET]",
	Short:   "Show target info",
	Aliases: []string{"view", "inspect"},
	Args:    cobra.RangeArgs(0, 1),
	GroupID: util.TARGET_GROUP,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		var target *apiclient.TargetDTO

		if len(args) == 0 {
			targetList, res, err := apiClient.TargetAPI.ListTargets(ctx).Verbose(true).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			if format.FormatFlag != "" {
				format.UnblockStdOut()
			}

			target = selection.GetTargetFromPrompt(targetList, "View")
			if format.FormatFlag != "" {
				format.BlockStdOut()
			}

		} else {
			target, err = apiclient_util.GetTarget(args[0], true)
			if err != nil {
				return err
			}
		}

		if target == nil {
			return nil
		}

		if format.FormatFlag != "" {
			formattedData := format.NewFormatter(target)
			formattedData.Print()
			return nil
		}

		info.Render(target, "", false)
		return nil
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return getTargetNameCompletions()
	},
}

func init() {
	format.RegisterFormatFlag(InfoCmd)
}
