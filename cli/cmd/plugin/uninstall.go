// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_plugin

import (
	"context"
	"log"

	"github.com/daytonaio/daytona/cli/api"
	"github.com/daytonaio/daytona/cli/cmd/views/plugins/list_view"
	"github.com/spf13/cobra"
)

var pluginUninstallCmd = &cobra.Command{
	Use:     "uninstall",
	Short:   "Uninstall plugin",
	Args:    cobra.NoArgs,
	Aliases: []string{"u"},
	Run: func(cmd *cobra.Command, args []string) {
		plugins, err := getPluginList()
		if err != nil {
			log.Fatal(err)
		}

		pluginToUninstall := list_view.GetPluginFromPrompt(plugins, "Choose a Plugin to uninstall")

		if pluginToUninstall == nil {
			return
		}

		apiClient, err := api.GetServerApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}
		ctx := context.Background()

		if pluginToUninstall.Type == list_view.PluginTypeProvisioner {
			_, err = apiClient.PluginAPI.UninstallProvisionerPlugin(ctx, pluginToUninstall.Name).Execute()
		} else if pluginToUninstall.Type == list_view.PluginTypeAgentService {
			_, err = apiClient.PluginAPI.UninstallAgentServicePlugin(ctx, pluginToUninstall.Name).Execute()
		} else {
			log.Fatal("Unknown plugin type")
		}

		if err != nil {
			log.Fatal(err)
		}
	},
}
