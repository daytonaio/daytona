// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package snapshot

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
	Use:   "logs [SNAPSHOT_ID | SNAPSHOT_NAME]",
	Short: "View the build logs of a snapshot",
	Long:  "View the build logs of a snapshot. With --follow the logs are streamed until the snapshot build reaches a terminal state.",
	Example: `  daytona snapshot logs my-snapshot:1.0
  daytona snapshot logs my-snapshot:1.0 --follow`,
	Args:    cobra.ExactArgs(1),
	Aliases: common.GetAliases("logs"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		apiClient, err := apiclient_cli.GetApiClient(nil, nil)
		if err != nil {
			return err
		}

		snapshot, res, err := apiClient.SnapshotsAPI.GetSnapshot(ctx, args[0]).Execute()
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
			Id:                   snapshot.Id,
			ServerUrl:            activeProfile.Api.Url,
			ServerApi:            activeProfile.Api,
			ActiveOrganizationId: activeProfile.ActiveOrganizationId,
			Follow:               util.Pointer(logsFollowFlag),
			ResourceType:         common.ResourceTypeSnapshot,
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
			snap, res, err := apiClient.SnapshotsAPI.GetSnapshot(ctx, snapshot.Id).Execute()
			if err != nil {
				return apiclient_cli.HandleErrorResponse(res, err)
			}

			if done, failed := isSnapshotBuildDone(snap.State); done {
				// Grace period so trailing log output is flushed before the
				// stream is canceled.
				time.Sleep(250 * time.Millisecond)
				stopLogs()
				if streamDone != nil {
					<-streamDone
				}
				if failed {
					if reason := snap.GetErrorReason(); reason != "" {
						return clierr.Newf(clierr.CategoryServer, "snapshot processing failed: %s", reason)
					}
					return clierr.New(clierr.CategoryServer, "snapshot processing failed")
				}
				return nil
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

// isSnapshotBuildDone reports whether the snapshot state is terminal for log
// streaming purposes and whether it represents a failure.
func isSnapshotBuildDone(state apiclient.SnapshotState) (done bool, failed bool) {
	switch state {
	case apiclient.SNAPSHOTSTATE_ERROR, apiclient.SNAPSHOTSTATE_BUILD_FAILED:
		return true, true
	case apiclient.SNAPSHOTSTATE_ACTIVE, apiclient.SNAPSHOTSTATE_INACTIVE:
		return true, false
	default:
		return false, false
	}
}

func init() {
	LogsCmd.Flags().BoolVarP(&logsFollowFlag, "follow", "f", false, "Follow the logs until the snapshot build reaches a terminal state")
}
