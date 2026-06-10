// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package snapshot

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
	Use:   "logs [SNAPSHOT_ID | SNAPSHOT_NAME]",
	Short: "View the build logs of a snapshot",
	Long:  "View the build logs of a snapshot. With --follow the logs are streamed until the snapshot build reaches a terminal state.",
	Example: `  daytona snapshot logs my-snapshot:1.0
  daytona snapshot logs my-snapshot:1.0 --follow`,
	Args:    requireSnapshotArg,
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

		return common.FollowBuildLogs(ctx, params, func(ctx context.Context) (bool, error) {
			snap, res, err := apiClient.SnapshotsAPI.GetSnapshot(ctx, snapshot.Id).Execute()
			if err != nil {
				return false, apiclient_cli.HandleErrorResponse(res, err)
			}
			done, failed := isSnapshotBuildDone(snap.State)
			if !done {
				return false, nil
			}
			if !failed {
				return true, nil
			}
			if reason := snap.GetErrorReason(); reason != "" {
				return true, clierr.Newf(clierr.CategoryServer, "snapshot processing failed: %s", reason)
			}
			return true, clierr.New(clierr.CategoryServer, "snapshot processing failed")
		})
	},
}

// requireSnapshotArg validates that exactly one snapshot ID or name argument
// is provided, returning a usage-category error otherwise.
func requireSnapshotArg(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return clierr.New(clierr.CategoryUsage, "missing required argument: snapshot ID or name")
	}
	if len(args) > 1 {
		return clierr.Newf(clierr.CategoryUsage, "expected a single snapshot ID or name argument, received %d arguments", len(args))
	}
	return nil
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
