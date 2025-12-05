// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package profile

import (
	"github.com/daytonaio/daytona/cli/internal"
	"github.com/spf13/cobra"
)

var ProfileCmd = &cobra.Command{
	Use:     "profile",
	Short:   "Manage Daytona profiles",
	Long:    "Commands for managing Daytona profiles",
	Aliases: []string{"profiles"},
	GroupID: internal.USER_GROUP,
}

func init() {
	ProfileCmd.AddCommand(ListCmd)
	ProfileCmd.AddCommand(UseCmd)
	ProfileCmd.AddCommand(AddCmd)
	ProfileCmd.AddCommand(DeleteCmd)
}
