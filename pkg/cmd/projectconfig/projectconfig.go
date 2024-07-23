// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package projectconfig

import (
	"github.com/spf13/cobra"
)

var ProjectConfigCmd = &cobra.Command{
	Use:     "project-config",
	Short:   "Manage project configs",
	Aliases: []string{"pc"},
}

func init() {
	ProjectConfigCmd.AddCommand(projectConfigListCmd)
	ProjectConfigCmd.AddCommand(projectConfigInfoCmd)
	ProjectConfigCmd.AddCommand(projectConfigAddCmd)
	ProjectConfigCmd.AddCommand(projectConfigUpdateCmd)
	ProjectConfigCmd.AddCommand(projectConfigDeleteCmd)
}
