// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"github.com/daytonaio/daytona/internal/util"
	"github.com/spf13/cobra"
)

var BuildCmd = &cobra.Command{
	Use:     "build",
	Short:   "Manage builds",
	Args:    cobra.NoArgs,
	GroupID: util.TARGET_GROUP,
	Aliases: []string{"builds"},
}

func init() {
	BuildCmd.AddCommand(buildListCmd)
	BuildCmd.AddCommand(buildInfoCmd)
	BuildCmd.AddCommand(buildRunCmd)
	BuildCmd.AddCommand(buildDeleteCmd)
	BuildCmd.AddCommand(buildLogsCmd)
}
