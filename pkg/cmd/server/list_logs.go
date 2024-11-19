// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/views/server"
	"github.com/spf13/cobra"
)

var listServerLogsCmd = &cobra.Command{
	Use:     "log-list",
	Aliases: []string{"ls", "lls"},
	Short:   "Output Daytona Server logs",
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
