// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package snapshot

import (
	"github.com/daytonaio/daytona/cli/internal"
	"github.com/spf13/cobra"
)

var SnapshotsCmd = &cobra.Command{
	Use:     "snapshot",
	Short:   "Manage Daytona snapshots",
	Long:    "Commands for managing Daytona snapshots",
	Aliases: []string{"snapshots"},
	GroupID: internal.SANDBOX_GROUP,
}

func init() {
	SnapshotsCmd.AddCommand(ListCmd)
	SnapshotsCmd.AddCommand(CreateCmd)
	SnapshotsCmd.AddCommand(PushCmd)
	SnapshotsCmd.AddCommand(DeleteCmd)
}
