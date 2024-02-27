// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_plugin

import (
	"context"

	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/server"
	"github.com/daytonaio/daytona/pkg/views/plugins"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

var pluginListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List installed plugins",
	Args:    cobra.NoArgs,
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		pluginList, err := getPluginList()
		if err != nil {
			log.Fatal(err)
		}

		plugins.ListPlugins(pluginList)
	},
}

func getPluginList() ([]plugins.PluginViewDTO, error) {
	apiClient, err := server.GetApiClient(nil)
	if err != nil {
		log.Fatal(err)
	}

	pluginList := []plugins.PluginViewDTO{}
	ctx := context.Background()

	providerPluginList, res, err := apiClient.PluginAPI.ListProviderPlugins(ctx).Execute()
	if err != nil {
		log.Fatal(apiclient.HandleErrorResponse(res, err))
	}

	for _, plugin := range providerPluginList {
		pluginList = append(pluginList, plugins.PluginViewDTO{
			Name:    *plugin.Name,
			Version: *plugin.Version,
			Type:    plugins.PluginTypeProvider,
		})
	}

	agentServicePluginList, res, err := apiClient.PluginAPI.ListAgentServicePlugins(ctx).Execute()
	if err != nil {
		log.Fatal(apiclient.HandleErrorResponse(res, err))
	}

	for _, plugin := range agentServicePluginList {
		pluginList = append(pluginList, plugins.PluginViewDTO{
			Name:    *plugin.Name,
			Version: *plugin.Version,
			Type:    plugins.PluginTypeAgentService,
		})
	}

	return pluginList, nil
}
