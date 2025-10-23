// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"context"

	"github.com/daytonaio/daytona/cli/apiclient"
	"github.com/daytonaio/daytona/cli/cmd/common"
	"github.com/spf13/cobra"
)

var LogsCmd = &cobra.Command{
	Use:     "logs [SANDBOX_ID] | [SANDBOX_NAME]",
	Short:   "Get sandbox logs",
	Args:    cobra.ExactArgs(1),
	Aliases: common.GetAliases("logs"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		apiClient, err := apiclient.GetApiClient(nil, nil)
		if err != nil {
			return err
		}

		sandboxIdOrNameArg := args[0]

		logs, res, err := apiClient.SandboxAPI.GetSandboxLogs(ctx, sandboxIdOrNameArg).Execute()
		if err != nil {
			return apiclient.HandleErrorResponse(res, err)
		}

		// Print the logs directly to stdout
		print(logs)

		return nil
	},
}

func init() {
	SandboxCmd.AddCommand(LogsCmd)
}
