// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package image

import (
	"github.com/daytonaio/daytona/cli/internal"
	"github.com/spf13/cobra"
)

var ImagesCmd = &cobra.Command{
	Use:     "image",
	Short:   "Manage Daytona images",
	Long:    "Commands for managing Daytona images",
	Aliases: []string{"images"},
	GroupID: internal.SANDBOX_GROUP,
}

func init() {
	ImagesCmd.AddCommand(ListCmd)
	ImagesCmd.AddCommand(CreateCmd)
	ImagesCmd.AddCommand(PushCmd)
	ImagesCmd.AddCommand(DeleteCmd)
	ImagesCmd.AddCommand(BuildCmd)
}
