// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"github.com/daytonaio/daytona/internal/util"
	"github.com/spf13/cobra"
)

var TargetCmd = &cobra.Command{
	Use:     "target",
	Short:   "Manage provider targets",
	GroupID: util.SERVER_GROUP,
}

func init() {
	TargetCmd.AddCommand(targetListCmd)
	TargetCmd.AddCommand(TargetSetCmd)
	TargetCmd.AddCommand(targetRemoveCmd)
	TargetCmd.AddCommand(targetSetDefaultCmd)
	TargetSetCmd.Flags().StringVarP(&pipeFile, "file", "f", "", "Path to JSON file for target configuration, use '-' to read from stdin")
}
