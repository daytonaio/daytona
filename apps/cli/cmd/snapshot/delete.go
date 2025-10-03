// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package snapshot

import (
	"context"
	"fmt"

	"github.com/daytonaio/apiclient"
	apiclient_cli "github.com/daytonaio/daytona/cli/apiclient"
	"github.com/daytonaio/daytona/cli/cmd/common"
	view_common "github.com/daytonaio/daytona/cli/views/common"
	"github.com/spf13/cobra"
)

var DeleteCmd = &cobra.Command{
	Use:     "delete [SNAPSHOT_ID]",
	Short:   "Delete a snapshot",
	Args:    cobra.MaximumNArgs(1),
	Aliases: common.GetAliases("delete"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		apiClient, err := apiclient_cli.GetApiClient(nil, nil)
		if err != nil {
			return err
		}

		// Handle case when no snapshot ID is provided and allFlag is true
		if len(args) == 0 {
			if allFlag {
				page := float32(1.0)
				limit := float32(200.0) // 200 is the maximum limit for the API
				var allSnapshots []apiclient.SnapshotDto

				for {
					snapshotBatch, res, err := apiClient.SnapshotsAPI.GetAllSnapshots(ctx).Page(page).Limit(limit).Execute()
					if err != nil {
						return apiclient_cli.HandleErrorResponse(res, err)
					}

					allSnapshots = append(allSnapshots, snapshotBatch.Items...)

					if len(snapshotBatch.Items) < int(limit) || page >= snapshotBatch.TotalPages {
						break
					}
					page++
				}

				if len(allSnapshots) == 0 {
					view_common.RenderInfoMessageBold("No snapshots to delete")
					return nil
				}

				var deletedCount int

				for _, snapshot := range allSnapshots {
					res, err := apiClient.SnapshotsAPI.RemoveSnapshot(ctx, snapshot.Id).Execute()
					if err != nil {
						fmt.Printf("Failed to delete snapshot %s: %s\n", snapshot.Id, apiclient_cli.HandleErrorResponse(res, err))
					} else {
						deletedCount++
					}
				}

				view_common.RenderInfoMessageBold(fmt.Sprintf("Deleted %d snapshots", deletedCount))
				return nil
			}
			return cmd.Help()
		}

		// Handle case when a snapshot ID is provided
		snapshotIdArg := args[0]

		res, err := apiClient.SnapshotsAPI.RemoveSnapshot(ctx, snapshotIdArg).Execute()
		if err != nil {
			return apiclient_cli.HandleErrorResponse(res, err)
		}

		view_common.RenderInfoMessageBold(fmt.Sprintf("Snapshot %s deleted", snapshotIdArg))

		return nil
	},
}

var allFlag bool

func init() {
	DeleteCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Delete all snapshots")
}
