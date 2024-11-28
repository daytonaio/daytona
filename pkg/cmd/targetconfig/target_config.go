// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targetconfig

import (
	"github.com/daytonaio/daytona/internal/util"
	"github.com/spf13/cobra"
)

var TargetConfigCmd = &cobra.Command{
	Use:     "target-config",
	Aliases: []string{"tc"},
	Short:   "Manage target configs",
	GroupID: util.SERVER_GROUP,
}

func init() {
	TargetConfigCmd.AddCommand(listCmd)
	TargetConfigCmd.AddCommand(TargetConfigAddCmd)
	TargetConfigCmd.AddCommand(removeCmd)
}
