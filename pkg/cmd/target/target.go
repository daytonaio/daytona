// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"github.com/daytonaio/daytona/internal/util"
	"github.com/spf13/cobra"
)

var TargetCmd = &cobra.Command{
	Use:     "target",
	Aliases: []string{"targets", "tg"},
	Args:    cobra.NoArgs,
	Short:   "Manage targets",
	GroupID: util.TARGET_GROUP,
}

func init() {
	TargetCmd.AddCommand(targetCreateCmd)
	TargetCmd.AddCommand(deleteCmd)
	TargetCmd.AddCommand(infoCmd)
	TargetCmd.AddCommand(restartCmd)
	TargetCmd.AddCommand(startCmd)
	TargetCmd.AddCommand(stopCmd)
	TargetCmd.AddCommand(logsCmd)
	TargetCmd.AddCommand(listCmd)
	TargetCmd.AddCommand(setDefaultCmd)
	TargetCmd.AddCommand(sshCmd)
}
