// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"github.com/daytonaio/daytona/internal/util"
	"github.com/spf13/cobra"
)

var BuildCmd = &cobra.Command{
	Use:     "build",
	Aliases: []string{"builds"},
	Short:   "Manage builds",
	GroupID: util.WORKSPACE_GROUP,
}

func init() {
	BuildCmd.AddCommand(buildListCmd)
	BuildCmd.AddCommand(buildInfoCmd)
	BuildCmd.AddCommand(buildRunCmd)
	BuildCmd.AddCommand(buildDeleteCmd)
	BuildCmd.AddCommand(buildLogsCmd)
}
