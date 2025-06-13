// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"context"
	"fmt"

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

		sandboxList, res, err := apiClient.SandboxAPI.ListSandboxes(ctx).Execute()
		if err != nil {
			return apiclient.HandleErrorResponse(res, err)
		}

		infoArg := args[0]
		var sandboxCount int

		for _, s := range sandboxList {
			if s.Id == args[0] {
				infoArg = s.Id
				sandboxCount++
			}
		}

		switch sandboxCount {
		case 0:
			return fmt.Errorf("sandbox %s not found", args[0])
		case 1:
			sb, res, err := apiClient.SandboxAPI.GetSandbox(ctx, infoArg).Verbose(verboseFlag).Execute()
			if err != nil {
				return apiclient.HandleErrorResponse(res, err)
			}

			if common.FormatFlag != "" {
				formattedData := common.NewFormatter(sb)
				formattedData.Print()
				return nil
			}

			sandbox.RenderInfo(sb, false)
		default:
			return fmt.Errorf("multiple sandboxes with name %s found - please use the sandbox ID instead", args[0])
		}

		return nil
	},
}

func init() {
	InfoCmd.Flags().BoolVarP(&verboseFlag, "verbose", "v", false, "Include verbose output")
	common.RegisterFormatFlag(InfoCmd)
}
