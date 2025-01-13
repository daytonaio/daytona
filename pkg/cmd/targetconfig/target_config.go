// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targetconfig

import (
	"github.com/daytonaio/daytona/internal/util"
	"github.com/spf13/cobra"
)

var TargetConfigCmd = &cobra.Command{
	Use:     "target-config",
	Short:   "Manage target configs",
	Args:    cobra.NoArgs,
	GroupID: util.SERVER_GROUP,
	Aliases: []string{"target-configs", "tc"},
}

func init() {
	TargetConfigCmd.AddCommand(listCmd)
	TargetConfigCmd.AddCommand(TargetConfigCreateCmd)
	TargetConfigCmd.AddCommand(deleteCmd)
}
