// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"context"
	"fmt"
	"time"

	apiclient_cli "github.com/daytonaio/daytona/cli/apiclient"
	"github.com/daytonaio/daytona/cli/cmd/common"
	view_common "github.com/daytonaio/daytona/cli/views/common"
	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
	"github.com/spf13/cobra"
)

var (
	forceFlag       bool
	stopWaitFlag    bool
	stopTimeoutFlag time.Duration
)

var StopCmd = &cobra.Command{
	Use:   "stop [SANDBOX_ID | SANDBOX_NAME]",
	Short: "Stop a sandbox",
	Example: `  daytona stop my-sandbox
  daytona stop my-sandbox --wait --timeout 2m
  daytona stop my-sandbox --force --format json`,
	Args: requireSandboxArg,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		apiClient, err := apiclient_cli.GetApiClient(nil, nil)
		if err != nil {
			return err
		}

		sandboxIdOrNameArg := args[0]

		// Pre-check so stopping an already-stopped sandbox succeeds idempotently.
		sandbox, res, err := apiClient.SandboxAPI.GetSandbox(ctx, sandboxIdOrNameArg).Execute()
		if err != nil {
			return apiclient_cli.HandleErrorResponse(res, err)
		}

		if sandbox.State != nil && *sandbox.State == apiclient.SANDBOXSTATE_STOPPED {
			if common.FormatFlag != "" {
				common.NewFormatter(sandbox).Print()
				return nil
			}
			view_common.RenderInfoMessageBold(fmt.Sprintf("Sandbox %s is already stopped", sandboxIdOrNameArg))
			return nil
		}

		req := apiClient.SandboxAPI.StopSandbox(ctx, sandboxIdOrNameArg)
		if forceFlag {
			req = req.Force(forceFlag)
		}
		_, res, err = req.Execute()
		if err != nil {
			return apiclient_cli.HandleErrorResponse(res, err)
		}

		if stopWaitFlag {
			err = common.AwaitSandboxState(ctx, apiClient, sandboxIdOrNameArg, stopTimeoutFlag, apiclient.SANDBOXSTATE_STOPPED)
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

		view_common.RenderInfoMessageBold(fmt.Sprintf("Sandbox %s stopped", sandboxIdOrNameArg))
		return nil
	},
}

func init() {
	StopCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Force stop the sandbox using SIGKILL")
	StopCmd.Flags().BoolVar(&stopWaitFlag, "wait", false, "Wait until the sandbox is stopped")
	StopCmd.Flags().DurationVar(&stopTimeoutFlag, "timeout", 5*time.Minute, "Maximum time to wait with --wait (0 waits indefinitely)")
	common.RegisterFormatFlagNoShorthand(StopCmd)
}
