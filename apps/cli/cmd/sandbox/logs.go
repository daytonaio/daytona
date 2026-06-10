// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"context"

	apiclient_cli "github.com/daytonaio/daytona/cli/apiclient"
	"github.com/daytonaio/daytona/cli/cmd/common"
	"github.com/daytonaio/daytona/cli/config"
	"github.com/daytonaio/daytona/cli/internal/clierr"
	"github.com/daytonaio/daytona/cli/util"
	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
	"github.com/spf13/cobra"
)

var logsFollowFlag bool

var LogsCmd = &cobra.Command{
	Use:   "logs [SANDBOX_ID | SANDBOX_NAME]",
	Short: "View the build logs of a sandbox",
	Long:  "View the build logs of a sandbox. With --follow the logs are streamed until the sandbox build reaches a terminal state.",
	Example: `  daytona logs my-sandbox
  daytona logs my-sandbox --follow`,
	Args:    requireSandboxArg,
	Aliases: common.GetAliases("logs"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		apiClient, err := apiclient_cli.GetApiClient(nil, nil)
		if err != nil {
			return err
		}

		sandbox, res, err := apiClient.SandboxAPI.GetSandbox(ctx, args[0]).Execute()
		if err != nil {
			return apiclient_cli.HandleErrorResponse(res, err)
		}

		c, err := config.GetConfig()
		if err != nil {
			return err
		}

		activeProfile, err := c.GetActiveProfile()
		if err != nil {
			return err
		}

		params := common.ReadLogParams{
			Id:                   sandbox.Id,
			ServerUrl:            activeProfile.Api.Url,
			ServerApi:            activeProfile.Api,
			ActiveOrganizationId: activeProfile.ActiveOrganizationId,
			Follow:               util.Pointer(logsFollowFlag),
			ResourceType:         common.ResourceTypeSandbox,
		}

		if !logsFollowFlag {
			return common.ReadBuildLogs(ctx, params)
		}

		return common.FollowBuildLogs(ctx, params, func(ctx context.Context) (bool, error) {
			sb, res, err := apiClient.SandboxAPI.GetSandbox(ctx, sandbox.Id).Execute()
			if err != nil {
				return false, apiclient_cli.HandleErrorResponse(res, err)
			}
			if sb.State == nil {
				return false, nil
			}
			done, failed := isSandboxBuildDone(*sb.State)
			if !done {
				return false, nil
			}
			if !failed {
				return true, nil
			}
			if reason := sb.GetErrorReason(); reason != "" {
				return true, clierr.Newf(clierr.CategoryServer, "sandbox processing failed: %s", reason)
			}
			return true, clierr.New(clierr.CategoryServer, "sandbox processing failed")
		})
	},
}

// isSandboxBuildDone reports whether the sandbox state is terminal for log
// streaming purposes and whether it represents a failure.
func isSandboxBuildDone(state apiclient.SandboxState) (done bool, failed bool) {
	switch state {
	case apiclient.SANDBOXSTATE_ERROR, apiclient.SANDBOXSTATE_BUILD_FAILED:
		return true, true
	case apiclient.SANDBOXSTATE_STARTED,
		apiclient.SANDBOXSTATE_STOPPED,
		apiclient.SANDBOXSTATE_STOPPING,
		apiclient.SANDBOXSTATE_ARCHIVED,
		apiclient.SANDBOXSTATE_ARCHIVING,
		apiclient.SANDBOXSTATE_DESTROYED,
		apiclient.SANDBOXSTATE_DESTROYING:
		return true, false
	default:
		return false, false
	}
}

func init() {
	LogsCmd.Flags().BoolVarP(&logsFollowFlag, "follow", "f", false, "Follow the logs until the sandbox build reaches a terminal state")
}
