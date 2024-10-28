// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaceconfig

import (
	"github.com/daytonaio/daytona/internal/util"
	"github.com/spf13/cobra"
)

var WorkspaceConfigCmd = &cobra.Command{
	Use:     "workspace-config",
	Short:   "Manage workspace configs",
	Aliases: []string{"wc"},
	GroupID: util.TARGET_GROUP,
}

func init() {
	WorkspaceConfigCmd.AddCommand(workspaceConfigListCmd)
	WorkspaceConfigCmd.AddCommand(workspaceConfigInfoCmd)
	WorkspaceConfigCmd.AddCommand(workspaceAddCmd)
	WorkspaceConfigCmd.AddCommand(workspaceConfigUpdateCmd)
	WorkspaceConfigCmd.AddCommand(workspaceConfigSetDefaultCmd)
	WorkspaceConfigCmd.AddCommand(workspaceConfigDeleteCmd)
}
