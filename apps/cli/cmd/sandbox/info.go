// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"context"

	"github.com/daytonaio/daytona/cli/apiclient"
	"github.com/daytonaio/daytona/cli/cmd/common"
	"github.com/daytonaio/daytona/cli/views/sandbox"
	"github.com/spf13/cobra"
)

var InfoCmd = &cobra.Command{
	Use:     "info [SANDBOX_ID]",
	Short:   "Get sandbox info",
	Args:    cobra.ExactArgs(1),
	Aliases: common.GetAliases("info"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		apiClient, err := apiclient.GetApiClient(nil, nil)
		if err != nil {
			return err
		}

		sandboxIdArg := args[0]

		sb, res, err := apiClient.SandboxAPI.GetSandbox(ctx, sandboxIdArg).Execute()
		if err != nil {
			return apiclient.HandleErrorResponse(res, err)
		}

		if common.FormatFlag != "" {
			formattedData := common.NewFormatter(sb)
			formattedData.Print()
			return nil
		}

		sandbox.RenderInfo(sb, false)

		return nil
	},
}

func init() {
	common.RegisterFormatFlag(InfoCmd)
}
