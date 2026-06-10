// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"context"
	"fmt"
	"time"

	apiclient_cli "github.com/daytonaio/daytona/cli/apiclient"
	"github.com/daytonaio/daytona/cli/cmd/common"
	"github.com/daytonaio/daytona/cli/internal/clierr"
	view_common "github.com/daytonaio/daytona/cli/views/common"
	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
	"github.com/spf13/cobra"
)

// requireSandboxArg validates that exactly one sandbox ID or name argument
// is provided, returning a usage-category error otherwise.
func requireSandboxArg(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return clierr.New(clierr.CategoryUsage, "missing required argument: sandbox ID or name")
	}
	if len(args) > 1 {
		return clierr.Newf(clierr.CategoryUsage, "expected a single sandbox ID or name argument, received %d arguments", len(args))
	}
	return nil
}

var (
	startWaitFlag    bool
	startTimeoutFlag time.Duration
)

var StartCmd = &cobra.Command{
	Use:   "start [SANDBOX_ID | SANDBOX_NAME]",
	Short: "Start a sandbox",
	Example: `  daytona start my-sandbox
  daytona start my-sandbox --wait --timeout 2m
  daytona start my-sandbox --wait --format json`,
	Args: requireSandboxArg,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		apiClient, err := apiclient_cli.GetApiClient(nil, nil)
		if err != nil {
			return err
		}

		sandboxIdOrNameArg := args[0]

		_, res, err := apiClient.SandboxAPI.StartSandbox(ctx, sandboxIdOrNameArg).Execute()
		if err != nil {
			return apiclient_cli.HandleErrorResponse(res, err)
		}

		if startWaitFlag {
			err = common.AwaitSandboxState(ctx, apiClient, sandboxIdOrNameArg, startTimeoutFlag, apiclient.SANDBOXSTATE_STARTED)
			if err != nil {
				return err
			}
		}

		if common.FormatFlag != "" {
			sandbox, res, err := apiClient.SandboxAPI.GetSandbox(ctx, sandboxIdOrNameArg).Execute()
			if err != nil {
				return apiclient_cli.HandleErrorResponse(res, err)
			}
			common.NewFormatter(sandbox).Print()
			return nil
		}

		view_common.RenderInfoMessageBold(fmt.Sprintf("Sandbox %s started", sandboxIdOrNameArg))

		return nil
	},
}

func init() {
	StartCmd.Flags().BoolVar(&startWaitFlag, "wait", false, "Wait until the sandbox is started")
	StartCmd.Flags().DurationVar(&startTimeoutFlag, "timeout", 5*time.Minute, "Maximum time to wait with --wait (0 waits indefinitely)")
	common.RegisterFormatFlag(StartCmd)
}
