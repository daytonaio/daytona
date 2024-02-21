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

var pluginListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List installed plugins",
	Args:    cobra.NoArgs,
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		plugins, err := getPluginList()
		if err != nil {
			log.Fatal(err)
		}

		list_view.ListPlugins(plugins)
	},
}

func getPluginList() ([]list_view.PluginViewDTO, error) {
	apiClient, err := api.GetServerApiClient(nil)
	if err != nil {
		log.Fatal(err)
	}

	plugins := []list_view.PluginViewDTO{}
	ctx := context.Background()

	provisionerPluginList, _, err := apiClient.PluginAPI.ListProvisionerPlugins(ctx).Execute()
	if err != nil {
		log.Fatal(err)
	}

	for _, plugin := range provisionerPluginList {
		plugins = append(plugins, list_view.PluginViewDTO{
			Name:    *plugin.Name,
			Version: *plugin.Version,
			Type:    list_view.PluginTypeProvisioner,
		})
	}

	agentServicePluginList, _, err := apiClient.PluginAPI.ListAgentServicePlugins(ctx).Execute()
	if err != nil {
		log.Fatal(err)
	}

	for _, plugin := range agentServicePluginList {
		plugins = append(plugins, list_view.PluginViewDTO{
			Name:    *plugin.Name,
			Version: *plugin.Version,
			Type:    list_view.PluginTypeAgentService,
		})
	}

	return plugins, nil
}
