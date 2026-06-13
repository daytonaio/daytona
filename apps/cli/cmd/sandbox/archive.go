// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"context"
	"fmt"
	"time"

	apiclient_cli "github.com/daytonaio/daytona/cli/apiclient"
	"github.com/daytonaio/daytona/cli/cmd/common"
	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
	"github.com/spf13/cobra"
)

var (
	archiveWaitFlag    bool
	archiveTimeoutFlag time.Duration
)

var ArchiveCmd = &cobra.Command{
	Use:   "archive [SANDBOX_ID | SANDBOX_NAME]",
	Short: "Archive a sandbox",
	Example: `  daytona archive my-sandbox
  daytona archive my-sandbox --wait --timeout 10m
  daytona archive my-sandbox --wait --format json`,
	Args: requireSandboxArg,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		apiClient, err := apiclient_cli.GetApiClient(nil, nil)
		if err != nil {
			return err
		}

		sandboxIdOrNameArg := args[0]

		_, res, err := apiClient.SandboxAPI.ArchiveSandbox(ctx, sandboxIdOrNameArg).Execute()
		if err != nil {
			return apiclient_cli.HandleErrorResponse(res, err)
		}

		return finishLifecycleAction(ctx, apiClient, sandboxIdOrNameArg, archiveWaitFlag, archiveTimeoutFlag, apiclient.SANDBOXSTATE_ARCHIVED, fmt.Sprintf("Sandbox %s marked for archival", sandboxIdOrNameArg))
	},
}

func init() {
	ArchiveCmd.Flags().BoolVar(&archiveWaitFlag, "wait", false, "Wait until the sandbox is archived")
	ArchiveCmd.Flags().DurationVar(&archiveTimeoutFlag, "timeout", 5*time.Minute, "Maximum time to wait with --wait (0 waits indefinitely)")
	common.RegisterFormatFlag(ArchiveCmd)
}
