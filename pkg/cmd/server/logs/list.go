// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package logs

import (
	"context"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/common"
	"github.com/daytonaio/daytona/pkg/views/server"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "Lists Daytona Server Log Files",
	Aliases: common.GetAliases("list"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		apiclient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		logFiles, res, err := apiclient.ServerAPI.GetServerLogFiles(ctx).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		server.ListServerLogs(logFiles)

		return nil
	},
}
