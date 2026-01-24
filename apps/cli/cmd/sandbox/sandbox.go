// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"github.com/daytonaio/daytona/cli/internal"
	"github.com/spf13/cobra"
)

var SandboxCmd = &cobra.Command{
	Use:     "sandbox",
	Short:   "Manage Daytona sandboxes",
	Long:    "Commands for managing Daytona sandboxes",
	Aliases: []string{"sandboxes"},
	GroupID: internal.SANDBOX_GROUP,
	Hidden:  true, // Deprecated: use top-level commands instead (e.g., "daytona start" instead of "daytona sandbox start")
}

func init() {
	SandboxCmd.AddCommand(ListCmd)
	SandboxCmd.AddCommand(CreateCmd)
	SandboxCmd.AddCommand(InfoCmd)
	SandboxCmd.AddCommand(DeleteCmd)
	SandboxCmd.AddCommand(StartCmd)
	SandboxCmd.AddCommand(StopCmd)
	SandboxCmd.AddCommand(ArchiveCmd)
	SandboxCmd.AddCommand(SSHCmd)
	SandboxCmd.AddCommand(ExecCmd)
}
