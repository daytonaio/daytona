// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_plugin

import (
	"github.com/spf13/cobra"
)

var pluginNameArg string

var PluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Manage plugins",
}

func init() {
	PluginCmd.AddCommand(pluginListCmd)
	PluginCmd.AddCommand(pluginUninstallCmd)
	PluginCmd.AddCommand(pluginInstallCmd)
}
