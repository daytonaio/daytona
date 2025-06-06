// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package snapshot

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/cli/apiclient"
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
		var snapshotId string
		var snapshotName string

		apiClient, err := apiclient.GetApiClient(nil, nil)
		if err != nil {
			return err
		}

		snapshotList, res, err := apiClient.SnapshotsAPI.GetAllSnapshots(ctx).Execute()
		if err != nil {
			return apiclient.HandleErrorResponse(res, err)
		}

		if len(snapshotList.Items) == 0 {
			view_common.RenderInfoMessageBold("No snapshots to delete")
			return nil
		}

		if len(args) == 0 {
			if allFlag {
				for _, snapshot := range snapshotList.Items {
					res, err := apiClient.SnapshotsAPI.RemoveSnapshot(ctx, snapshot.Id).Execute()
					if err != nil {
						view_common.RenderInfoMessageBold(fmt.Sprintf("Failed to delete snapshot %s: %s", snapshot.Id, apiclient.HandleErrorResponse(res, err)))
					} else {
						view_common.RenderInfoMessageBold(fmt.Sprintf("Snapshot %s deleted", snapshot.Id))
					}
				}

				return nil
			}
			return cmd.Help()
		}

		for _, snapshot := range snapshotList.Items {
			if snapshot.Id == args[0] || snapshot.Name == args[0] {
				snapshotId = snapshot.Id
				snapshotName = snapshot.Name
				break
			}
		}

		res, err = apiClient.SnapshotsAPI.RemoveSnapshot(ctx, snapshotId).Execute()
		if err != nil {
			return apiclient.HandleErrorResponse(res, err)
		}

		view_common.RenderInfoMessageBold(fmt.Sprintf("Snapshot %s deleted", snapshotName))
		return nil
	},
}

var allFlag bool

func init() {
	DeleteCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Delete all snapshots")
}
