// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"github.com/spf13/cobra"
)

var targetNameArg string

var TargetCmd = &cobra.Command{
	Use:   "target",
	Short: "Manage provider targets",
}

func init() {
	TargetCmd.AddCommand(targetListCmd)
	TargetCmd.AddCommand(targetSetCmd)
	TargetCmd.AddCommand(targetRemoveCmd)
}
