// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package mcp

import (
	"github.com/spf13/cobra"
)

var MCPCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Manage Daytona MCP Server",
	Long:  "Commands for managing Daytona MCP Server",
}

func init() {
	MCPCmd.AddCommand(InitCmd)
	MCPCmd.AddCommand(StartCmd)
	MCPCmd.AddCommand(ConfigCmd)
}
