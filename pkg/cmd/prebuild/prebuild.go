// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package prebuild

import (
	"github.com/daytonaio/daytona/internal/util"
	"github.com/spf13/cobra"
)

var PrebuildCmd = &cobra.Command{
	Use:     "prebuild",
	Short:   "Manage prebuilds",
	Args:    cobra.NoArgs,
	GroupID: util.TARGET_GROUP,
	Aliases: []string{"prebuilds", "pb"},
}

func init() {
	PrebuildCmd.AddCommand(addCmd)
	PrebuildCmd.AddCommand(listCmd)
	PrebuildCmd.AddCommand(infoCmd)
	PrebuildCmd.AddCommand(updateCmd)
	PrebuildCmd.AddCommand(deleteCmd)
}
