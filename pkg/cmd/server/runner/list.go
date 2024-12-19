// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runner

import (
	"context"

	"github.com/daytonaio/daytona/pkg/cmd/format"
	"github.com/daytonaio/daytona/pkg/views/server/runner/list"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List runners",
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		runners, res, err := apiClient.RunnerAPI.ListRunners(ctx).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		if format.FormatFlag != "" {
			formattedData := format.NewFormatter(runners)
			formattedData.Print()
			return nil
		}

		list.ListRunners(runners)
		return nil
	},
}

func init() {
	format.RegisterFormatFlag(listCmd)
}
