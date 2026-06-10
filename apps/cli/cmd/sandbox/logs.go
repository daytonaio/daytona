// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"context"
	"time"

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
	Use:     "logs [SANDBOX_ID | SANDBOX_NAME]",
	Short:   "View the build logs of a sandbox",
	Long:    "View the build logs of a sandbox. With --follow the logs are streamed until the sandbox build reaches a terminal state.",
	Args:    cobra.ExactArgs(1),
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

		logsCtx, stopLogs := context.WithCancel(ctx)
		defer stopLogs()

		streamDone := make(chan error, 1)
		go func() {
			streamDone <- common.ReadBuildLogs(logsCtx, params)
		}()

		for {
			sb, res, err := apiClient.SandboxAPI.GetSandbox(ctx, sandbox.Id).Execute()
			if err != nil {
				return apiclient_cli.HandleErrorResponse(res, err)
			}

			if sb.State != nil {
				if done, failed := isSandboxBuildDone(*sb.State); done {
					// Grace period so trailing log output is flushed before the
					// stream is canceled.
					time.Sleep(250 * time.Millisecond)
					stopLogs()
					if streamDone != nil {
						<-streamDone
					}
					if failed {
						if reason := sb.GetErrorReason(); reason != "" {
							return clierr.Newf(clierr.CategoryServer, "sandbox processing failed: %s", reason)
						}
						return clierr.New(clierr.CategoryServer, "sandbox processing failed")
					}
					return nil
				}
			}

			select {
			case err := <-streamDone:
				streamDone = nil
				if err != nil {
					return err
				}
			case <-time.After(time.Second):
			}
		}
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
